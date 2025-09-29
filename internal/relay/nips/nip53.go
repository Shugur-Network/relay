package nips

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	nostr "github.com/nbd-wtf/go-nostr"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// ValidateLiveStreamingEvent validates NIP-53 live streaming events (kind 30311)
func ValidateLiveStreamingEvent(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-53: Validating live streaming event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 30311 {
		return fmt.Errorf("invalid kind for live streaming event: expected 30311, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateLiveStreamingEventTags(event); err != nil {
		return fmt.Errorf("invalid live streaming event tags: %w", err)
	}

	logger.Debug("NIP-53: Live streaming event validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateLiveChatMessage validates NIP-53 live chat messages (kind 1311)
func ValidateLiveChatMessage(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-53: Validating live chat message",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 1311 {
		return fmt.Errorf("invalid kind for live chat message: expected 1311, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateLiveChatMessageTags(event); err != nil {
		return fmt.Errorf("invalid live chat message tags: %w", err)
	}

	logger.Debug("NIP-53: Live chat message validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateMeetingSpace validates NIP-53 meeting space events (kind 30312)
func ValidateMeetingSpace(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-53: Validating meeting space event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 30312 {
		return fmt.Errorf("invalid kind for meeting space: expected 30312, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateMeetingSpaceTags(event); err != nil {
		return fmt.Errorf("invalid meeting space tags: %w", err)
	}

	logger.Debug("NIP-53: Meeting space validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateMeetingRoomEvent validates NIP-53 meeting room events (kind 30313)
func ValidateMeetingRoomEvent(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-53: Validating meeting room event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 30313 {
		return fmt.Errorf("invalid kind for meeting room event: expected 30313, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateMeetingRoomEventTags(event); err != nil {
		return fmt.Errorf("invalid meeting room event tags: %w", err)
	}

	logger.Debug("NIP-53: Meeting room event validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateRoomPresence validates NIP-53 room presence events (kind 10312)
func ValidateRoomPresence(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-53: Validating room presence event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 10312 {
		return fmt.Errorf("invalid kind for room presence: expected 10312, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateRoomPresenceTags(event); err != nil {
		return fmt.Errorf("invalid room presence tags: %w", err)
	}

	logger.Debug("NIP-53: Room presence validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// validateLiveStreamingEventTags validates tags for live streaming events
func validateLiveStreamingEventTags(event *nostr.Event) error {
	var hasDTag bool
	var hasHostParticipant bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateLiveActivityDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "title":
			if err := validateLiveActivityTitleTag(tag); err != nil {
				return err
			}
		case "summary":
			if err := validateLiveActivitySummaryTag(tag); err != nil {
				return err
			}
		case "image":
			if err := validateImageURL(tag[1]); err != nil {
				return fmt.Errorf("invalid image tag: %w", err)
			}
		case "streaming":
			if err := validateStreamingURLTag(tag); err != nil {
				return err
			}
		case "recording":
			if err := validateRecordingURLTag(tag); err != nil {
				return err
			}
		case "starts":
			if err := validateTimestampTagNIP53(tag, "starts"); err != nil {
				return err
			}
		case "ends":
			if err := validateTimestampTagNIP53(tag, "ends"); err != nil {
				return err
			}
		case "status":
			if err := validateLiveStreamingStatusTag(tag); err != nil {
				return err
			}
		case "current_participants":
			if err := validateParticipantCountTag(tag, "current_participants"); err != nil {
				return err
			}
		case "total_participants":
			if err := validateParticipantCountTag(tag, "total_participants"); err != nil {
				return err
			}
		case "p":
			if err := validateLiveStreamingParticipantTag(tag); err != nil {
				return err
			}
			// Check if this is a Host participant - handle both formats
			var role string
			if len(tag) == 2 && strings.Contains(tag[1], ",") {
				// Comma-separated format: "pubkey,relay,role,proof"
				parts := strings.Split(tag[1], ",")
				if len(parts) > 2 {
					role = parts[2]
				}
			} else if len(tag) >= 4 {
				// Separate elements format
				role = tag[3]
			}
			
			if strings.ToLower(role) == "host" {
				hasHostParticipant = true
			}
		case "t":
			if err := validateHashtagTag(tag); err != nil {
				return err
			}
		case "relays":
			if err := validateRelaysTagNIP53(tag); err != nil {
				return err
			}
		case "pinned":
			if err := validatePinnedMessageTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("live streaming event must have a d tag")
	}

	// Optional but recommended: at least one Host participant
	if !hasHostParticipant {
		logger.Warn("NIP-53: Live streaming event should have at least one Host participant",
			zap.String("event_id", event.ID))
	}

	return nil
}

// validateLiveChatMessageTags validates tags for live chat messages
func validateLiveChatMessageTags(event *nostr.Event) error {
	var hasATag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "a":
			if err := validateLiveActivityReferenceTag(tag); err != nil {
				return err
			}
			hasATag = true
		case "e":
			if err := validateLiveChatReplyTag(tag); err != nil {
				return err
			}
		case "q":
			if err := validateQuoteTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasATag {
		return fmt.Errorf("live chat message must have an a tag referencing the activity")
	}

	return nil
}

// validateMeetingSpaceTags validates tags for meeting space events
func validateMeetingSpaceTags(event *nostr.Event) error {
	var hasDTag bool
	var hasRoomTag bool
	var hasStatusTag bool
	var hasServiceTag bool
	var hasHostParticipant bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateLiveActivityDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "room":
			if err := validateRoomNameTag(tag); err != nil {
				return err
			}
			hasRoomTag = true
		case "summary":
			if err := validateLiveActivitySummaryTag(tag); err != nil {
				return err
			}
		case "image":
			if err := validateImageURL(tag[1]); err != nil {
				return fmt.Errorf("invalid image tag: %w", err)
			}
		case "status":
			if err := validateMeetingSpaceStatusTag(tag); err != nil {
				return err
			}
			hasStatusTag = true
		case "service":
			if err := validateServiceURLTag(tag); err != nil {
				return err
			}
			hasServiceTag = true
		case "endpoint":
			if err := validateEndpointURLTag(tag); err != nil {
				return err
			}
		case "t":
			if err := validateHashtagTag(tag); err != nil {
				return err
			}
		case "p":
			if err := validateMeetingSpaceParticipantTag(tag); err != nil {
				return err
			}
			// Check if this is a Host/Owner participant - handle both formats
			var role string
			if len(tag) == 2 && strings.Contains(tag[1], ",") {
				// Comma-separated format: "pubkey,relay,role,proof"
				parts := strings.Split(tag[1], ",")
				if len(parts) > 2 {
					role = parts[2]
				}
			} else if len(tag) >= 4 {
				// Separate elements format
				role = tag[3]
			}
			
			if strings.ToLower(role) == "host" || strings.ToLower(role) == "owner" {
				hasHostParticipant = true
			}
		case "relays":
			if err := validateRelaysTagNIP53(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("meeting space must have a d tag")
	}

	if !hasRoomTag {
		return fmt.Errorf("meeting space must have a room tag")
	}

	if !hasStatusTag {
		return fmt.Errorf("meeting space must have a status tag")
	}

	if !hasServiceTag {
		return fmt.Errorf("meeting space must have a service tag")
	}

	if !hasHostParticipant {
		return fmt.Errorf("meeting space must have at least one Host or Owner participant")
	}

	return nil
}

// validateMeetingRoomEventTags validates tags for meeting room events
func validateMeetingRoomEventTags(event *nostr.Event) error {
	var hasDTag bool
	var hasATag bool
	var hasTitleTag bool
	var hasStartsTag bool
	var hasStatusTag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateLiveActivityDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "a":
			if err := validateMeetingSpaceReferenceTag(tag); err != nil {
				return err
			}
			hasATag = true
		case "title":
			if err := validateLiveActivityTitleTag(tag); err != nil {
				return err
			}
			hasTitleTag = true
		case "summary":
			if err := validateLiveActivitySummaryTag(tag); err != nil {
				return err
			}
		case "image":
			if err := validateImageURL(tag[1]); err != nil {
				return fmt.Errorf("invalid image tag: %w", err)
			}
		case "starts":
			if err := validateTimestampTagNIP53(tag, "starts"); err != nil {
				return err
			}
			hasStartsTag = true
		case "ends":
			if err := validateTimestampTagNIP53(tag, "ends"); err != nil {
				return err
			}
		case "status":
			if err := validateMeetingRoomStatusTag(tag); err != nil {
				return err
			}
			hasStatusTag = true
		case "current_participants":
			if err := validateParticipantCountTag(tag, "current_participants"); err != nil {
				return err
			}
		case "total_participants":
			if err := validateParticipantCountTag(tag, "total_participants"); err != nil {
				return err
			}
		case "p":
			if err := validateMeetingRoomParticipantTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("meeting room event must have a d tag")
	}

	if !hasATag {
		return fmt.Errorf("meeting room event must have an a tag referencing the parent space")
	}

	if !hasTitleTag {
		return fmt.Errorf("meeting room event must have a title tag")
	}

	if !hasStartsTag {
		return fmt.Errorf("meeting room event must have a starts tag")
	}

	if !hasStatusTag {
		return fmt.Errorf("meeting room event must have a status tag")
	}

	return nil
}

// validateRoomPresenceTags validates tags for room presence events
func validateRoomPresenceTags(event *nostr.Event) error {
	var hasATag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "a":
			if err := validateRoomPresenceReferenceTag(tag); err != nil {
				return err
			}
			hasATag = true
		case "hand":
			if err := validateHandRaisedTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasATag {
		return fmt.Errorf("room presence event must have an a tag referencing the room")
	}

	return nil
}

// validateLiveActivityDTag validates the d tag for live activities
func validateLiveActivityDTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("d tag must have exactly 2 elements")
	}

	identifier := tag[1]
	if identifier == "" {
		return fmt.Errorf("d tag identifier cannot be empty")
	}

	if len(identifier) > 200 {
		return fmt.Errorf("d tag identifier too long (max 200 characters)")
	}

	return nil
}

// validateLiveActivityTitleTag validates the title tag for live activities
func validateLiveActivityTitleTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("title tag must have exactly 2 elements")
	}

	title := tag[1]
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if len(title) > 500 {
		return fmt.Errorf("title too long (max 500 characters)")
	}

	return nil
}

// validateLiveActivitySummaryTag validates the summary tag for live activities
func validateLiveActivitySummaryTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("summary tag must have exactly 2 elements")
	}

	summary := tag[1]
	if len(summary) > 2000 {
		return fmt.Errorf("summary too long (max 2000 characters)")
	}

	return nil
}

// validateStreamingURLTag validates the streaming tag for live streaming events
func validateStreamingURLTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("streaming tag must have exactly 2 elements")
	}

	streamingURL := tag[1]
	if streamingURL == "" {
		return fmt.Errorf("streaming URL cannot be empty")
	}

	// Validate URL format
	u, err := url.Parse(streamingURL)
	if err != nil {
		return fmt.Errorf("invalid streaming URL format: %w", err)
	}

	// Allow http/https schemes for streaming URLs
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("streaming URL must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("streaming URL must have a host")
	}

	if len(streamingURL) > 2000 {
		return fmt.Errorf("streaming URL too long (max 2000 characters)")
	}

	return nil
}

// validateRecordingURLTag validates the recording tag for live streaming events
func validateRecordingURLTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("recording tag must have exactly 2 elements")
	}

	recordingURL := tag[1]
	if recordingURL == "" {
		return fmt.Errorf("recording URL cannot be empty")
	}

	// Validate URL format
	u, err := url.Parse(recordingURL)
	if err != nil {
		return fmt.Errorf("invalid recording URL format: %w", err)
	}

	// Allow http/https schemes for recording URLs
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("recording URL must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("recording URL must have a host")
	}

	if len(recordingURL) > 2000 {
		return fmt.Errorf("recording URL too long (max 2000 characters)")
	}

	return nil
}

// validateTimestampTagNIP53 validates timestamp tags (starts/ends) for live activities
func validateTimestampTagNIP53(tag nostr.Tag, tagName string) error {
	if len(tag) != 2 {
		return fmt.Errorf("%s tag must have exactly 2 elements", tagName)
	}

	timestampStr := tag[1]
	if timestampStr == "" {
		return fmt.Errorf("%s timestamp cannot be empty", tagName)
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s timestamp: %w", tagName, err)
	}

	// Validate timestamp is reasonable (not negative, not too far in future)
	if timestamp < 0 {
		return fmt.Errorf("%s timestamp cannot be negative", tagName)
	}

	// Don't allow timestamps too far in the future (10 years)
	maxFutureTime := time.Now().AddDate(10, 0, 0).Unix()
	if timestamp > maxFutureTime {
		return fmt.Errorf("%s timestamp is too far in the future", tagName)
	}

	return nil
}

// validateLiveStreamingStatusTag validates the status tag for live streaming events
func validateLiveStreamingStatusTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("status tag must have exactly 2 elements")
	}

	status := tag[1]
	validStatuses := map[string]bool{
		"planned": true,
		"live":    true,
		"ended":   true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be 'planned', 'live', or 'ended'")
	}

	return nil
}

