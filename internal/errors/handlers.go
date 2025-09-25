package errors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// Define a custom type for context keys to avoid collisions
type contextKey string

const requestIDKey contextKey = "request_id"

// HandlerFunc is a function type that can return an error
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handler wraps HandlerFunc with automatic error handling
type Handler struct {
	errorMiddleware *ErrorMiddleware
	handlerFunc     HandlerFunc
	logger          *zap.Logger
}

// NewHandler creates a new error-aware handler
func NewHandler(handlerFunc HandlerFunc) *Handler {
	return &Handler{
		errorMiddleware: NewErrorMiddleware(),
		handlerFunc:     handlerFunc,
		logger:          logger.New("error_handler"),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add request ID to context for error tracking
	requestID := generateRequestID()
	ctx := context.WithValue(r.Context(), requestIDKey, requestID)
	r = r.WithContext(ctx)
	
	// Add request ID header to response
	w.Header().Set("X-Request-ID", requestID)
	
	// Call the handler function and handle any errors
	if err := h.handlerFunc(w, r); err != nil {
		h.errorMiddleware.HandleError(w, r, err)
		return
	}
}

// WrapHandler wraps a standard http.HandlerFunc with error handling
func WrapHandler(handlerFunc func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return NewHandler(handlerFunc)
}

// WebSocketHandler is a specialized handler for WebSocket connections
type WebSocketHandler struct {
	errorMiddleware *ErrorMiddleware
	logger          *zap.Logger
}

// NewWebSocketHandler creates a new WebSocket error handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		errorMiddleware: NewErrorMiddleware(),
		logger:          logger.New("websocket_error_handler"),
	}
}

// HandleWebSocketError handles WebSocket-specific errors
func (wh *WebSocketHandler) HandleWebSocketError(conn interface{}, operation string, err error) {
	if err == nil {
		return
	}
	
	// Convert to WebSocket error
	wsErr := WebSocketError(operation, err)
	
	// Log the WebSocket error
	wh.logger.Error("WebSocket error occurred",
		zap.String("operation", operation),
		zap.String("error_type", string(wsErr.Type)),
		zap.String("error_code", wsErr.Code),
		zap.String("severity", string(wsErr.Severity)),
		zap.Error(err))
	
	// Note: For WebSocket connections, we can't send HTTP error responses
	// The error handling here is primarily for logging and metrics
}

// DatabaseHandler provides error handling specifically for database operations
type DatabaseHandler struct {
	logger *zap.Logger
}

// NewDatabaseHandler creates a new database error handler
func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{
		logger: logger.New("database_error_handler"),
	}
}

// HandleDatabaseError processes database errors and returns appropriately wrapped errors
func (dh *DatabaseHandler) HandleDatabaseError(operation string, err error) error {
	if err == nil {
		return nil
	}
	
	// Classify database errors
	var appErr *AppError
	
	// Check for connection errors
	if isConnectionError(err) {
		appErr = DatabaseConnectionError(err)
	} else if isTimeoutError(err) {
		appErr = QueryTimeoutError(operation, "30s") // Default timeout
	} else {
		// Generic database error
		appErr = DatabaseError(operation, err)
	}
	
	// Log the database error
	dh.logger.Error("Database operation failed",
		zap.String("operation", operation),
		zap.String("error_type", string(appErr.Type)),
		zap.String("error_code", appErr.Code),
		zap.String("severity", string(appErr.Severity)),
		zap.Error(err))
	
	return appErr
}

// RelayHandler provides error handling for relay-specific operations
type RelayHandler struct {
	logger *zap.Logger
}

// NewRelayHandler creates a new relay error handler
func NewRelayHandler() *RelayHandler {
	return &RelayHandler{
		logger: logger.New("relay_error_handler"),
	}
}

// HandleEventError processes event-related errors
func (rh *RelayHandler) HandleEventError(eventID, operation string, err error) error {
	if err == nil {
		return nil
	}
	
	var appErr *AppError
	
	// Classify event errors based on the original error
	switch {
	case isValidationError(err):
		appErr = EventValidationError(eventID, err.Error())
	case isRateLimitError(err):
		appErr = RateLimitError("event processing")
	case isDatabaseError(err):
		appErr = DatabaseError(operation, err)
	default:
		appErr = InternalError(fmt.Sprintf("Event %s failed", operation), err)
	}
	
	// Log the event error
	rh.logger.Error("Event operation failed",
		zap.String("event_id", eventID),
		zap.String("operation", operation),
		zap.String("error_type", string(appErr.Type)),
		zap.String("error_code", appErr.Code),
		zap.Error(err))
	
	return appErr
}

// HandleSubscriptionError processes subscription-related errors
func (rh *RelayHandler) HandleSubscriptionError(subID, operation string, err error) error {
	if err == nil {
		return nil
	}
	
	var appErr *AppError
	
	if isValidationError(err) {
		appErr = SubscriptionError(subID, err.Error())
	} else if isRateLimitError(err) {
		appErr = RateLimitError("subscription")
	} else {
		appErr = InternalError(fmt.Sprintf("Subscription %s failed", operation), err)
	}
	
	// Log the subscription error
	rh.logger.Error("Subscription operation failed",
		zap.String("subscription_id", subID),
		zap.String("operation", operation),
		zap.String("error_type", string(appErr.Type)),
		zap.String("error_code", appErr.Code),
		zap.Error(err))
	
	return appErr
}

// Helper functions to classify errors

func isConnectionError(err error) bool {
	// Add logic to detect connection errors
	// This could check for specific error types or error messages
	return false // Placeholder
}

func isTimeoutError(err error) bool {
	// Add logic to detect timeout errors
	return false // Placeholder
}

func isValidationError(err error) bool {
	// Add logic to detect validation errors
	return false // Placeholder
}

func isRateLimitError(err error) bool {
	// Add logic to detect rate limit errors
	return false // Placeholder
}

func isDatabaseError(err error) bool {
	// Add logic to detect database errors
	return false // Placeholder
}

func generateRequestID() string {
	// Simple timestamp-based request ID
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}