package nips

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	nostr "github.com/nbd-wtf/go-nostr"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// ValidateBadgeDefinition validates NIP-58 badge definition events (kind 30009)
func ValidateBadgeDefinition(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-58: Validating badge definition event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 30009 {
		return fmt.Errorf("invalid kind for badge definition: expected 30009, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateBadgeDefinitionTags(event); err != nil {
		return fmt.Errorf("invalid badge definition tags: %w", err)
	}

	logger.Debug("NIP-58: Badge definition validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateBadgeAward validates NIP-58 badge award events (kind 8)
func ValidateBadgeAward(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-58: Validating badge award event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 8 {
		return fmt.Errorf("invalid kind for badge award: expected 8, got %d", event.Kind)
	}

	// Badge awards SHOULD have empty content
	if event.Content != "" {
		logger.Warn("NIP-58: Badge award content should be empty",
			zap.String("event_id", event.ID),
			zap.String("content", event.Content))
	}

	// Validate required tags
	if err := validateBadgeAwardTags(event); err != nil {
		return fmt.Errorf("invalid badge award tags: %w", err)
	}

	logger.Debug("NIP-58: Badge award validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateProfileBadges validates NIP-58 profile badges events (kind 30008)
func ValidateProfileBadges(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-58: Validating profile badges event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 30008 {
		return fmt.Errorf("invalid kind for profile badges: expected 30008, got %d", event.Kind)
	}

	// Profile badges SHOULD have empty content
	if event.Content != "" {
		logger.Warn("NIP-58: Profile badges content should be empty",
			zap.String("event_id", event.ID),
			zap.String("content", event.Content))
	}

	// Validate required tags
	if err := validateProfileBadgesTags(event); err != nil {
		return fmt.Errorf("invalid profile badges tags: %w", err)
	}

	logger.Debug("NIP-58: Profile badges validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// validateBadgeDefinitionTags validates tags for badge definition events
func validateBadgeDefinitionTags(event *nostr.Event) error {
	var hasDTag bool
	var hasNameTag bool
	var dTagValue string

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateBadgeDTag(tag); err != nil {
				return err
			}
			hasDTag = true
			dTagValue = tag[1]
		case "name":
			if err := validateBadgeNameTag(tag); err != nil {
				return err
			}
			hasNameTag = true
		case "image":
			if err := validateBadgeImageTag(tag); err != nil {
				return err
			}
		case "description":
			if err := validateBadgeDescriptionTag(tag); err != nil {
				return err
			}
		case "thumb":
			if err := validateBadgeThumbTag(tag); err != nil {
				return err
			}
		case "dim":
			if err := validateBadgeDimTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("badge definition must have a d tag")
	}

	if !hasNameTag {
		return fmt.Errorf("badge definition must have a name tag")
	}

	// Validate d tag uniqueness requirements
	if dTagValue == "" {
		return fmt.Errorf("d tag value cannot be empty")
	}

	if len(dTagValue) > 100 {
		return fmt.Errorf("d tag value too long (max 100 characters)")
	}

	return nil
}

// validateBadgeAwardTags validates tags for badge award events
func validateBadgeAwardTags(event *nostr.Event) error {
	var hasATag bool
	var hasPTag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "a":
			if err := validateBadgeAwardATag(tag); err != nil {
				return err
			}
			hasATag = true
		case "p":
			if err := validateBadgeAwardPTag(tag); err != nil {
				return err
			}
			hasPTag = true
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasATag {
		return fmt.Errorf("badge award must have an a tag referencing a badge definition")
	}

	if !hasPTag {
		return fmt.Errorf("badge award must have at least one p tag")
	}

	return nil
}

// validateProfileBadgesTags validates tags for profile badges events
func validateProfileBadgesTags(event *nostr.Event) error {
	var hasDTag bool
	var aTags []nostr.Tag
	var eTags []nostr.Tag

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateProfileBadgesDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "a":
			if err := validateProfileBadgesATag(tag); err != nil {
				return err
			}
			aTags = append(aTags, tag)
		case "e":
			if err := validateProfileBadgesETag(tag); err != nil {
				return err
			}
			eTags = append(eTags, tag)
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("profile badges must have a d tag with value 'profile_badges'")
	}

	// Validate a/e tag pairs
	if err := validateBadgePairs(aTags, eTags); err != nil {
		return err
	}

	return nil
}

// validateBadgeDTag validates the d tag for badge definitions
func validateBadgeDTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("d tag must have exactly 2 elements")
	}

	badgeID := tag[1]
	if badgeID == "" {
		return fmt.Errorf("badge ID cannot be empty")
	}

	if len(badgeID) > 100 {
		return fmt.Errorf("badge ID too long (max 100 characters)")
	}

	// Badge ID should contain only alphanumeric characters, hyphens, and underscores
	validBadgeID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validBadgeID.MatchString(badgeID) {
		return fmt.Errorf("badge ID contains invalid characters (only alphanumeric, hyphens, and underscores allowed)")
	}

	return nil
}

// validateBadgeNameTag validates the name tag for badge definitions
func validateBadgeNameTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("name tag must have exactly 2 elements")
	}

	name := tag[1]
	if len(name) > 200 {
		return fmt.Errorf("badge name too long (max 200 characters)")
	}

	return nil
}

// validateBadgeImageTag validates the image tag for badge definitions
func validateBadgeImageTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("image tag must have 2 or 3 elements")
	}

	imageURL := tag[1]
	if err := validateImageURL(imageURL); err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}

	// Validate dimensions if provided
	if len(tag) == 3 {
		if err := validateImageDimensions(tag[2]); err != nil {
			return fmt.Errorf("invalid image dimensions: %w", err)
		}
	}

	return nil
}

