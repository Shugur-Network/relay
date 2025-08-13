// pkg/relay/nips/nip20.go

package nips

import (
	"encoding/json"
	"fmt"
	"time"

	nostr "github.com/nbd-wtf/go-nostr"
)

// CommandResultType defines the type of command result
type CommandResultType string

const (
	// OK indicates a successful command
	OK CommandResultType = "OK"

	// Error indicates a failed command
	Error CommandResultType = "ERROR"
)

// CommandResult represents the NIP-20 command result
type CommandResult struct {
	Type        CommandResultType `json:"type"`
	EventID     string            `json:"event_id,omitempty"`
	Message     string            `json:"message,omitempty"`
	SuccessFlag bool              `json:"success,omitempty"`
	ErrorCode   string            `json:"error_code,omitempty"`
	Timestamp   int64             `json:"timestamp,omitempty"`
}

// NewOKResult creates a success result for an event
func NewOKResult(eventID string, success bool, message string) *CommandResult {
	return &CommandResult{
		Type:        OK,
		EventID:     eventID,
		SuccessFlag: success,
		Message:     message,
		Timestamp:   time.Now().Unix(),
	}
}

// NewErrorResult creates an error result
func NewErrorResult(eventID string, code string, message string) *CommandResult {
	return &CommandResult{
		Type:      Error,
		EventID:   eventID,
		ErrorCode: code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}
}

// ToJSON converts the command result to a JSON array as per NIP-20
func (cr *CommandResult) ToJSON() ([]byte, error) {
	var result []interface{}

	if cr.Type == OK {
		// ["OK", <event_id>, <success>, <message>]
		result = []interface{}{string(cr.Type), cr.EventID, cr.SuccessFlag}
		if cr.Message != "" {
			result = append(result, cr.Message)
		}
	} else {
		// ["ERROR", <event_id>, <error_code>, <error_message>]
		result = []interface{}{string(cr.Type), cr.EventID, cr.ErrorCode, cr.Message}
	}

	return json.Marshal(result)
}

// Standard error codes as per NIP-20
const (
	ErrorCodeInvalidEvent    = "invalid"      // Event validation failed
	ErrorCodePowerLevels     = "pow"          // Proof of work requirements not met
	ErrorCodeRateLimited     = "rate-limited" // Rate limit exceeded
	ErrorCodeForbidden       = "forbidden"    // Operation not allowed
	ErrorCodeExpired         = "expired"      // Event has expired
	ErrorCodeDuplicate       = "duplicate"    // Event already exists
	ErrorCodeBlacklisted     = "blocked"      // Event/author is blocked
	ErrorCodeRestricted      = "restricted"   // Operation restricted
	ErrorCodeDatabaseError   = "error"        // Database operation failed
	ErrorCodeInvalidFilter   = "unsupported"  // Filter not supported
	ErrorCodeSubscriptionEnd = "shutdown"     // Subscription ended
)

// Standard error messages
const (
	MsgInvalidEvent         = "event validation failed"
	MsgPowerLevels          = "proof of work requirements not met"
	MsgRateLimited          = "too many events"
	MsgForbidden            = "operation not allowed"
	MsgExpired              = "event has expired"
	MsgDuplicate            = "event already exists"
	MsgBlacklisted          = "event/author is blocked"
	MsgRestricted           = "operation restricted"
	MsgDatabaseError        = "database operation failed"
	MsgInvalidFilter        = "filter not supported"
	MsgSubscriptionEnd      = "subscription ended"
	MsgInvalidPubkey        = "invalid pubkey"
	MsgMissingRecipient     = "must have at least one recipient"
	MsgMissingContent       = "content is required"
	MsgInvalidBase64Content = "must be base64 encoded"
)