// validateMeetingSpaceStatusTag validates the status tag for meeting spaces
func validateMeetingSpaceStatusTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("status tag must have exactly 2 elements")
	}

	status := tag[1]
	validStatuses := map[string]bool{
		"open":    true,
		"private": true,
		"closed":  true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be 'open', 'private', or 'closed'")
	}

	return nil
}

// validateMeetingRoomStatusTag validates the status tag for meeting room events
func validateMeetingRoomStatusTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("status tag must have exactly 2 elements")
	}

	status := tag[1]
	validStatuses := map[string]bool{
		"planned": true,
		"live":    true,
		"ended":   true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be 'planned', 'live', or 'ended'")
	}

	return nil
}

// validateParticipantCountTag validates participant count tags
func validateParticipantCountTag(tag nostr.Tag, tagName string) error {
	if len(tag) != 2 {
		return fmt.Errorf("%s tag must have exactly 2 elements", tagName)
	}

	countStr := tag[1]
	if countStr == "" {
		return fmt.Errorf("%s count cannot be empty", tagName)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return fmt.Errorf("invalid %s count: %w", tagName, err)
	}

	if count < 0 {
		return fmt.Errorf("%s count cannot be negative", tagName)
	}

	// Reasonable upper limit
	if count > 1000000 {
		return fmt.Errorf("%s count too large (max 1,000,000)", tagName)
	}

	return nil
}

