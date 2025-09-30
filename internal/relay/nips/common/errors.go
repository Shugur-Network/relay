package common

import (
	"fmt"
	"strings"
)

// ErrorFormatter provides standardized error formatting for NIPs
type ErrorFormatter struct {
	nipNumber string
	eventName string
}

// NewErrorFormatter creates a new error formatter for a specific NIP
func NewErrorFormatter(nipNumber, eventName string) *ErrorFormatter {
	return &ErrorFormatter{
		nipNumber: nipNumber,
		eventName: eventName,
	}
}

// FormatError creates a standardized error message
func (ef *ErrorFormatter) FormatError(message string, args ...interface{}) error {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return fmt.Errorf("NIP-%s %s validation failed: %s", ef.nipNumber, ef.eventName, message)
}

// FormatTagError creates an error message specifically for tag validation failures
func (ef *ErrorFormatter) FormatTagError(tagName, message string, args ...interface{}) error {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return fmt.Errorf("NIP-%s %s validation failed: invalid '%s' tag: %s",
		ef.nipNumber, ef.eventName, tagName, message)
}

// FormatMissingTagError creates a standardized error for missing required tags
func (ef *ErrorFormatter) FormatMissingTagError(tagName string) error {
	return fmt.Errorf("NIP-%s %s validation failed: missing required '%s' tag",
		ef.nipNumber, ef.eventName, tagName)
}

// FormatInvalidKindError creates a standardized error for wrong event kinds
func (ef *ErrorFormatter) FormatInvalidKindError(expected, actual int) error {
	return fmt.Errorf("NIP-%s %s validation failed: invalid event kind: expected %d, got %d",
		ef.nipNumber, ef.eventName, expected, actual)
}

// Common error messages and codes used across NIPs

// ValidationErrorCode represents standardized error codes
type ValidationErrorCode string

const (
	ErrorCodeInvalidKind      ValidationErrorCode = "invalid_kind"
	ErrorCodeMissingTag       ValidationErrorCode = "missing_tag"
	ErrorCodeInvalidTag       ValidationErrorCode = "invalid_tag"
	ErrorCodeInvalidPubkey    ValidationErrorCode = "invalid_pubkey"
	ErrorCodeInvalidEventID   ValidationErrorCode = "invalid_event_id"
	ErrorCodeInvalidURL       ValidationErrorCode = "invalid_url"
	ErrorCodeInvalidContent   ValidationErrorCode = "invalid_content"
	ErrorCodeMissingRecipient ValidationErrorCode = "missing_recipient"
	ErrorCodeInvalidSignature ValidationErrorCode = "invalid_signature"
	ErrorCodeExpiredEvent     ValidationErrorCode = "expired_event"
	ErrorCodeUnauthorized     ValidationErrorCode = "unauthorized"
)

// StandardError represents a standardized validation error
type StandardError struct {
	Code      ValidationErrorCode `json:"code"`
	Message   string              `json:"message"`
	NIPNumber string              `json:"nip_number"`
	EventName string              `json:"event_name"`
	TagName   string              `json:"tag_name,omitempty"`
	Expected  interface{}         `json:"expected,omitempty"`
	Actual    interface{}         `json:"actual,omitempty"`
}

// Error implements the error interface
func (se StandardError) Error() string {
	parts := []string{
		fmt.Sprintf("NIP-%s %s validation failed", se.NIPNumber, se.EventName),
		fmt.Sprintf("[%s]", se.Code),
		se.Message,
	}

	if se.TagName != "" {
		parts = append(parts, fmt.Sprintf("(tag: %s)", se.TagName))
	}

	return strings.Join(parts, " ")
}

// NewStandardError creates a new standardized error
func NewStandardError(nipNumber, eventName string, code ValidationErrorCode, message string) *StandardError {
	return &StandardError{
		Code:      code,
		Message:   message,
		NIPNumber: nipNumber,
		EventName: eventName,
	}
}

// WithTag adds tag information to the error
func (se *StandardError) WithTag(tagName string) *StandardError {
	se.TagName = tagName
	return se
}

// WithExpected adds expected value information to the error
func (se *StandardError) WithExpected(expected interface{}) *StandardError {
	se.Expected = expected
	return se
}

// WithActual adds actual value information to the error
func (se *StandardError) WithActual(actual interface{}) *StandardError {
	se.Actual = actual
	return se
}

// Common error messages that can be reused across NIPs

var (
	// Generic validation errors
	ErrEventNil               = "event is nil"
	ErrInvalidContentEmpty    = "content must not be empty"
	ErrInvalidContentNonEmpty = "content must be empty"

	// Tag-related errors
	ErrTagMissingValue  = "tag is missing required value"
	ErrTagInvalidFormat = "tag has invalid format"
	ErrTagDuplicate     = "duplicate tag not allowed"
	ErrTagTooMany       = "too many tags of this type"
	ErrTagTooFew        = "insufficient tags of this type"

	// Pubkey/ID validation errors
	ErrInvalidPubkeyFormat    = "invalid pubkey format (must be 64-character hex)"
	ErrInvalidEventIDFormat   = "invalid event ID format (must be 64-character hex)"
	ErrInvalidSignatureFormat = "invalid signature format (must be 128-character hex)"

	// URL validation errors
	ErrInvalidURLFormat = "invalid URL format"
	ErrURLMissingScheme = "URL must include scheme (http/https)"
	ErrURLMissingHost   = "URL must include host"

	// Content validation errors
	ErrContentTooLong       = "content exceeds maximum length"
	ErrContentInvalidJSON   = "content is not valid JSON"
	ErrContentInvalidBase64 = "content is not valid base64"

	// Authorization errors
	ErrUnauthorizedDeletion     = "only event author can delete their events"
	ErrUnauthorizedModification = "only event author can modify their events"

	// Time-related errors
	ErrEventExpired       = "event has expired"
	ErrInvalidTimestamp   = "invalid timestamp format"
	ErrTimestampTooOld    = "event timestamp is too old"
	ErrTimestampTooFuture = "event timestamp is too far in the future"
)

// FormatDMError formats Direct Message related errors (used in NIP-04, NIP-17, NIP-44)
func FormatDMError(errorType string) string {
	switch errorType {
	case "missing_recipient":
		return "Direct message must have recipient 'p' tag"
	case "invalid_pubkey":
		return "Invalid recipient pubkey format"
	case "invalid_content":
		return "Invalid encrypted content format"
	case "missing_encrypted_tag":
		return "Missing 'encrypted' tag for encrypted message"
	default:
		return fmt.Sprintf("Direct message validation error: %s", errorType)
	}
}
