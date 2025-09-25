package errors

import (
	"net/http"

	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

var (
	// Global error middleware instance
	globalErrorMiddleware *ErrorMiddleware
	
	// Global specialized handlers
	globalWebSocketHandler *WebSocketHandler
	globalDatabaseHandler  *DatabaseHandler
	globalRelayHandler     *RelayHandler
)

// InitErrorHandling initializes the global error handling system
func InitErrorHandling() {
	globalErrorMiddleware = NewErrorMiddleware()
	globalWebSocketHandler = NewWebSocketHandler()
	globalDatabaseHandler = NewDatabaseHandler()
	globalRelayHandler = NewRelayHandler()
	
	logger.Info("Error handling system initialized",
		zap.String("component", "error_middleware"))
}

// GetErrorMiddleware returns the global error middleware instance
func GetErrorMiddleware() *ErrorMiddleware {
	if globalErrorMiddleware == nil {
		InitErrorHandling()
	}
	return globalErrorMiddleware
}

// GetWebSocketHandler returns the global WebSocket error handler
func GetWebSocketHandler() *WebSocketHandler {
	if globalWebSocketHandler == nil {
		InitErrorHandling()
	}
	return globalWebSocketHandler
}

// GetDatabaseHandler returns the global database error handler
func GetDatabaseHandler() *DatabaseHandler {
	if globalDatabaseHandler == nil {
		InitErrorHandling()
	}
	return globalDatabaseHandler
}

// GetRelayHandler returns the global relay error handler
func GetRelayHandler() *RelayHandler {
	if globalRelayHandler == nil {
		InitErrorHandling()
	}
	return globalRelayHandler
}

// HandleHTTPError is a convenience function for handling HTTP errors
func HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	GetErrorMiddleware().HandleError(w, r, err)
}

// HandleWebSocketError is a convenience function for handling WebSocket errors
func HandleWebSocketError(conn interface{}, operation string, err error) {
	GetWebSocketHandler().HandleWebSocketError(conn, operation, err)
}

// HandleDatabaseError is a convenience function for handling database errors
func HandleDatabaseError(operation string, err error) error {
	return GetDatabaseHandler().HandleDatabaseError(operation, err)
}

// HandleEventError is a convenience function for handling event errors
func HandleEventError(eventID, operation string, err error) error {
	return GetRelayHandler().HandleEventError(eventID, operation, err)
}

// HandleSubscriptionError is a convenience function for handling subscription errors
func HandleSubscriptionError(subID, operation string, err error) error {
	return GetRelayHandler().HandleSubscriptionError(subID, operation, err)
}

// RecoveryMiddleware returns a middleware that recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return GetErrorMiddleware().RecoveryMiddleware(next)
}