// validateLiveStreamingParticipantTag validates the p tag for live streaming events
func validateLiveStreamingParticipantTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("p tag must have at least 2 elements")
	}

	// Handle comma-separated format from nak tool: "pubkey,relay,role,proof"
	// or separate elements: ["p", "pubkey", "relay", "role", "proof"]
	var pubkey, relay, role, proof string
	
	if len(tag) == 2 && strings.Contains(tag[1], ",") {
		// Comma-separated format
		parts := strings.Split(tag[1], ",")
		if len(parts) < 1 || len(parts) > 4 {
			return fmt.Errorf("participant info must have 1-4 comma-separated parts")
		}
		pubkey = parts[0]
		if len(parts) > 1 {
			relay = parts[1]
		}
		if len(parts) > 2 {
			role = parts[2]
		}
		if len(parts) > 3 {
			proof = parts[3]
		}
	} else {
		// Separate elements format
		if len(tag) > 5 {
			return fmt.Errorf("p tag must have at most 5 elements")
		}
		pubkey = tag[1]
		if len(tag) > 2 {
			relay = tag[2]
		}
		if len(tag) > 3 {
			role = tag[3]
		}
		if len(tag) > 4 {
			proof = tag[4]
		}
	}

	// Validate pubkey
	if len(pubkey) != 64 {
		return fmt.Errorf("participant pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	if !isHexChar64(pubkey) {
		return fmt.Errorf("participant pubkey must be valid hex")
	}

	// Optional relay hint validation
	if relay != "" {
		if err := validateRelayURL(relay); err != nil {
			return fmt.Errorf("invalid participant relay hint: %w", err)
		}
	}

	// Optional role validation
	if role != "" {
		validRoles := map[string]bool{
			"Host":        true,
			"Speaker":     true,
			"Participant": true,
			"Moderator":   true,
			"Owner":       true,
		}
		
		if !validRoles[role] {
			// Allow custom roles but warn
			if len(role) > 50 {
				return fmt.Errorf("participant role too long (max 50 characters)")
			}
		}
	}

	// Optional proof validation
	if proof != "" {
		if len(proof) != 128 {
			return fmt.Errorf("participant proof must be 128 hex characters, got %d", len(proof))
		}
		
		if !regexp.MustCompile(`^[a-fA-F0-9]{128}$`).MatchString(proof) {
			return fmt.Errorf("participant proof must be valid hex")
		}
	}

	return nil
}