// validateBadgeDescriptionTag validates the description tag for badge definitions
func validateBadgeDescriptionTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("description tag must have exactly 2 elements")
	}

	description := tag[1]
	if len(description) > 1000 {
		return fmt.Errorf("badge description too long (max 1000 characters)")
	}

	return nil
}

// validateBadgeThumbTag validates the thumb tag for badge definitions
func validateBadgeThumbTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("thumb tag must have 2 or 3 elements")
	}

	thumbURL := tag[1]
	if err := validateImageURL(thumbURL); err != nil {
		return fmt.Errorf("invalid thumbnail URL: %w", err)
	}

	// Validate dimensions if provided
	if len(tag) == 3 {
		if err := validateImageDimensions(tag[2]); err != nil {
			return fmt.Errorf("invalid thumbnail dimensions: %w", err)
		}
	}

	return nil
}

// validateBadgeDimTag validates the dim tag for badge definitions
func validateBadgeDimTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("dim tag must have exactly 2 elements")
	}

	if err := validateImageDimensions(tag[1]); err != nil {
		return fmt.Errorf("invalid badge dimensions: %w", err)
	}

	return nil
}

// validateBadgeAwardATag validates the a tag for badge award events
func validateBadgeAwardATag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("a tag must have exactly 2 elements")
	}

	aTagValue := tag[1]
	if err := validateBadgeDefinitionReference(aTagValue); err != nil {
		return fmt.Errorf("invalid badge definition reference: %w", err)
	}

	return nil
}

// validateBadgeAwardPTag validates the p tag for badge award events
func validateBadgeAwardPTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("p tag must have 2 or 3 elements")
	}

	pubkey := tag[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("awarded pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	// Validate hex format
	if !isHexChar64(pubkey) {
		return fmt.Errorf("awarded pubkey must be valid hex")
	}

	// Optional relay hint validation
	if len(tag) == 3 {
		relayHint := tag[2]
		if relayHint != "" {
			if err := validateBadgeRelayURL(relayHint); err != nil {
				return fmt.Errorf("invalid relay hint: %w", err)
			}
		}
	}

	return nil
}

// validateProfileBadgesDTag validates the d tag for profile badges events
func validateProfileBadgesDTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("d tag must have exactly 2 elements")
	}

	if tag[1] != "profile_badges" {
		return fmt.Errorf("profile badges d tag must have value 'profile_badges', got '%s'", tag[1])
	}

	return nil
}

// validateProfileBadgesATag validates the a tag for profile badges events
func validateProfileBadgesATag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("a tag must have exactly 2 elements")
	}

	aTagValue := tag[1]
	if err := validateBadgeDefinitionReference(aTagValue); err != nil {
		return fmt.Errorf("invalid badge definition reference: %w", err)
	}

	return nil
}

