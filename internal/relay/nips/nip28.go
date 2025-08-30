package nips

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-28: Public Chat
// https://github.com/nostr-protocol/nips/blob/master/28.md
//
// Event kinds:
// - 40: channel create
// - 41: channel metadata
// - 42: channel message
// - 43: hide message
// - 44: mute user

// ValidatePublicChat validates NIP-28 public chat events
func ValidatePublicChat(evt *nostr.Event) error {
	logger.Debug("NIP-28: Validating public chat event",
		zap.String("event_id", evt.ID),
		zap.Int("kind", evt.Kind))

	switch evt.Kind {
	case 40:
		return validateChannelCreate(evt)
	case 41:
		return validateChannelMetadata(evt)
	case 42:
		return validateChannelMessage(evt)
	case 43:
		return validateHideMessage(evt)
	case 44:
		return validateMuteUser(evt)
	default:
		return fmt.Errorf("invalid event kind for public chat: %d", evt.Kind)
	}
}

// validateChannelCreate validates channel creation events (kind 40)
func validateChannelCreate(evt *nostr.Event) error {
	if evt.Kind != 40 {
		return fmt.Errorf("invalid event kind for channel creation: %d", evt.Kind)
	}

	// Per NIP-28: "Client SHOULD include basic channel metadata" - treating as mandatory
	if evt.Content == "" {
		return fmt.Errorf("channel creation must have content with metadata (NIP-28 SHOULD requirement)")
	}

	// Validate content as JSON metadata
	var metadata ChannelMetadata
	if err := json.Unmarshal([]byte(evt.Content), &metadata); err != nil {
		return fmt.Errorf("invalid channel metadata JSON: %v", err)
	}

	// Name is required for basic channel metadata
	if metadata.Name == "" {
		return fmt.Errorf("channel name is required in metadata (NIP-28 basic metadata)")
	}

	// Validate relays format if present
	if metadata.Relays != nil {
		for _, relay := range metadata.Relays {
			if !strings.HasPrefix(relay, "wss://") && !strings.HasPrefix(relay, "ws://") {
				return fmt.Errorf("invalid relay URL format: %s", relay)
			}
		}
	}

	logger.Info("NIP-28: Channel created",
		zap.String("channel_id", evt.ID),
		zap.String("name", metadata.Name),
		zap.String("creator", evt.PubKey))

	return nil
}

// validateChannelMetadata validates channel metadata update events (kind 41)
func validateChannelMetadata(evt *nostr.Event) error {
	if evt.Kind != 41 {
		return fmt.Errorf("invalid event kind for channel metadata: %d", evt.Kind)
	}

	// Must have an "e" tag referencing the channel creation event
	channelRef := findChannelReference(evt)
	if channelRef == "" {
		return fmt.Errorf("channel metadata update must reference channel creation event with 'e' tag")
	}

	// Content should contain updated metadata
	if evt.Content == "" {
		return fmt.Errorf("channel metadata update must have content")
	}

	// Validate metadata JSON format
	var metadata ChannelMetadata
	if err := json.Unmarshal([]byte(evt.Content), &metadata); err != nil {
		return fmt.Errorf("invalid channel metadata JSON: %v", err)
	}

	// Validate relays format if present
	if metadata.Relays != nil {
		for _, relay := range metadata.Relays {
			if !strings.HasPrefix(relay, "wss://") && !strings.HasPrefix(relay, "ws://") {
				return fmt.Errorf("invalid relay URL format: %s", relay)
			}
		}
	}

	logger.Info("NIP-28: Channel metadata updated",
		zap.String("channel_id", channelRef),
		zap.String("update_id", evt.ID),
		zap.String("updater", evt.PubKey))

	return nil
}

// validateChannelMessage validates channel message events (kind 42)
func validateChannelMessage(evt *nostr.Event) error {
	if evt.Kind != 42 {
		return fmt.Errorf("invalid event kind for channel message: %d", evt.Kind)
	}

	// Must have content
	if evt.Content == "" {
		return fmt.Errorf("channel message must have content")
	}

	// Must reference a channel with "e" tag marked as "root"
	channelRef := findChannelReference(evt)
	if channelRef == "" {
		return fmt.Errorf("channel message must reference channel with 'e' tag marked as 'root'")
	}

	// If it's a reply, validate reply structure
	if isReply(evt) {
		if err := validateReplyStructure(evt); err != nil {
			return fmt.Errorf("invalid reply structure: %v", err)
		}
	}

	logger.Debug("NIP-28: Channel message validated",
		zap.String("channel_id", channelRef),
		zap.String("message_id", evt.ID),
		zap.String("author", evt.PubKey),
		zap.Bool("is_reply", isReply(evt)))

	return nil
}

// validateHideMessage validates hide message events (kind 43)
func validateHideMessage(evt *nostr.Event) error {
	if evt.Kind != 43 {
		return fmt.Errorf("invalid event kind for hide message: %d", evt.Kind)
	}

	// Must have "e" tag referencing the message to hide
	messageRef := findMessageReference(evt)
	if messageRef == "" {
		return fmt.Errorf("hide message event must reference message with 'e' tag")
	}

	// Content may contain optional reason
	if evt.Content != "" {
		var reason HideReason
		if err := json.Unmarshal([]byte(evt.Content), &reason); err != nil {
			return fmt.Errorf("invalid hide reason JSON: %v", err)
		}
	}

	logger.Info("NIP-28: Message hidden",
		zap.String("message_id", messageRef),
		zap.String("hidden_by", evt.PubKey))

	return nil
}

