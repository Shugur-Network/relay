package errors

import (
	"fmt"
	"net"
	"syscall"
	
	"github.com/gorilla/websocket"
)

// Relay-specific error constructors

// WebSocketError creates an error for WebSocket-related issues
func WebSocketError(operation string, cause error) *AppError {
	// Determine specific WebSocket error type
	var code string
	var severity ErrorSeverity = SeverityMedium
	var userMessage string
	
	if websocket.IsCloseError(cause, websocket.CloseNormalClosure) {
		code = "WS_NORMAL_CLOSURE"
		severity = SeverityLow
		userMessage = "Connection closed normally."
	} else if websocket.IsCloseError(cause, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		code = "WS_ABNORMAL_CLOSURE"
		severity = SeverityMedium
		userMessage = "Connection lost unexpectedly."
	} else if websocket.IsUnexpectedCloseError(cause, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		code = "WS_UNEXPECTED_CLOSURE"
		severity = SeverityMedium
		userMessage = "Connection closed unexpectedly."
	} else {
		code = "WS_ERROR"
		severity = SeverityMedium
		userMessage = "WebSocket connection error occurred."
	}
	
	return Wrap(cause, ErrorTypeNetwork, code, fmt.Sprintf("WebSocket %s failed", operation)).
		WithSeverity(severity).
		WithUserMessage(userMessage)
}

// EventValidationError creates an error for event validation failures
func EventValidationError(eventID, reason string) *AppError {
	return New(ErrorTypeValidation, "EVENT_VALIDATION_FAILED", fmt.Sprintf("Event validation failed: %s", reason)).
		WithSeverity(SeverityLow).
		WithDetails(fmt.Sprintf("Event ID: %s", eventID)).
		WithUserMessage("The submitted event is invalid. Please check the event format and try again.")
}

// SubscriptionError creates an error for subscription-related issues
func SubscriptionError(subID, reason string) *AppError {
	return New(ErrorTypeValidation, "SUBSCRIPTION_ERROR", fmt.Sprintf("Subscription error: %s", reason)).
		WithSeverity(SeverityLow).
		WithDetails(fmt.Sprintf("Subscription ID: %s", subID)).
		WithUserMessage("The subscription request is invalid. Please check your filter parameters.")
}

// FilterError creates an error for filter validation issues
func FilterError(reason string) *AppError {
	return New(ErrorTypeValidation, "FILTER_ERROR", fmt.Sprintf("Filter validation failed: %s", reason)).
		WithSeverity(SeverityLow).
		WithUserMessage("The filter parameters are invalid. Please check your request.")
}

// ConnectionLimitError creates an error when connection limits are exceeded
func ConnectionLimitError(currentCount, maxCount int) *AppError {
	return New(ErrorTypeRateLimit, "CONNECTION_LIMIT_EXCEEDED", 
		fmt.Sprintf("Connection limit exceeded: %d/%d", currentCount, maxCount)).
		WithSeverity(SeverityMedium).
		WithUserMessage("Too many active connections. Please try again later.")
}

// ClientBannedError creates an error for banned clients
func ClientBannedError(reason string, duration string) *AppError {
	return New(ErrorTypeAuthorization, "CLIENT_BANNED", fmt.Sprintf("Client banned: %s", reason)).
		WithSeverity(SeverityMedium).
		WithDetails(fmt.Sprintf("Ban duration: %s", duration)).
		WithUserMessage("Your client has been temporarily banned due to policy violations.")
}

// NostrProtocolError creates an error for Nostr protocol violations
func NostrProtocolError(command, reason string) *AppError {
	return New(ErrorTypeValidation, "PROTOCOL_ERROR", fmt.Sprintf("Nostr protocol error in %s: %s", command, reason)).
		WithSeverity(SeverityLow).
		WithUserMessage("The request doesn't comply with the Nostr protocol. Please check your client implementation.")
}

// DatabaseConnectionError creates an error for database connection issues
func DatabaseConnectionError(cause error) *AppError {
	return Wrap(cause, ErrorTypeDatabase, "DB_CONNECTION_ERROR", "Database connection failed").
		WithSeverity(SeverityCritical).
		WithUserMessage("Database is temporarily unavailable. Please try again later.")
}

// QueryTimeoutError creates an error for database query timeouts
func QueryTimeoutError(query string, timeoutDuration string) *AppError {
	return New(ErrorTypeTimeout, "QUERY_TIMEOUT", fmt.Sprintf("Database query timed out after %s", timeoutDuration)).
		WithSeverity(SeverityMedium).
		WithDetails(fmt.Sprintf("Query: %s", query)).
		WithUserMessage("The database query took too long. Please try again.")
}

// HealthCheckError creates an error for health check failures
func HealthCheckError(component, reason string) *AppError {
	return New(ErrorTypeInternal, "HEALTH_CHECK_FAILED", fmt.Sprintf("Health check failed for %s: %s", component, reason)).
		WithSeverity(SeverityHigh).
		WithUserMessage("System health check failed. Service may be degraded.")
}

