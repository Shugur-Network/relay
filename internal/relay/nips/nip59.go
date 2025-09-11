package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-59: Gift Wrap
// https://github.com/nostr-protocol/nips/blob/master/59.md

// ValidateGiftWrapEvent validates NIP-59 gift wrap events
func ValidateGiftWrapEvent(evt *nostr.Event) error {
	switch evt.Kind {
	case 1059:
		return validateGiftWrapOuter(evt)
	case 13194:
		return validateWalletConnectEvent(evt)
	default:
		return fmt.Errorf("invalid event kind for gift wrap: %d", evt.Kind)
	}
}

// validateGiftWrapOuter validates outer gift wrap events (kind 1059)
func validateGiftWrapOuter(evt *nostr.Event) error {
	if evt.Kind != 1059 {
		return fmt.Errorf("invalid event kind for gift wrap: %d", evt.Kind)
	}

	// Must have "p" tag with recipient pubkey
	hasPTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			hasPTag = true
			// Validate pubkey format
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid pubkey in 'p' tag: %s", tag[1])
			}
			break
		}
	}

	if !hasPTag {
		return fmt.Errorf("gift wrap must have 'p' tag with recipient")
	}

	// Content must be encrypted (non-empty)
	if evt.Content == "" {
		return fmt.Errorf("gift wrap must have encrypted content")
	}

	// Validate NIP-44 format using existing validator
	if err := ValidateNIP44Payload(evt.Content); err != nil {
		return fmt.Errorf("invalid NIP-44 content in gift wrap: %w", err)
	}

	// CreatedAt should be randomized for privacy
	// We can't validate this strictly, but we can check it's reasonable
	if evt.CreatedAt == 0 {
		return fmt.Errorf("gift wrap must have created_at timestamp")
	}

	return nil
}

// validateWalletConnectEvent validates wallet connect events (kind 13194)
func validateWalletConnectEvent(evt *nostr.Event) error {
	if evt.Kind != 13194 {
		return fmt.Errorf("invalid event kind for wallet connect: %d", evt.Kind)
	}

	// Should have appropriate tags for wallet connect
	// This is more flexible validation as the spec may evolve
	return nil
}

// IsGiftWrapEvent checks if an event is a gift wrap event
func IsGiftWrapEvent(evt *nostr.Event) bool {
	return evt.Kind == 13 || evt.Kind == 1059 || evt.Kind == 13194
}

// IsSealEvent checks if an event is a seal event (kind 13)
func IsSealEvent(evt *nostr.Event) bool {
	return evt.Kind == 13
}

// IsOuterGiftWrap checks if an event is an outer gift wrap (kind 1059)
func IsOuterGiftWrap(evt *nostr.Event) bool {
	return evt.Kind == 1059
}

// IsWalletConnectEvent checks if an event is a wallet connect event (kind 13194)
func IsWalletConnectEvent(evt *nostr.Event) bool {
	return evt.Kind == 13194
}