// validateMeetingSpaceParticipantTag validates the p tag for meeting spaces
func validateMeetingSpaceParticipantTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("p tag must have at least 2 elements")
	}

	// Handle comma-separated format from nak tool: "pubkey,relay,role,proof"
	// or separate elements: ["p", "pubkey", "relay", "role", "proof"]
	var pubkey, relay, role, proof string
	
	if len(tag) == 2 && strings.Contains(tag[1], ",") {
		// Comma-separated format
		parts := strings.Split(tag[1], ",")
		if len(parts) < 1 || len(parts) > 4 {
			return fmt.Errorf("participant info must have 1-4 comma-separated parts")
		}
		pubkey = parts[0]
		if len(parts) > 1 {
			relay = parts[1]
		}
		if len(parts) > 2 {
			role = parts[2]
		}
		if len(parts) > 3 {
			proof = parts[3]
		}
	} else {
		// Separate elements format
		if len(tag) > 5 {
			return fmt.Errorf("p tag must have at most 5 elements")
		}
		pubkey = tag[1]
		if len(tag) > 2 {
			relay = tag[2]
		}
		if len(tag) > 3 {
			role = tag[3]
		}
		if len(tag) > 4 {
			proof = tag[4]
		}
	}

	// Validate pubkey
	if len(pubkey) != 64 {
		return fmt.Errorf("participant pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	if !isHexChar64(pubkey) {
		return fmt.Errorf("participant pubkey must be valid hex")
	}

	// Optional relay hint validation
	if relay != "" {
		if err := validateRelayURL(relay); err != nil {
			return fmt.Errorf("invalid participant relay hint: %w", err)
		}
	}

	// Optional role validation
	if role != "" {
		validRoles := map[string]bool{
			"Owner":     true,
			"Host":      true,
			"Moderator": true,
			"Speaker":   true,
		}
		
		if !validRoles[role] {
			// Allow custom roles but validate length
			if len(role) > 50 {
				return fmt.Errorf("participant role too long (max 50 characters)")
			}
		}
	}

	// Optional proof validation
	if proof != "" {
		if len(proof) != 128 {
			return fmt.Errorf("participant proof must be 128 hex characters, got %d", len(proof))
		}
		
		if !regexp.MustCompile(`^[a-fA-F0-9]{128}$`).MatchString(proof) {
			return fmt.Errorf("participant proof must be valid hex")
		}
	}

	return nil
}

