package common

import (
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidationHelper combines Validator, TagValidator, and ErrorFormatter for convenience
type ValidationHelper struct {
	*Validator
	*TagValidator
	*ErrorFormatter
}

// NewValidationHelper creates a complete validation helper for a NIP
func NewValidationHelper(nipNumber string, eventKind int, eventName string) *ValidationHelper {
	return &ValidationHelper{
		Validator:      NewValidator(nipNumber, eventKind, eventName),
		TagValidator:   NewTagValidator(),
		ErrorFormatter: NewErrorFormatter(nipNumber, eventName),
	}
}

// Utility functions that were duplicated across multiple NIPs

// IsHexChar checks if a character is a valid hexadecimal character
func IsHexChar(char rune) bool {
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')
}

// IsHexString checks if a string contains only hexadecimal characters
func IsHexString(s string) bool {
	for _, char := range s {
		if !IsHexChar(char) {
			return false
		}
	}
	return true
}

// IsAddressable checks if an event is addressable according to NIP-33
// This was duplicated in nip01.go and other files
func IsAddressable(evt nostr.Event) bool {
	return evt.Kind >= 30000 && evt.Kind < 40000 && GetTagValue(evt, "d") != ""
}

// GetTagValue returns the first t[1] found for the given key, or "" if not found
// This function exists in nip01.go and should be migrated to use the common version
func GetTagValue(evt nostr.Event, key string) string {
	for _, t := range evt.Tags {
		if len(t) >= 2 && t[0] == key {
			return t[1]
		}
	}
	return ""
}

// IsParameterizedReplaceableKind checks if an event kind is parameterized replaceable (30000-39999)
func IsParameterizedReplaceableKind(kind int) bool {
	return kind >= 30000 && kind < 40000
}

// IsEphemeralKind checks if an event kind is ephemeral (20000-29999)
func IsEphemeralKind(kind int) bool {
	return kind >= 20000 && kind < 30000
}

// IsRegularKind checks if an event kind is regular (0-9999)
func IsRegularKind(kind int) bool {
	return kind >= 0 && kind < 10000
}

// IsReplaceableKind checks if an event kind is replaceable (10000-19999)
func IsReplaceableKind(kind int) bool {
	return kind >= 10000 && kind < 20000
}

// Common validation patterns that appear across multiple NIPs

// ValidateBasicEvent performs the most common validations that nearly all NIPs need
func ValidateBasicEvent(event *nostr.Event, nipNumber string, expectedKind int, eventName string) error {
	helper := NewValidationHelper(nipNumber, expectedKind, eventName)
	return helper.ValidateBasics(event)
}

// ValidateEventWithRequiredTags validates basic event structure plus required tags
func ValidateEventWithRequiredTags(event *nostr.Event, nipNumber string, expectedKind int, eventName string, requiredTags ...string) error {
	helper := NewValidationHelper(nipNumber, expectedKind, eventName)

	if err := helper.ValidateBasics(event); err != nil {
		return err
	}

	if err := helper.ValidateRequiredTags(event, requiredTags...); err != nil {
		return helper.ErrorFormatter.FormatError("%v", err)
	}

	return nil
}

// ValidateEventWithCallback performs basic validation and custom logic, then logs success
func ValidateEventWithCallback(event *nostr.Event, nipNumber string, expectedKind int, eventName string, customValidation func(*ValidationHelper, *nostr.Event) error) error {
	helper := NewValidationHelper(nipNumber, expectedKind, eventName)

	if err := helper.ValidateBasics(event); err != nil {
		return err
	}

	if customValidation != nil {
		if err := customValidation(helper, event); err != nil {
			return err // Error should already be formatted by custom validation
		}
	}

	helper.LogSuccess(event)
	return nil
}
