package nips

import (
	"fmt"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-09: Event Deletion
// https://github.com/nostr-protocol/nips/blob/master/09.md

// ValidateEventDeletion validates NIP-09 event deletion events (kind 5)
func ValidateEventDeletion(evt *nostr.Event) error {
	logger.Debug("NIP-09: Validating event deletion", 
		zap.String("event_id", evt.ID),
		zap.String("pubkey", evt.PubKey))
		
	if evt.Kind != 5 {
		logger.Warn("NIP-09: Invalid event kind for deletion", 
			zap.String("event_id", evt.ID),
			zap.Int("kind", evt.Kind))
		return fmt.Errorf("invalid event kind for event deletion: %d", evt.Kind)
	}

	// Must have at least one "e" tag referencing the event(s) to delete
	hasEventTag := false
	eventCount := 0
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasEventTag = true
			eventCount++
			// Validate event ID format (should be 64-char hex)
			if len(tag[1]) != 64 {
				logger.Warn("NIP-09: Invalid event ID in 'e' tag", 
					zap.String("deletion_event_id", evt.ID),
					zap.String("invalid_event_id", tag[1]))
				return fmt.Errorf("invalid event ID in 'e' tag: %s", tag[1])
			}
		}
	}

	if !hasEventTag {
		logger.Warn("NIP-09: Deletion event missing required 'e' tags", 
			zap.String("event_id", evt.ID))
		return fmt.Errorf("deletion event must reference at least one event with 'e' tag")
	}
	
	logger.Debug("NIP-09: Valid deletion event", 
		zap.String("event_id", evt.ID),
		zap.Int("target_events", eventCount))

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