// validateMeetingRoomParticipantTag validates the p tag for meeting room events
func validateMeetingRoomParticipantTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("p tag must have at least 2 elements")
	}

	// Handle comma-separated format from nak tool: "pubkey,relay,role"
	// or separate elements: ["p", "pubkey", "relay", "role"]
	var pubkey, relay, role string
	
	if len(tag) == 2 && strings.Contains(tag[1], ",") {
		// Comma-separated format
		parts := strings.Split(tag[1], ",")
		if len(parts) < 1 || len(parts) > 3 {
			return fmt.Errorf("participant info must have 1-3 comma-separated parts")
		}
		pubkey = parts[0]
		if len(parts) > 1 {
			relay = parts[1]
		}
		if len(parts) > 2 {
			role = parts[2]
		}
	} else {
		// Separate elements format
		if len(tag) > 4 {
			return fmt.Errorf("p tag must have at most 4 elements")
		}
		pubkey = tag[1]
		if len(tag) > 2 {
			relay = tag[2]
		}
		if len(tag) > 3 {
			role = tag[3]
		}
	}

	// Validate pubkey
	if len(pubkey) != 64 {
		return fmt.Errorf("participant pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	if !isHexChar64(pubkey) {
		return fmt.Errorf("participant pubkey must be valid hex")
	}

	// Optional relay hint validation
	if relay != "" {
		if err := validateRelayURL(relay); err != nil {
			return fmt.Errorf("invalid participant relay hint: %w", err)
		}
	}

	// Optional role validation
	if role != "" {
		validRoles := map[string]bool{
			"Speaker":     true,
			"Participant": true,
			"Moderator":   true,
		}
		
		if !validRoles[role] {
			// Allow custom roles but validate length
			if len(role) > 50 {
				return fmt.Errorf("participant role too long (max 50 characters)")
			}
		}
	}

	return nil
}

