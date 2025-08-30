package nips

import (
	"fmt"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-25: Reactions
// https://github.com/nostr-protocol/nips/blob/master/25.md

// ValidateReaction validates NIP-25 reaction events (kind 7)
func ValidateReaction(evt *nostr.Event) error {
	logger.Debug("NIP-25: Validating reaction event",
		zap.String("event_id", evt.ID),
		zap.String("pubkey", evt.PubKey))

	if evt.Kind != 7 {
		logger.Warn("NIP-25: Invalid event kind for reaction",
			zap.String("event_id", evt.ID),
			zap.Int("kind", evt.Kind))
		return fmt.Errorf("invalid event kind for reaction: %d", evt.Kind)
	}

	// Must have at least one "e" tag referencing the reacted event
	hasEventTag := false
	hasPubkeyTag := false

	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e":
				hasEventTag = true
				// Validate event ID format
				if len(tag[1]) != 64 {
					return fmt.Errorf("invalid event ID in 'e' tag: %s", tag[1])
				}
			case "p":
				hasPubkeyTag = true
				// Validate pubkey format
				if len(tag[1]) != 64 {
					return fmt.Errorf("invalid pubkey in 'p' tag: %s", tag[1])
				}
			}
		}
	}

	if !hasEventTag {
		return fmt.Errorf("reaction must reference at least one event with 'e' tag")
	}

	if !hasPubkeyTag {
		return fmt.Errorf("reaction must reference the author with 'p' tag")
	}

	// Content should contain the reaction (usually emoji or "+"/"-")
	// Empty content is allowed (interpreted as "like")

	return nil
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
