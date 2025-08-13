package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-33: Parameterized Replaceable Events
// https://github.com/nostr-protocol/nips/blob/master/33.md

// ValidateParameterizedReplaceableEvent validates NIP-33 parameterized replaceable events
func ValidateParameterizedReplaceableEvent(evt *nostr.Event) error {
	// Check if this is a parameterized replaceable event kind
	if !IsParameterizedReplaceableKind(evt.Kind) {
		return fmt.Errorf("invalid event kind for parameterized replaceable event: %d", evt.Kind)
	}

	// Must have "d" tag for identification
	hasDTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			// The "d" tag value can be empty or any string
			break
		}
	}

	if !hasDTag {
		return fmt.Errorf("parameterized replaceable event must have 'd' tag")
	}

	return nil
}

// IsParameterizedReplaceableKind checks if a kind is parameterized replaceable
func IsParameterizedReplaceableKind(kind int) bool {
	// Parameterized replaceable events are in ranges:
	// 30000-39999: Parameterized replaceable events
	return (kind >= 30000 && kind <= 39999)
}

// IsParameterizedReplaceableEvent checks if an event is parameterized replaceable
func IsParameterizedReplaceableEvent(evt *nostr.Event) bool {
	return IsParameterizedReplaceableKind(evt.Kind)
}

// GetDTagValue returns the "d" tag value from a parameterized replaceable event
func GetDTagValue(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			return tag[1]
		}
	}
	return ""
}

// ValidateSpecificParameterizedEvent validates specific parameterized replaceable event kinds
func ValidateSpecificParameterizedEvent(evt *nostr.Event) error {
	switch evt.Kind {
	case 30000, 30001, 30002, 30003:
		return validateGenericParameterizedEvent(evt)
	default:
		// For other kinds, just validate the basic requirement
		return ValidateParameterizedReplaceableEvent(evt)
	}
}

// validateGenericParameterizedEvent validates generic parameterized events (30000-30003)
func validateGenericParameterizedEvent(evt *nostr.Event) error {
	if evt.Kind < 30000 || evt.Kind > 30003 {
		return fmt.Errorf("invalid event kind for generic parameterized event: %d", evt.Kind)
	}

	// Must have "d" tag
	if err := ValidateParameterizedReplaceableEvent(evt); err != nil {
		return err
	}

	return nil
}
