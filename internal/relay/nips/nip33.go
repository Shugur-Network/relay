package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-33: Addressable Events
// https://github.com/nostr-protocol/nips/blob/master/33.md

// ValidateAddressableEvent validates NIP-33 addressable events
func ValidateAddressableEvent(evt *nostr.Event) error {
	// Check if this is a addressable event kind
	if !IsAddressableKind(evt.Kind) {
		return fmt.Errorf("invalid event kind for addressable event: %d", evt.Kind)
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
		return fmt.Errorf("addressable event must have 'd' tag")
	}

	return nil
}

// IsAddressableKind checks if a kind is addressable
func IsAddressableKind(kind int) bool {
	// Addressable events are in range 30000-39999
	// This includes Time Capsule Replaceable events (kind 31995)
	return kind >= 30000 && kind <= 39999
}

// IsAddressableEvent checks if an event is addressable
func IsAddressableEvent(evt *nostr.Event) bool {
	return IsAddressableKind(evt.Kind)
}

// GetDTagValue returns the "d" tag value from a addressable event
func GetDTagValue(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			return tag[1]
		}
	}
	return ""
}

// ValidateSpecificParameterizedEvent validates specific addressable event kinds
func ValidateSpecificParameterizedEvent(evt *nostr.Event) error {
	switch evt.Kind {
	case 30000, 30001, 30002, 30003:
		return validateGenericParameterizedEvent(evt)
	default:
		// For other kinds, just validate the basic requirement
		return ValidateAddressableEvent(evt)
	}
}

// validateGenericParameterizedEvent validates generic parameterized events (30000-30003)
func validateGenericParameterizedEvent(evt *nostr.Event) error {
	if evt.Kind < 30000 || evt.Kind > 30003 {
		return fmt.Errorf("invalid event kind for generic parameterized event: %d", evt.Kind)
	}

	// Must have "d" tag
	if err := ValidateAddressableEvent(evt); err != nil {
		return err
	}

	return nil
}