// validateHashtagTag validates hashtag tags (t)
func validateHashtagTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("t tag must have exactly 2 elements")
	}

	hashtag := tag[1]
	if hashtag == "" {
		return fmt.Errorf("hashtag cannot be empty")
	}

	if len(hashtag) > 100 {
		return fmt.Errorf("hashtag too long (max 100 characters)")
	}

	// Hashtags should not contain spaces or special characters
	if strings.ContainsAny(hashtag, " \t\n\r") {
		return fmt.Errorf("hashtag cannot contain whitespace")
	}

	return nil
}

// validateRelaysTagNIP53 validates the relays tag for NIP-53
func validateRelaysTagNIP53(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("relays tag must have at least 2 elements")
	}

	// Handle both formats:
	// 1. Comma-separated in single element: ["relays", "wss://relay1.com,wss://relay2.com"]
	// 2. Multiple elements: ["relays", "wss://relay1.com", "wss://relay2.com"]
	
	var relayURLs []string
	
	if len(tag) == 2 && strings.Contains(tag[1], ",") {
		// Comma-separated format
		relayURLs = strings.Split(tag[1], ",")
	} else {
		// Multiple elements format
		for i := 1; i < len(tag); i++ {
			relayURLs = append(relayURLs, tag[i])
		}
	}

	// Validate each relay URL
	for i, relayURL := range relayURLs {
		relayURL = strings.TrimSpace(relayURL)
		if relayURL == "" {
			return fmt.Errorf("relay URL %d cannot be empty", i+1)
		}
		
		if err := validateRelayURL(relayURL); err != nil {
			return fmt.Errorf("invalid relay URL %d: %w", i+1, err)
		}
	}

	return nil
}

// validatePinnedMessageTag validates the pinned message tag
func validatePinnedMessageTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("pinned tag must have exactly 2 elements")
	}

	eventID := tag[1]
	if len(eventID) != 64 {
		return fmt.Errorf("pinned event ID must be 64 hex characters, got %d", len(eventID))
	}

	if !isHexChar64(eventID) {
		return fmt.Errorf("pinned event ID must be valid hex")
	}

	return nil
}

// validateLiveActivityReferenceTag validates the a tag for live activity references
func validateLiveActivityReferenceTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("a tag must have 2 or 3 elements")
	}

	aTagValue := tag[1]
	if err := validateLiveActivityReference(aTagValue); err != nil {
		return fmt.Errorf("invalid live activity reference: %w", err)
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid live activity reference relay hint: %w", err)
		}
	}

	return nil
}

// validateMeetingSpaceReferenceTag validates the a tag for meeting space references
func validateMeetingSpaceReferenceTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("a tag must have 2 or 3 elements")
	}

	aTagValue := tag[1]
	if err := validateMeetingSpaceReference(aTagValue); err != nil {
		return fmt.Errorf("invalid meeting space reference: %w", err)
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid meeting space reference relay hint: %w", err)
		}
	}

	return nil
}

// validateRoomPresenceReferenceTag validates the a tag for room presence references
func validateRoomPresenceReferenceTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 4 {
		return fmt.Errorf("a tag must have 2-4 elements")
	}

	aTagValue := tag[1]
	// Room presence can reference either meeting spaces (30312) or live streaming (30311)
	if err := validateRoomReference(aTagValue); err != nil {
		return fmt.Errorf("invalid room reference: %w", err)
	}

	// Optional relay hint validation
	if len(tag) >= 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid room reference relay hint: %w", err)
		}
	}

	// Optional root marker validation
	if len(tag) == 4 && tag[3] != "" {
		if tag[3] != "root" {
			return fmt.Errorf("invalid marker: must be 'root' or empty")
		}
	}

	return nil
}