// validateProfileBadgesETag validates the e tag for profile badges events
func validateProfileBadgesETag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("e tag must have 2 or 3 elements")
	}

	eventID := tag[1]
	if len(eventID) != 64 {
		return fmt.Errorf("badge award event ID must be 64 hex characters, got %d", len(eventID))
	}

	// Validate hex format
	if !isHexChar64(eventID) {
		return fmt.Errorf("badge award event ID must be valid hex")
	}

	// Optional relay hint validation
	if len(tag) == 3 {
		relayHint := tag[2]
		if relayHint != "" {
			if err := validateBadgeRelayURL(relayHint); err != nil {
				return fmt.Errorf("invalid relay hint: %w", err)
			}
		}
	}

	return nil
}

// validateBadgePairs ensures a and e tags are properly paired in profile badges
func validateBadgePairs(aTags, eTags []nostr.Tag) error {
	// Profile badges can have zero badges
	if len(aTags) == 0 && len(eTags) == 0 {
		return nil
	}

	// a and e tags must be paired
	if len(aTags) != len(eTags) {
		return fmt.Errorf("a tags and e tags must be paired: found %d a tags and %d e tags", len(aTags), len(eTags))
	}

	// Additional validation could check if the referenced badge award actually contains the same a tag
	// but this would require database access which we don't have in basic validation

	return nil
}

// validateBadgeDefinitionReference validates the format of a badge definition reference (a tag value)
func validateBadgeDefinitionReference(aTagValue string) error {
	// Format: kind:pubkey:d_tag_value
	parts := strings.Split(aTagValue, ":")
	if len(parts) != 3 {
		return fmt.Errorf("badge definition reference must be in format 'kind:pubkey:d_tag_value', got '%s'", aTagValue)
	}

	// Validate kind
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in badge reference: %s", parts[0])
	}
	if kind != 30009 {
		return fmt.Errorf("badge definition reference must reference kind 30009, got %d", kind)
	}

	// Validate pubkey
	pubkey := parts[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey in badge reference must be 64 hex characters, got %d", len(pubkey))
	}
	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey in badge reference must be valid hex")
	}

	// Validate d tag value
	dTagValue := parts[2]
	if dTagValue == "" {
		return fmt.Errorf("d tag value in badge reference cannot be empty")
	}
	if len(dTagValue) > 100 {
		return fmt.Errorf("d tag value in badge reference too long (max 100 characters)")
	}

	return nil
}

// validateImageURL validates that a URL is properly formatted and uses appropriate schemes
func validateImageURL(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image URL cannot be empty")
	}

	u, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Only allow http/https schemes for security
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("image URL must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("image URL must have a host")
	}

	if len(imageURL) > 2000 {
		return fmt.Errorf("image URL too long (max 2000 characters)")
	}

	return nil
}

// validateImageDimensions validates the format of image dimensions (widthxheight)
func validateImageDimensions(dimensions string) error {
	if dimensions == "" {
		return fmt.Errorf("dimensions cannot be empty when specified")
	}

	parts := strings.Split(dimensions, "x")
	if len(parts) != 2 {
		return fmt.Errorf("dimensions must be in format 'widthxheight', got '%s'", dimensions)
	}

	// Validate width
	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid width in dimensions: %s", parts[0])
	}
	if width <= 0 || width > 10000 {
		return fmt.Errorf("width must be between 1 and 10000 pixels, got %d", width)
	}

	// Validate height
	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid height in dimensions: %s", parts[1])
	}
	if height <= 0 || height > 10000 {
		return fmt.Errorf("height must be between 1 and 10000 pixels, got %d", height)
	}

	return nil
}

// isHexChar64 validates a 64-character hex string
func isHexChar64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, char := range s {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}

// validateBadgeRelayURL validates relay URL format for badges
func validateBadgeRelayURL(relayURL string) error {
	if relayURL == "" {
		return fmt.Errorf("relay URL cannot be empty")
	}

	u, err := url.Parse(relayURL)
	if err != nil {
		return fmt.Errorf("invalid relay URL format: %w", err)
	}

	// WebSocket schemes for relay URLs
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("relay URL must use ws or wss scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("relay URL must have a host")
	}

	if len(relayURL) > 1000 {
		return fmt.Errorf("relay URL too long (max 1000 characters)")
	}

	return nil
}