// Pre-formatted error messages
var (
	ErrInvalidEvent    = FormatErrorMessage(ErrorCodeInvalidEvent, MsgInvalidEvent)
	ErrPowerLevels     = FormatErrorMessage(ErrorCodePowerLevels, MsgPowerLevels)
	ErrRateLimited     = FormatErrorMessage(ErrorCodeRateLimited, MsgRateLimited)
	ErrForbidden       = FormatErrorMessage(ErrorCodeForbidden, MsgForbidden)
	ErrExpired         = FormatErrorMessage(ErrorCodeExpired, MsgExpired)
	ErrDuplicate       = FormatErrorMessage(ErrorCodeDuplicate, MsgDuplicate)
	ErrBlacklisted     = FormatErrorMessage(ErrorCodeBlacklisted, MsgBlacklisted)
	ErrRestricted      = FormatErrorMessage(ErrorCodeRestricted, MsgRestricted)
	ErrDatabaseError   = FormatErrorMessage(ErrorCodeDatabaseError, MsgDatabaseError)
	ErrInvalidFilter   = FormatErrorMessage(ErrorCodeInvalidFilter, MsgInvalidFilter)
	ErrSubscriptionEnd = FormatErrorMessage(ErrorCodeSubscriptionEnd, MsgSubscriptionEnd)
)

// IsStandardErrorCode checks if the error code is a standard NIP-20 error code
func IsStandardErrorCode(code string) bool {
	switch code {
	case ErrorCodeInvalidEvent,
		ErrorCodePowerLevels,
		ErrorCodeRateLimited,
		ErrorCodeForbidden,
		ErrorCodeExpired,
		ErrorCodeDuplicate,
		ErrorCodeBlacklisted,
		ErrorCodeRestricted,
		ErrorCodeDatabaseError,
		ErrorCodeInvalidFilter,
		ErrorCodeSubscriptionEnd:
		return true
	default:
		return false
	}
}

// FormatErrorMessage formats an error message according to NIP-20
func FormatErrorMessage(code string, message string) string {
	if IsStandardErrorCode(code) {
		return code + ": " + message
	}
	return "error: " + message
}

// ValidateCommandResult validates NIP-20 command result events (kind 24133)
func ValidateCommandResult(evt *nostr.Event) error {
	if evt.Kind != 24133 {
		return fmt.Errorf("invalid event kind for command result: %d", evt.Kind)
	}

	// Must have at least one "p" tag (recipient)
	hasRecipient := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			// Validate pubkey format
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid pubkey in 'p' tag")
			}
			hasRecipient = true
		}
	}

	if !hasRecipient {
		return fmt.Errorf("command result must have at least one recipient 'p' tag")
	}

	// Must have JSON content
	if evt.Content == "" {
		return fmt.Errorf("command result must have content")
	}

	// Parse and validate JSON content
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(evt.Content), &result); err != nil {
		return fmt.Errorf("command result content must be valid JSON")
	}

	// Check for required fields
	if _, hasResult := result["result"]; !hasResult {
		return fmt.Errorf("command result must have 'result' field")
	}

	if _, hasCommand := result["command"]; !hasCommand {
		return fmt.Errorf("command result must have 'command' field")
	}

	if _, hasMessage := result["message"]; !hasMessage {
		return fmt.Errorf("command result must have 'message' field")
	}

	// Validate result type
	if resultType, ok := result["result"].(string); ok {
		if resultType != "success" && resultType != "error" {
			return fmt.Errorf("command result type must be 'success' or 'error'")
		}
	} else {
		return fmt.Errorf("command result 'result' field must be a string")
	}

	return nil
}

// IsCommandResult checks if an event is a command result
func IsCommandResult(evt *nostr.Event) bool {
	return evt.Kind == 24133
}

// FormatDMError formats a direct message validation error
func FormatDMError(errorType string) string {
	switch errorType {
	case "invalid_pubkey":
		return FormatErrorMessage(ErrorCodeInvalidEvent, MsgInvalidPubkey)
	case "missing_recipient":
		return FormatErrorMessage(ErrorCodeInvalidEvent, MsgMissingRecipient)
	case "missing_content":
		return FormatErrorMessage(ErrorCodeInvalidEvent, MsgMissingContent)
	case "invalid_base64":
		return FormatErrorMessage(ErrorCodeInvalidEvent, MsgInvalidBase64Content)
	default:
		return FormatErrorMessage(ErrorCodeInvalidEvent, MsgInvalidEvent)
	}
}
