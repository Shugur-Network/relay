package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-09: Event Deletion
// https://github.com/nostr-protocol/nips/blob/master/09.md

// ValidateEventDeletion validates NIP-09 event deletion events (kind 5)
func ValidateEventDeletion(evt *nostr.Event) error {
	if evt.Kind != 5 {
		return fmt.Errorf("invalid event kind for event deletion: %d", evt.Kind)
	}

	// Must have at least one "e" tag referencing the event(s) to delete
	hasEventTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasEventTag = true
			// Validate event ID format (should be 64-char hex)
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid event ID in 'e' tag: %s", tag[1])
			}
		}
	}

	if !hasEventTag {
		return fmt.Errorf("deletion event must reference at least one event with 'e' tag")
	}

	return nil
}

// ValidateDeletionAuth returns an error if any "e"‑tagged event in `tags`
// is ALREADY KNOWN (lookup(id) ⇒ author) and its author differs from `deleter`.
func ValidateDeletionAuth(
	tags []nostr.Tag,
	deleter string,
	lookup func(evt string) (event nostr.Event, ok bool),
) error {
	for _, t := range tags {
		if len(t) >= 2 && t[0] == "e" {
			id := t[1]
			if event, ok := lookup(id); ok && event.PubKey != deleter {
				return fmt.Errorf("unauthorized delete of %s", id)
			}
		}
	}
	return nil
}

func IsDeletionEvent(evt nostr.Event) bool {
	return evt.Kind == 5
}