// ConfigurationError creates an error for configuration issues
func ConfigurationError(field, reason string) *AppError {
	return New(ErrorTypeInternal, "CONFIGURATION_ERROR", fmt.Sprintf("Configuration error in %s: %s", field, reason)).
		WithSeverity(SeverityCritical).
		WithUserMessage("Service is misconfigured. Please contact system administrator.")
}

// StorageError creates an error for storage/file system issues
func StorageError(operation, path string, cause error) *AppError {
	return Wrap(cause, ErrorTypeInternal, "STORAGE_ERROR", fmt.Sprintf("Storage %s failed for %s", operation, path)).
		WithSeverity(SeverityHigh).
		WithUserMessage("File system error occurred. Please try again.")
}

// NetworkError creates an error for network-related issues
func NetworkError(operation string, cause error) *AppError {
	var code string
	var severity ErrorSeverity = SeverityMedium
	var userMessage = "Network error occurred. Please check your connection."
	
	// Classify network errors
	if netErr, ok := cause.(net.Error); ok {
		if netErr.Timeout() {
			code = "NETWORK_TIMEOUT"
			userMessage = "Network operation timed out. Please try again."
		} else if netErr.Temporary() {
			code = "NETWORK_TEMPORARY"
			severity = SeverityLow
			userMessage = "Temporary network error. Please try again."
		} else {
			code = "NETWORK_ERROR"
			severity = SeverityHigh
		}
	} else if opErr, ok := cause.(*net.OpError); ok {
		if opErr.Op == "dial" {
			code = "NETWORK_DIAL_FAILED"
			severity = SeverityHigh
			userMessage = "Failed to establish network connection."
		} else if opErr.Op == "read" {
			code = "NETWORK_READ_FAILED"
			userMessage = "Failed to read from network connection."
		} else if opErr.Op == "write" {
			code = "NETWORK_WRITE_FAILED"
			userMessage = "Failed to write to network connection."
		} else {
			code = "NETWORK_OP_FAILED"
		}
	} else if sysErr, ok := cause.(*syscall.Errno); ok {
		switch *sysErr {
		case syscall.ECONNREFUSED:
			code = "CONNECTION_REFUSED"
			severity = SeverityHigh
			userMessage = "Connection refused by remote server."
		case syscall.ECONNRESET:
			code = "CONNECTION_RESET"
			userMessage = "Connection was reset by remote server."
		case syscall.ETIMEDOUT:
			code = "CONNECTION_TIMEOUT"
			userMessage = "Connection timed out."
		default:
			code = "SYSTEM_ERROR"
		}
	} else {
		code = "NETWORK_UNKNOWN"
	}
	
	return Wrap(cause, ErrorTypeNetwork, code, fmt.Sprintf("Network %s failed", operation)).
		WithSeverity(severity).
		WithUserMessage(userMessage)
}

// MetricsError creates an error for metrics collection issues
func MetricsError(operation string, cause error) *AppError {
	return Wrap(cause, ErrorTypeInternal, "METRICS_ERROR", fmt.Sprintf("Metrics %s failed", operation)).
		WithSeverity(SeverityLow). // Metrics errors are typically low severity
		WithUserMessage("Metrics collection error occurred.")
}

// AuthenticationError creates an authentication error
func AuthenticationError(reason string) *AppError {
	return New(ErrorTypeAuthentication, "AUTH_FAILED", fmt.Sprintf("Authentication failed: %s", reason)).
		WithSeverity(SeverityMedium).
		WithUserMessage("Authentication failed. Please provide valid credentials.")
}

// AuthorizationError creates an authorization error
func AuthorizationError(operation, reason string) *AppError {
	return New(ErrorTypeAuthorization, "ACCESS_DENIED", fmt.Sprintf("Access denied for %s: %s", operation, reason)).
		WithSeverity(SeverityMedium).
		WithUserMessage("You don't have permission to perform this action.")
}

// ExternalServiceError creates an error for external service failures
func ExternalServiceError(service, operation string, cause error) *AppError {
	return Wrap(cause, ErrorTypeExternal, "EXTERNAL_SERVICE_ERROR", 
		fmt.Sprintf("External service %s failed during %s", service, operation)).
		WithSeverity(SeverityMedium).
		WithUserMessage("An external service is temporarily unavailable. Please try again later.")
}

// IsRecoverable determines if an error is recoverable (can be retried)
func IsRecoverable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrorTypeTimeout, ErrorTypeNetwork, ErrorTypeDatabase:
			// Network and database errors are often recoverable
			return appErr.Severity != SeverityCritical
		case ErrorTypeRateLimit:
			// Rate limit errors are recoverable after waiting
			return true
		case ErrorTypeExternal:
			// External service errors may be recoverable
			return true
		case ErrorTypeValidation, ErrorTypeAuthentication, ErrorTypeAuthorization, ErrorTypeNotFound:
			// These are typically not recoverable without changing the request
			return false
		case ErrorTypeInternal:
			// Internal errors may or may not be recoverable
			return appErr.Severity == SeverityLow || appErr.Severity == SeverityMedium
		}
	}
	return false
}

// ShouldRetry determines if an operation should be retried based on the error
func ShouldRetry(err error, attemptCount int, maxAttempts int) bool {
	if attemptCount >= maxAttempts {
		return false
	}
	
	return IsRecoverable(err)
}