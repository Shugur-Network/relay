package nips

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-28: Public Chat
// https://github.com/nostr-protocol/nips/blob/master/28.md

// ValidatePublicChat validates NIP-28 public chat events
func ValidatePublicChat(evt *nostr.Event) error {
	switch evt.Kind {
	case 40:
		return validateChannelCreation(evt)
	case 41:
		return validateChannelMessage(evt)
	case 42:
		return validateChannelHideMessage(evt)
	case 43:
		return validateChannelMuteUser(evt)
	case 44:
		return validateChannelHideMessage(evt) // Same validation as kind 42
	default:
		return fmt.Errorf("invalid event kind for public chat: %d", evt.Kind)
	}
}

// validateChannelCreation validates channel creation events (kind 40)
func validateChannelCreation(evt *nostr.Event) error {
	if evt.Kind != 40 {
		return fmt.Errorf("invalid event kind for channel creation: %d", evt.Kind)
	}

	// Content should contain channel metadata (JSON)
	if evt.Content == "" {
		return fmt.Errorf("channel creation must have content with metadata")
	}

	return nil
}

// validateChannelMessage validates channel message events (kind 41)
func validateChannelMessage(evt *nostr.Event) error {
	if evt.Kind != 41 {
		return fmt.Errorf("invalid event kind for channel message: %d", evt.Kind)
	}

	// Must have "e" tag referencing the channel
	hasChannelTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasChannelTag = true
			// Allow any non-empty channel ID (relax validation for compatibility)
			if tag[1] == "" {
				return fmt.Errorf("invalid channel ID in 'e' tag: empty")
			}
			break
		}
	}

	if !hasChannelTag {
		return fmt.Errorf("channel message must reference channel with 'e' tag")
	}

	// Content should contain the message
	if evt.Content == "" {
		return fmt.Errorf("channel message must have content")
	}

	return nil
}

// validateChannelHideMessage validates channel hide message events (kind 42, 44)
func validateChannelHideMessage(evt *nostr.Event) error {
	if evt.Kind != 42 && evt.Kind != 44 {
		return fmt.Errorf("invalid event kind for channel hide message: %d", evt.Kind)
	}

	// Must have "e" tag referencing the message to hide
	hasMessageTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasMessageTag = true
			// Allow any non-empty message ID (relax validation for compatibility)
			if tag[1] == "" {
				return fmt.Errorf("invalid message ID in 'e' tag: empty")
			}
			break
		}
	}

	if !hasMessageTag {
		return fmt.Errorf("channel hide message must reference message with 'e' tag")
	}

	return nil
}

// validateChannelMuteUser validates channel mute user events (kind 43)
func validateChannelMuteUser(evt *nostr.Event) error {
	if evt.Kind != 43 {
		return fmt.Errorf("invalid event kind for channel mute user: %d", evt.Kind)
	}

	// Must have "p" tag referencing the user to mute
	hasUserTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			hasUserTag = true
			// Validate pubkey format
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid pubkey in 'p' tag: %s", tag[1])
			}
			break
		}
	}

	if !hasUserTag {
		return fmt.Errorf("channel mute user must reference user with 'p' tag")
	}

	return nil
}

// IsPublicChat checks if an event is a public chat event
func IsPublicChat(evt *nostr.Event) bool {
	return evt.Kind >= 40 && evt.Kind <= 44
}

// GetPublicChatEventType returns a human-readable type for public chat events
func GetPublicChatEventType(kind int) string {
	switch kind {
	case 40:
		return "channel-creation"
	case 41:
		return "channel-message"
	case 42:
		return "channel-hide-message"
	case 43:
		return "channel-mute-user"
	case 44:
		return "channel-hide-message"
	default:
		return "unknown"
	}
}
