package relay

import (
	"fmt"
)

// Common error types for the relay
var (
	// Validation errors
	ErrInvalidEventID       = fmt.Errorf("invalid event ID format")
	ErrInvalidTimeRange     = fmt.Errorf("'since' timestamp is after 'until' timestamp")
	ErrTimeTooFarInFuture   = fmt.Errorf("'until' timestamp is too far in the future")
	ErrTooManyTagValues     = fmt.Errorf("too many values in tag filter (max 20)")
	ErrInvalidPubkey        = fmt.Errorf("invalid pubkey format")
	ErrInvalidSignature     = fmt.Errorf("invalid signature")
	ErrUnknownKind          = fmt.Errorf("unknown event kind")
	ErrContentTooLarge      = fmt.Errorf("content exceeds maximum allowed size")
	ErrTooManyTags          = fmt.Errorf("too many tags")
	ErrTagsTooLarge         = fmt.Errorf("tags exceed maximum total size")
	ErrInvalidFilter        = fmt.Errorf("invalid filter")
	ErrEventFromFuture      = fmt.Errorf("event timestamp is too far in the future")
	ErrEventTooOld          = fmt.Errorf("event timestamp is too old")
	ErrInvalidJSON          = fmt.Errorf("invalid JSON format")
	ErrInvalidSealContent   = fmt.Errorf("invalid sealed content")
	ErrPubkeyBlacklisted    = fmt.Errorf("pubkey is blacklisted")
	ErrUnauthorizedDeletion = fmt.Errorf("only the event author can delete events")
	ErrMissingRequiredTag   = fmt.Errorf("missing required tag")
	ErrInvalidDelegation    = fmt.Errorf("invalid delegation")
	ErrUnencryptedDM        = fmt.Errorf("direct messages must be encrypted")
	ErrMissingContent       = fmt.Errorf("content is required for this event kind")

	// Rate limit errors
	ErrRateLimitExceeded    = fmt.Errorf("rate limit exceeded")
	ErrTooManyConnections   = fmt.Errorf("too many connections")
	ErrTooManySubscriptions = fmt.Errorf("too many subscriptions")
	ErrClientBanned         = fmt.Errorf("client temporarily banned due to excessive requests")

	// Database errors
	ErrDatabaseRead      = fmt.Errorf("database read error")
	ErrDatabaseWrite     = fmt.Errorf("database write error")
	ErrDuplicateEvent    = fmt.Errorf("event already exists")
	ErrEventNotFound     = fmt.Errorf("event not found")
	ErrTransactionFailed = fmt.Errorf("database transaction failed")

	// Connection errors
	ErrConnectionClosed = fmt.Errorf("connection closed")
	ErrWriteTimeout     = fmt.Errorf("write timeout")
	ErrReadTimeout      = fmt.Errorf("read timeout")
	ErrPingTimeout      = fmt.Errorf("ping timeout")
	ErrIdleTimeout      = fmt.Errorf("connection idle timeout")

	// Subscription errors
	ErrSubscriptionNotFound  = fmt.Errorf("subscription not found")
	ErrInvalidSubscriptionID = fmt.Errorf("invalid subscription ID")
	ErrQueryTimeout          = fmt.Errorf("query timeout")
)

// FormatError returns a formatted error message suitable for clients
func FormatError(err error, code string) string {
	if code == "" {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return fmt.Sprintf("%s: %s", code, err.Error())
}

// IsTemporaryError returns true if the error is temporary and the operation could succeed if retried
func IsTemporaryError(err error) bool {
	switch err {
	case ErrDatabaseRead, ErrDatabaseWrite, ErrTransactionFailed,
		ErrWriteTimeout, ErrReadTimeout, ErrQueryTimeout:
		return true
	default:
		return false
	}
}

// IsPermanentError returns true if the error is permanent and retrying won't help
func IsPermanentError(err error) bool {
	switch err {
	case ErrInvalidEventID, ErrInvalidPubkey, ErrInvalidSignature,
		ErrUnknownKind, ErrContentTooLarge, ErrTooManyTags,
		ErrEventFromFuture, ErrEventTooOld, ErrInvalidJSON,
		ErrPubkeyBlacklisted, ErrUnauthorizedDeletion,
		ErrMissingRequiredTag, ErrInvalidDelegation,
		ErrUnencryptedDM, ErrMissingContent, ErrDuplicateEvent:
		return true
	default:
		return false
	}
}
