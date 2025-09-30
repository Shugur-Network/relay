package nips

import (
	"github.com/Shugur-Network/relay/internal/relay/nips/common"
	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-25: Reactions
// https://github.com/nostr-protocol/nips/blob/master/25.md

// ValidateReaction validates NIP-25 reaction events (kind 7)
func ValidateReaction(evt *nostr.Event) error {
	return common.ValidateEventWithCallback(
		evt,
		"25",       // NIP number
		7,          // Expected event kind
		"reaction", // Event name for logging
		func(helper *common.ValidationHelper, event *nostr.Event) error {
			// Validate required tags
			if err := helper.ValidateRequiredTags(event, "e", "p"); err != nil {
				return helper.ErrorFormatter.FormatError("%v", err)
			}

			// Validate event ID format in "e" tags
			for _, tag := range event.Tags {
				if len(tag) >= 2 && tag[0] == "e" {
					if err := helper.ValidateEventID(tag[1]); err != nil {
						return helper.FormatTagError("e", "invalid event ID: %v", err)
					}
				}
			}

			// Validate pubkey format in "p" tags
			for _, tag := range event.Tags {
				if len(tag) >= 2 && tag[0] == "p" {
					if err := helper.ValidatePubkey(tag[1]); err != nil {
						return helper.FormatTagError("p", "invalid pubkey: %v", err)
					}
				}
			}

			// Content should contain the reaction (usually emoji or "+"/"-")
			// Empty content is allowed (interpreted as "like")

			return nil
		},
	)
}

// IsReaction checks if an event is a reaction
func IsReaction(evt *nostr.Event) bool {
	return evt.Kind == 7
}

// GetReactionContent returns the reaction content or default "like" for empty content
func GetReactionContent(evt *nostr.Event) string {
	if evt.Content == "" {
		return "+" // Default like reaction
	}
	return evt.Content
}