// validateMuteUser validates mute user events (kind 44)
func validateMuteUser(evt *nostr.Event) error {
	if evt.Kind != 44 {
		return fmt.Errorf("invalid event kind for mute user: %d", evt.Kind)
	}

	// Must have "p" tag referencing the user to mute
	userRef := findUserReference(evt)
	if userRef == "" {
		return fmt.Errorf("mute user event must reference user with 'p' tag")
	}

	// Validate pubkey format
	if len(userRef) != 64 {
		return fmt.Errorf("invalid pubkey format in 'p' tag")
	}

	// Content may contain optional reason
	if evt.Content != "" {
		var reason MuteReason
		if err := json.Unmarshal([]byte(evt.Content), &reason); err != nil {
			return fmt.Errorf("invalid mute reason JSON: %v", err)
		}
	}

	logger.Info("NIP-28: User muted",
		zap.String("muted_user", userRef),
		zap.String("muted_by", evt.PubKey))

	return nil
}

// Helper functions

// findChannelReference finds the channel reference in "e" tags
func findChannelReference(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 4 && tag[0] == "e" && tag[3] == "root" {
			return tag[1]
		} else if len(tag) == 2 && tag[0] == "e" {
			// Check if the second element contains comma-separated values with root marker
			parts := strings.Split(tag[1], ",")
			if len(parts) >= 3 && strings.TrimSpace(parts[2]) == "root" {
				return strings.TrimSpace(parts[0])
			}
		}
	}

	// Fallback: look for any "e" tag if no "root" marker found
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			// Extract just the event ID part (before any comma)
			parts := strings.Split(tag[1], ",")
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}

// findMessageReference finds the message reference in "e" tags
func findMessageReference(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			// Extract just the event ID part (before any comma)
			parts := strings.Split(tag[1], ",")
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}

// findUserReference finds the user reference in "p" tags
func findUserReference(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			return tag[1]
		}
	}
	return ""
}

// isReply checks if the event is a reply to another message
func isReply(evt *nostr.Event) bool {
	rootCount := 0
	replyCount := 0

	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			// Handle both standard format ["e", "id", "relay", "marker"]
			// and comma-separated format ["e", "id,relay,marker"]
			if len(tag) >= 4 && tag[3] == "reply" {
				replyCount++
			} else if len(tag) >= 4 && tag[3] == "root" {
				rootCount++
			} else if len(tag) == 2 {
				// Check if the second element contains comma-separated values
				parts := strings.Split(tag[1], ",")
				if len(parts) >= 3 {
					marker := strings.TrimSpace(parts[2])
					switch marker {
					case "reply":
						replyCount++
					case "root":
						rootCount++
					}
				}
			}
		}
	}

	// It's a reply if there's any reply tag (proper or improper structure)
	// This allows us to validate reply structure even for malformed replies
	return replyCount > 0
}

// validateReplyStructure validates the structure of reply messages
func validateReplyStructure(evt *nostr.Event) error {
	hasRoot := false
	hasReply := false
	hasPTag := false

	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			// Handle both standard format ["e", "id", "relay", "marker"]
			// and comma-separated format ["e", "id,relay,marker"]
			if len(tag) >= 4 {
				switch tag[3] {
				case "root":
					hasRoot = true
				case "reply":
					hasReply = true
				}
			} else if len(tag) == 2 {
				// Check if the second element contains comma-separated values
				parts := strings.Split(tag[1], ",")
				if len(parts) >= 3 {
					marker := strings.TrimSpace(parts[2])
					switch marker {
					case "root":
						hasRoot = true
					case "reply":
						hasReply = true
					}
				}
			}
		}
		if len(tag) >= 2 && tag[0] == "p" {
			hasPTag = true
		}
	}

	if !hasRoot {
		return fmt.Errorf("reply must have 'e' tag marked as 'root'")
	}
	if !hasReply {
		return fmt.Errorf("reply must have 'e' tag marked as 'reply'")
	}
	// Per NIP-28: "Clients SHOULD append p tags to replies" - treating as mandatory
	if !hasPTag {
		return fmt.Errorf("reply must have 'p' tag referencing the author being replied to (NIP-28 SHOULD requirement)")
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
		return "channel_create"
	case 41:
		return "channel_metadata"
	case 42:
		return "channel_message"
	case 43:
		return "hide_message"
	case 44:
		return "mute_user"
	default:
		return "unknown"
	}
}

// Data structures for NIP-28

// ChannelMetadata represents channel metadata structure
type ChannelMetadata struct {
	Name    string   `json:"name"`
	About   string   `json:"about,omitempty"`
	Picture string   `json:"picture,omitempty"`
	Relays  []string `json:"relays,omitempty"`
}

// HideReason represents the reason for hiding a message
type HideReason struct {
	Reason string `json:"reason,omitempty"`
}

// MuteReason represents the reason for muting a user
type MuteReason struct {
	Reason string `json:"reason,omitempty"`
}
