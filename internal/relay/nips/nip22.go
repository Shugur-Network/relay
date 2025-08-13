package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-22: Comment
// https://github.com/nostr-protocol/nips/blob/master/22.md

// ValidateComment validates NIP-22 comment events (kind 1111)
func ValidateComment(evt *nostr.Event) error {
	if evt.Kind != 1111 {
		return fmt.Errorf("invalid event kind for comment: %d", evt.Kind)
	}

	// Must have at least one "e" tag referencing the commented event
	hasEventTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasEventTag = true
			// Validate event ID format
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid event ID in 'e' tag: %s", tag[1])
			}
			break
		}
	}

	if !hasEventTag {
		return fmt.Errorf("comment must reference at least one event with 'e' tag")
	}

	// Content should contain the comment text
	if evt.Content == "" {
		return fmt.Errorf("comment must have content")
	}

	return nil
}

// IsComment checks if an event is a comment
func IsComment(evt *nostr.Event) bool {
	return evt.Kind == 1111
}