// validateLiveChatReplyTag validates the e tag for live chat message replies
func validateLiveChatReplyTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("e tag must have 2 or 3 elements")
	}

	eventID := tag[1]
	if len(eventID) != 64 {
		return fmt.Errorf("reply event ID must be 64 hex characters, got %d", len(eventID))
	}

	if !isHexChar64(eventID) {
		return fmt.Errorf("reply event ID must be valid hex")
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid reply event relay hint: %w", err)
		}
	}

	return nil
}

// validateQuoteTag validates the q tag for quoted events
func validateQuoteTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 4 {
		return fmt.Errorf("q tag must have 2-4 elements")
	}

	eventIDOrAddr := tag[1]
	if eventIDOrAddr == "" {
		return fmt.Errorf("quoted event ID or address cannot be empty")
	}

	// Could be event ID (64 hex chars) or event address (kind:pubkey:dtag) or event address with relay (kind:pubkey:dtag,relay)
	if len(eventIDOrAddr) == 64 {
		// Event ID
		if !isHexChar64(eventIDOrAddr) {
			return fmt.Errorf("quoted event ID must be valid hex")
		}
	} else {
		// Event address - handle format with optional comma-separated relay
		var addressPart, relayPart string
		if strings.Contains(eventIDOrAddr, ",") {
			parts := strings.SplitN(eventIDOrAddr, ",", 2)
			addressPart = parts[0]
			relayPart = parts[1]
		} else {
			addressPart = eventIDOrAddr
		}
		
		// Validate basic format: kind:pubkey:dtag
		addressParts := strings.Split(addressPart, ":")
		if len(addressParts) != 3 {
			return fmt.Errorf("quoted event address must be in format 'kind:pubkey:d_tag_value'")
		}
		
		// Validate kind is numeric
		if _, err := strconv.Atoi(addressParts[0]); err != nil {
			return fmt.Errorf("invalid kind in quoted event address: %s", addressParts[0])
		}
		
		// Validate pubkey format
		if len(addressParts[1]) != 64 || !isHexChar64(addressParts[1]) {
			return fmt.Errorf("invalid pubkey in quoted event address")
		}
		
		// Validate relay if provided in comma-separated format
		if relayPart != "" {
			if err := validateRelayURL(relayPart); err != nil {
				return fmt.Errorf("invalid relay in quoted event address: %w", err)
			}
		}
	}

	// Optional relay hint validation in separate element
	if len(tag) >= 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid quoted event relay hint: %w", err)
		}
	}

	// Optional pubkey validation (for regular events)
	if len(tag) == 4 && tag[3] != "" {
		pubkey := tag[3]
		if len(pubkey) != 64 || !isHexChar64(pubkey) {
			return fmt.Errorf("invalid pubkey in q tag")
		}
	}

	return nil
}

// validateRoomNameTag validates the room name tag for meeting spaces
func validateRoomNameTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("room tag must have exactly 2 elements")
	}

	roomName := tag[1]
	if roomName == "" {
		return fmt.Errorf("room name cannot be empty")
	}

	if len(roomName) > 200 {
		return fmt.Errorf("room name too long (max 200 characters)")
	}

	return nil
}

// validateServiceURLTag validates the service URL tag for meeting spaces
func validateServiceURLTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("service tag must have exactly 2 elements")
	}

	serviceURL := tag[1]
	if serviceURL == "" {
		return fmt.Errorf("service URL cannot be empty")
	}

	// Validate URL format
	u, err := url.Parse(serviceURL)
	if err != nil {
		return fmt.Errorf("invalid service URL format: %w", err)
	}

	// Allow http/https schemes for service URLs
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("service URL must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("service URL must have a host")
	}

	if len(serviceURL) > 2000 {
		return fmt.Errorf("service URL too long (max 2000 characters)")
	}

	return nil
}

