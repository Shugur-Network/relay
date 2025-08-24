package nips

import (
	"fmt"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-17: Private Direct Messages
// https://github.com/nostr-protocol/nips/blob/master/17.md

// ValidatePrivateDirectMessage validates NIP-17 private direct message events
func ValidatePrivateDirectMessage(evt *nostr.Event) error {
	logger.Debug("NIP-17: Validating private direct message", 
		zap.String("event_id", evt.ID),
		zap.Int("kind", evt.Kind),
		zap.String("pubkey", evt.PubKey))
		
	switch evt.Kind {
	case 14:
		return validateChatMessage(evt)
	case 15:
		return validateFileMessage(evt)
	case 1059:
		return validateGiftWrap(evt)
	case 10050:
		return validateDMRelayList(evt)
	default:
		logger.Warn("NIP-17: Invalid event kind for private direct message", 
			zap.String("event_id", evt.ID),
			zap.Int("kind", evt.Kind))
		return fmt.Errorf("invalid event kind for private direct message: %d", evt.Kind)
	}
}

// validateChatMessage validates chat messages (kind 14)
func validateChatMessage(evt *nostr.Event) error {
	if evt.Kind != 14 {
		logger.Warn("NIP-17: Invalid event kind for chat message", 
			zap.String("event_id", evt.ID),
			zap.Int("kind", evt.Kind))
		return fmt.Errorf("invalid event kind for chat message: %d", evt.Kind)
	}

	// Should have "p" tag with recipient pubkey
	hasPTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			hasPTag = true
			// Validate pubkey format
			if len(tag[1]) != 64 {
				logger.Warn("NIP-17: Invalid pubkey in 'p' tag", 
					zap.String("event_id", evt.ID),
					zap.String("invalid_pubkey", tag[1]))
				return fmt.Errorf("invalid pubkey in 'p' tag: %s", tag[1])
			}
			break
		}
	}

	if !hasPTag {
		logger.Warn("NIP-17: Chat message missing required 'p' tag", 
			zap.String("event_id", evt.ID))
		return fmt.Errorf("chat message should have 'p' tag with recipient")
	}

	return nil
}

// validateFileMessage validates file messages (kind 15)
func validateFileMessage(evt *nostr.Event) error {
	if evt.Kind != 15 {
		return fmt.Errorf("invalid event kind for file message: %d", evt.Kind)
	}

	// Should have "p" tag with recipient pubkey
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
		return fmt.Errorf("file message should have 'p' tag with recipient")
	}

	return nil
}

// validateGiftWrap validates gift wrap events (kind 1059)
func validateGiftWrap(evt *nostr.Event) error {
	if evt.Kind != 1059 {
		return fmt.Errorf("invalid event kind for gift wrap: %d", evt.Kind)
	}

	// Must have exactly one "p" tag with recipient pubkey
	recipientCount := 0
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			recipientCount++
			// Validate recipient pubkey format
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid recipient pubkey format in gift wrap event: %s", tag[1])
			}
			// Validate hex format
			for _, c := range tag[1] {
				if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
					return fmt.Errorf("invalid hex format in recipient pubkey: %s", tag[1])
				}
			}
		}
	}
	
	if recipientCount != 1 {
		return fmt.Errorf("gift wrap events must have exactly one recipient")
	}

	// Gift wrap events must have content
	if evt.Content == "" {
		return fmt.Errorf("gift wrap events must have content")
	}

	// Gift wrap events contain encrypted content (NIP-44 encrypted payloads)
	// The content should be a non-empty string, but we don't validate the encryption format
	// as different clients may use different encryption methods

	return nil
}

// validateDMRelayList validates DM relay list events (kind 10050)
func validateDMRelayList(evt *nostr.Event) error {
	if evt.Kind != 10050 {
		return fmt.Errorf("invalid event kind for DM relay list: %d", evt.Kind)
	}

	// Should have "relay" tags
	hasRelayTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			hasRelayTag = true
			break
		}
	}

	if !hasRelayTag {
		return fmt.Errorf("DM relay list should have at least one 'relay' tag")
	}

	return nil
}

// IsPrivateDirectMessage checks if an event is a private direct message
func IsPrivateDirectMessage(evt *nostr.Event) bool {
	return evt.Kind == 14 || evt.Kind == 15 || evt.Kind == 1059 || evt.Kind == 10050
}

// IsGiftWrap checks if an event is a gift wrap
func IsGiftWrap(evt *nostr.Event) bool {
	return evt.Kind == 1059
}
