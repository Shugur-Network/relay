package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-23: Long-form Content
// https://github.com/nostr-protocol/nips/blob/master/23.md

// ValidateLongFormContent validates NIP-23 long-form content events (kind 30023)
func ValidateLongFormContent(evt *nostr.Event) error {
	if evt.Kind != 30023 {
		return fmt.Errorf("invalid event kind for long-form content: %d", evt.Kind)
	}

	// Must have "d" tag for parameterized replaceable events
	hasDTag := false
	hasTitle := false

	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "d":
				hasDTag = true
			case "title":
				hasTitle = true
			}
		}
	}

	if !hasDTag {
		return fmt.Errorf("long-form content must have 'd' tag")
	}

	// Should have title tag
	if !hasTitle {
		return fmt.Errorf("long-form content should have 'title' tag")
	}

	// Content should contain the article content
	if evt.Content == "" {
		return fmt.Errorf("long-form content must have content")
	}

	return nil
}

// IsLongFormContent checks if an event is long-form content
func IsLongFormContent(evt *nostr.Event) bool {
	return evt.Kind == 30023
}