// validateEndpointURLTag validates the endpoint URL tag for meeting spaces
func validateEndpointURLTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("endpoint tag must have exactly 2 elements")
	}

	endpointURL := tag[1]
	if endpointURL == "" {
		return fmt.Errorf("endpoint URL cannot be empty")
	}

	// Validate URL format
	u, err := url.Parse(endpointURL)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL format: %w", err)
	}

	// Allow http/https schemes for endpoint URLs
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("endpoint URL must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("endpoint URL must have a host")
	}

	if len(endpointURL) > 2000 {
		return fmt.Errorf("endpoint URL too long (max 2000 characters)")
	}

	return nil
}

// validateHandRaisedTag validates the hand raised tag for room presence
func validateHandRaisedTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("hand tag must have exactly 2 elements")
	}

	handValue := tag[1]
	if handValue != "0" && handValue != "1" {
		return fmt.Errorf("hand value must be '0' or '1'")
	}

	return nil
}

// validateLiveActivityReference validates the format of a live activity reference (a tag value)
func validateLiveActivityReference(aTagValue string) error {
	// Format: kind:pubkey:d_tag_value
	parts := strings.Split(aTagValue, ":")
	if len(parts) != 3 {
		return fmt.Errorf("live activity reference must be in format 'kind:pubkey:d_tag_value', got '%s'", aTagValue)
	}

	// Validate kind
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in live activity reference: %s", parts[0])
	}
	if kind != 30311 {
		return fmt.Errorf("live activity reference must reference kind 30311, got %d", kind)
	}

	// Validate pubkey
	pubkey := parts[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey in live activity reference must be 64 hex characters, got %d", len(pubkey))
	}
	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey in live activity reference must be valid hex")
	}

	// Validate d tag value
	dTagValue := parts[2]
	if dTagValue == "" {
		return fmt.Errorf("d tag value in live activity reference cannot be empty")
	}
	if len(dTagValue) > 200 {
		return fmt.Errorf("d tag value in live activity reference too long (max 200 characters)")
	}

	return nil
}

// validateMeetingSpaceReference validates the format of a meeting space reference (a tag value)
func validateMeetingSpaceReference(aTagValue string) error {
	// Format: kind:pubkey:d_tag_value
	parts := strings.Split(aTagValue, ":")
	if len(parts) != 3 {
		return fmt.Errorf("meeting space reference must be in format 'kind:pubkey:d_tag_value', got '%s'", aTagValue)
	}

	// Validate kind
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in meeting space reference: %s", parts[0])
	}
	if kind != 30312 {
		return fmt.Errorf("meeting space reference must reference kind 30312, got %d", kind)
	}

	// Validate pubkey
	pubkey := parts[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey in meeting space reference must be 64 hex characters, got %d", len(pubkey))
	}
	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey in meeting space reference must be valid hex")
	}

	// Validate d tag value
	dTagValue := parts[2]
	if dTagValue == "" {
		return fmt.Errorf("d tag value in meeting space reference cannot be empty")
	}
	if len(dTagValue) > 200 {
		return fmt.Errorf("d tag value in meeting space reference too long (max 200 characters)")
	}

	return nil
}

// validateRoomReference validates the format of a room reference for presence events
func validateRoomReference(aTagValue string) error {
	// Format: kind:pubkey:d_tag_value
	parts := strings.Split(aTagValue, ":")
	if len(parts) != 3 {
		return fmt.Errorf("room reference must be in format 'kind:pubkey:d_tag_value', got '%s'", aTagValue)
	}

	// Validate kind
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in room reference: %s", parts[0])
	}
	// Room presence can reference either live streaming (30311) or meeting spaces (30312)
	if kind != 30311 && kind != 30312 {
		return fmt.Errorf("room reference must reference kind 30311 or 30312, got %d", kind)
	}

	// Validate pubkey
	pubkey := parts[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey in room reference must be 64 hex characters, got %d", len(pubkey))
	}
	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey in room reference must be valid hex")
	}

	// Validate d tag value
	dTagValue := parts[2]
	if dTagValue == "" {
		return fmt.Errorf("d tag value in room reference cannot be empty")
	}
	if len(dTagValue) > 200 {
		return fmt.Errorf("d tag value in room reference too long (max 200 characters)")
	}

	return nil
}