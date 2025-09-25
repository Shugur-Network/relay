package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"go.uber.org/zap"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization" 
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeDatabase       ErrorType = "database"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeExternal       ErrorType = "external"
)

// ErrorSeverity represents the severity level of errors
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"      // Minor issues, application continues normally
	SeverityMedium   ErrorSeverity = "medium"   // Notable issues that may affect user experience
	SeverityHigh     ErrorSeverity = "high"     // Serious issues that significantly impact functionality
	SeverityCritical ErrorSeverity = "critical" // Critical issues that may cause system instability
)

// AppError represents a structured application error
type AppError struct {
	Type        ErrorType     `json:"type"`
	Code        string        `json:"code"`
	Message     string        `json:"message"`
	Details     string        `json:"details,omitempty"`
	Severity    ErrorSeverity `json:"severity"`
	Timestamp   time.Time     `json:"timestamp"`
	RequestID   string        `json:"request_id,omitempty"`
	UserMessage string        `json:"user_message,omitempty"`
	Cause       error         `json:"-"`
	StackTrace  string        `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap implements the Unwrap interface for error wrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ErrorResponse represents the JSON response format for errors
type ErrorResponse struct {
	Error struct {
		Type        ErrorType `json:"type"`
		Code        string    `json:"code"`
		Message     string    `json:"message"`
		Timestamp   time.Time `json:"timestamp"`
		RequestID   string    `json:"request_id,omitempty"`
	} `json:"error"`
}

// ErrorMiddleware handles error processing and response formatting
type ErrorMiddleware struct {
	logger *zap.Logger
}

// NewErrorMiddleware creates a new error middleware instance
func NewErrorMiddleware() *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger.New("error_middleware"),
	}
}

// New creates a new AppError with stack trace capture
func New(errorType ErrorType, code string, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Severity:   SeverityMedium, // Default severity
		Timestamp:  time.Now(),
		StackTrace: captureStackTrace(),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, code string, message string) *AppError {
	appErr := &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Severity:   SeverityMedium,
		Timestamp:  time.Now(),
		Cause:      err,
		StackTrace: captureStackTrace(),
	}
	
	// If the original error has details, include them
	if err != nil {
		appErr.Details = err.Error()
	}
	
	return appErr
}

// WithSeverity sets the severity level of an error
func (e *AppError) WithSeverity(severity ErrorSeverity) *AppError {
	e.Severity = severity
	return e
}

// WithDetails adds additional details to an error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithUserMessage sets a user-friendly message
func (e *AppError) WithUserMessage(message string) *AppError {
	e.UserMessage = message
	return e
}

// WithRequestID associates an error with a request ID
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// HandleError processes an error and sends appropriate HTTP response
func (em *ErrorMiddleware) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError
	
	// Convert to AppError if it isn't already
	if ae, ok := err.(*AppError); ok {
		appErr = ae
	} else {
		// Create a generic internal error for unknown error types
		appErr = Wrap(err, ErrorTypeInternal, "INTERNAL_ERROR", "An internal error occurred")
		appErr.Severity = SeverityHigh
	}
	
	// Add request ID if available
	if requestID := getRequestID(r); requestID != "" {
		appErr.RequestID = requestID
	}
	
	// Log the error with appropriate level based on severity
	em.logError(appErr, r)
	
	// Increment error metrics
	metrics.IncrementErrorCount()
	
	// Send HTTP response
	em.sendErrorResponse(w, appErr)
}

// logError logs an error with appropriate severity level
func (em *ErrorMiddleware) logError(err *AppError, r *http.Request) {
	fields := []zap.Field{
		zap.String("error_type", string(err.Type)),
		zap.String("error_code", err.Code),
		zap.String("severity", string(err.Severity)),
		zap.Time("timestamp", err.Timestamp),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr),
	}
	
	if err.RequestID != "" {
		fields = append(fields, zap.String("request_id", err.RequestID))
	}
	
	if err.Details != "" {
		fields = append(fields, zap.String("details", err.Details))
	}
	
	if err.Cause != nil {
		fields = append(fields, zap.Error(err.Cause))
	}
	
	// Log stack trace for high severity errors
	if err.Severity == SeverityHigh || err.Severity == SeverityCritical {
		fields = append(fields, zap.String("stack_trace", err.StackTrace))
	}
	
	// Choose log level based on severity
	switch err.Severity {
	case SeverityLow:
		em.logger.Info(err.Message, fields...)
	case SeverityMedium:
		em.logger.Warn(err.Message, fields...)
	case SeverityHigh:
		em.logger.Error(err.Message, fields...)
	case SeverityCritical:
		em.logger.Error(err.Message, fields...)
		// For critical errors, also log to separate critical log or alert system
		em.logger.Error("CRITICAL ERROR DETECTED", fields...)
	}
}

// sendErrorResponse sends a structured JSON error response
func (em *ErrorMiddleware) sendErrorResponse(w http.ResponseWriter, err *AppError) {
	statusCode := getHTTPStatusCode(err.Type)
	
	response := ErrorResponse{
		Error: struct {
			Type        ErrorType `json:"type"`
			Code        string    `json:"code"`
			Message     string    `json:"message"`
			Timestamp   time.Time `json:"timestamp"`
			RequestID   string    `json:"request_id,omitempty"`
		}{
			Type:      err.Type,
			Code:      err.Code,
			Message:   getUserFriendlyMessage(err),
			Timestamp: err.Timestamp,
			RequestID: err.RequestID,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		em.logger.Error("Failed to encode error response", zap.Error(encodeErr))
		// Fallback to plain text response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getHTTPStatusCode maps error types to HTTP status codes
func getHTTPStatusCode(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeDatabase, ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeExternal:
		return http.StatusBadGateway
	case ErrorTypeNetwork:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// getUserFriendlyMessage returns a user-friendly error message
func getUserFriendlyMessage(err *AppError) string {
	if err.UserMessage != "" {
		return err.UserMessage
	}
	
	// Provide default user-friendly messages based on error type
	switch err.Type {
	case ErrorTypeValidation:
		return "The request contains invalid data. Please check your input and try again."
	case ErrorTypeAuthentication:
		return "Authentication required. Please provide valid credentials."
	case ErrorTypeAuthorization:
		return "You don't have permission to perform this action."
	case ErrorTypeNotFound:
		return "The requested resource was not found."
	case ErrorTypeRateLimit:
		return "Too many requests. Please wait before trying again."
	case ErrorTypeTimeout:
		return "The request timed out. Please try again."
	case ErrorTypeDatabase:
		return "A database error occurred. Please try again later."
	case ErrorTypeNetwork:
		return "A network error occurred. Please check your connection and try again."
	case ErrorTypeExternal:
		return "An external service error occurred. Please try again later."
	default:
		return "An unexpected error occurred. Please try again."
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// getRequestID extracts request ID from request context or headers
func getRequestID(r *http.Request) string {
	// Try to get from context first
	if requestID := r.Context().Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	
	// Fallback to headers
	return r.Header.Get("X-Request-ID")
}

// RecoveryMiddleware is a middleware that recovers from panics and converts them to structured errors
func (em *ErrorMiddleware) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				var err error
				if e, ok := recovered.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("panic: %v", recovered)
				}
				
				panicErr := Wrap(err, ErrorTypeInternal, "PANIC_RECOVERED", "An unexpected error occurred")
				panicErr.Severity = SeverityCritical
				
				em.HandleError(w, r, panicErr)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// Common error constructors for frequently used error types

// ValidationError creates a validation error
func ValidationError(code, message string) *AppError {
	return New(ErrorTypeValidation, code, message).
		WithSeverity(SeverityLow).
		WithUserMessage("Please check your input and try again.")
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *AppError {
	return New(ErrorTypeNotFound, "NOT_FOUND", fmt.Sprintf("%s not found", resource)).
		WithSeverity(SeverityLow).
		WithUserMessage("The requested resource was not found.")
}

// DatabaseError creates a database error
func DatabaseError(operation string, cause error) *AppError {
	return Wrap(cause, ErrorTypeDatabase, "DATABASE_ERROR", fmt.Sprintf("Database %s failed", operation)).
		WithSeverity(SeverityHigh).
		WithUserMessage("A database error occurred. Please try again later.")
}

// RateLimitError creates a rate limit error
func RateLimitError(resource string) *AppError {
	return New(ErrorTypeRateLimit, "RATE_LIMIT_EXCEEDED", fmt.Sprintf("Rate limit exceeded for %s", resource)).
		WithSeverity(SeverityMedium).
		WithUserMessage("Too many requests. Please wait before trying again.")
}

// TimeoutError creates a timeout error
func TimeoutError(operation string) *AppError {
	return New(ErrorTypeTimeout, "TIMEOUT", fmt.Sprintf("%s operation timed out", operation)).
		WithSeverity(SeverityMedium).
		WithUserMessage("The request timed out. Please try again.")
}

// InternalError creates an internal error
func InternalError(message string, cause error) *AppError {
	return Wrap(cause, ErrorTypeInternal, "INTERNAL_ERROR", message).
		WithSeverity(SeverityHigh).
		WithUserMessage("An internal error occurred. Please try again.")
}