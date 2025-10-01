package nips

import (
	"fmt"

	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/relay/nips/common"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-09: Event Deletion
// https://github.com/nostr-protocol/nips/blob/master/09.md

// ValidateEventDeletion validates NIP-09 event deletion events (kind 5)
func ValidateEventDeletion(evt *nostr.Event) error {
	return common.ValidateEventWithCallback(
		evt,
		"09",             // NIP number
		5,                // Expected event kind
		"event deletion", // Event name for logging
		func(helper *common.ValidationHelper, event *nostr.Event) error {
			// Must have at least one "e" tag referencing the event(s) to delete
			if err := helper.ValidateRequiredTag(event, "e"); err != nil {
				return helper.ErrorFormatter.FormatError("deletion event must reference at least one event with 'e' tag")
			}

			// Validate event ID format in "e" tags and count them
			eventCount := 0
			for _, tag := range event.Tags {
				if len(tag) >= 2 && tag[0] == "e" {
					eventCount++
					if err := helper.ValidateEventID(tag[1]); err != nil {
						logger.Warn("NIP-09: Invalid event ID in 'e' tag",
							zap.String("deletion_event_id", event.ID),
							zap.String("invalid_event_id", tag[1]))
						return helper.FormatTagError("e", "invalid event ID: %v", err)
					}
				}
			}

			// Log the count of target events for debugging
			logger.Debug("NIP-09: Valid deletion event",
				zap.String("event_id", event.ID),
				zap.Int("target_events", eventCount))

			return nil
		},
	)
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
