package nips

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-72: Moderated Communities (Reddit-style Nostr Communities)
//
// Event Kinds:
//   - 34550: Community Definition
//   - 1111: Community Post  
//   - 4550: Community Post Approval Event

// ValidateCommunityDefinition validates a community definition event (kind 34550)
func ValidateCommunityDefinition(event *nostr.Event) error {
	// Validate event kind
	if event.Kind != 34550 {
		return fmt.Errorf("community definition event must be kind 34550")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid community definition structure: %w", err)
	}

	// Validate required tags
	var hasDTag bool
	var hasNameTag bool
	var hasModerators bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			hasDTag = true
			if err := validateCommunityIdentifier(tag[1]); err != nil {
				return fmt.Errorf("invalid d tag: %w", err)
			}
		case "name":
			hasNameTag = true
			if err := validateCommunityName(tag[1]); err != nil {
				return fmt.Errorf("invalid name tag: %w", err)
			}
		case "description":
			if err := validateCommunityDescription(tag[1]); err != nil {
				return fmt.Errorf("invalid description tag: %w", err)
			}
		case "image":
			if err := validateCommunityImageTag(tag); err != nil {
				return fmt.Errorf("invalid image tag: %w", err)
			}
		case "relay":
			if err := validateCommunityRelay(tag); err != nil {
				return fmt.Errorf("invalid relay tag: %w", err)
			}
		case "moderators", "p":
			hasModerators = true
			if len(tag) < 2 {
				return fmt.Errorf("moderator tag must have pubkey")
			}
			pubkey := tag[1]
			// Accept both 64-char uncompressed and 66-char compressed pubkeys
			if len(pubkey) == 64 {
				if !isHexString(pubkey) {
					return fmt.Errorf("moderator pubkey must be valid hex")
				}
			} else if len(pubkey) == 66 && (strings.HasPrefix(pubkey, "02") || strings.HasPrefix(pubkey, "03")) {
				if !isHexString(pubkey) {
					return fmt.Errorf("moderator pubkey must be valid hex")
				}
			} else {
				return fmt.Errorf("moderator pubkey must be 64 characters (uncompressed) or 66 characters with 02/03 prefix (compressed), got %d", len(pubkey))
			}
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("community definition must have d tag")
	}
	if !hasNameTag {
		return fmt.Errorf("community definition must have name tag")
	}
	if !hasModerators {
		return fmt.Errorf("community definition must have at least one moderator (p or moderators tag)")
	}

	return nil
}

// ValidateCommunityPost validates a community post event (kind 1111)
func ValidateCommunityPost(event *nostr.Event) error {
	// Validate event kind
	if event.Kind != 1111 {
		return fmt.Errorf("community post event must be kind 1111")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid community post structure: %w", err)
	}

	// Must have community reference
	var hasCommunityA bool
	var hasCommunityK bool
	var hasLowercaseA bool
	var hasLowercaseK bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "A":
			// Uppercase A tag for community reference
			if len(tag) >= 2 && strings.HasPrefix(tag[1], "34550:") {
				hasCommunityA = true
				if err := validateCommunityReference(tag[1], "34550"); err != nil {
					return fmt.Errorf("invalid community A tag: %w", err)
				}
			}
		case "a":
			// Lowercase a tag for addressable event reference
			if len(tag) >= 2 && strings.HasPrefix(tag[1], "34550:") {
				hasLowercaseA = true
				if err := validateCommunityReference(tag[1], "34550"); err != nil {
					return fmt.Errorf("invalid community a tag: %w", err)
				}
			}
		case "K":
			// Uppercase K tag for kind
			hasCommunityK = true
			if len(tag) >= 2 && tag[1] != "34550" && tag[1] != "1111" {
				return fmt.Errorf("k tag must reference kind 34550 (community) or 1111 (community post)")
			}
		case "k":
			// Lowercase k tag for kind
			hasLowercaseK = true
			if len(tag) >= 2 && tag[1] != "34550" && tag[1] != "1111" {
				return fmt.Errorf("k tag must reference kind 34550 (community) or 1111 (community post)")
			}
		case "P":
			// Uppercase P tag for pubkey (allow both 64-char uncompressed and 66-char compressed with 02/03 prefix)
			if len(tag) >= 2 {
				pubkey := tag[1]
				if len(pubkey) != 64 && (len(pubkey) != 66 || (!strings.HasPrefix(pubkey, "02") && !strings.HasPrefix(pubkey, "03"))) {
					return fmt.Errorf("p tag must contain valid pubkey (64 hex chars or 66 chars with 02/03 prefix)")
				}
			}
		case "p":
			// Lowercase p tag for pubkey (allow both 64-char uncompressed and 66-char compressed with 02/03 prefix)
			if len(tag) >= 2 {
				pubkey := tag[1]
				if len(pubkey) != 64 && (len(pubkey) != 66 || (!strings.HasPrefix(pubkey, "02") && !strings.HasPrefix(pubkey, "03"))) {
					return fmt.Errorf("p tag must contain valid pubkey (64 hex chars or 66 chars with 02/03 prefix)")
				}
			}
		case "e":
			// Event reference for replies
			if len(tag) >= 2 && len(tag[1]) != 64 {
				return fmt.Errorf("e tag must contain 64-character event ID")
			}
		}
	}

	// Determine if this is a top-level community post or a reply
	isReply := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			isReply = true
			break
		}
	}

	// Validate required tags based on whether this is a top-level post or reply
	if isReply {
		// For replies, require lowercase tags
		if !hasLowercaseA {
			return fmt.Errorf("community reply must have lowercase a tag for NIP-72 compliance")
		}
		if !hasLowercaseK {
			return fmt.Errorf("community reply must have lowercase k tag for NIP-72 compliance")
		}
		// Validate that k tag has the correct value for replies (should be kind of event being replied to)
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "k" {
				if tag[1] != "1111" && tag[1] != "34550" {
					return fmt.Errorf("community reply k tag must be 1111 or 34550, got: %s", tag[1])
				}
			}
		}
	} else {
		// For top-level posts, require uppercase tags
		if !hasCommunityA {
			return fmt.Errorf("community post must have uppercase A tag referencing a community")
		}
		if !hasCommunityK {
			return fmt.Errorf("community post must have uppercase K tag with kind 34550")
		}
		// Validate that K tag has the correct value for top-level posts (should be 34550)
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "K" {
				if tag[1] != "34550" {
					return fmt.Errorf("community post K tag must be 34550, got: %s", tag[1])
				}
			}
		}
	}

	return nil
}

// ValidateApprovalEvent validates a community post approval event (kind 4550)
func ValidateApprovalEvent(event *nostr.Event) error {
	// Validate event kind
	if event.Kind != 4550 {
		return fmt.Errorf("approval event must be kind 4550")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid approval event structure: %w", err)
	}

	// Required tags validation
	var hasATag bool
	var hasETag bool
	var hasPTag bool
	var hasKTag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "a":
			hasATag = true
			if len(tag) >= 2 && strings.HasPrefix(tag[1], "34550:") {
				if err := validateCommunityReference(tag[1], "34550"); err != nil {
					return fmt.Errorf("invalid community a tag in approval: %w", err)
				}
			}
		case "e":
			hasETag = true
			if len(tag) >= 2 && len(tag[1]) != 64 {
				return fmt.Errorf("e tag must contain 64-character event ID")
			}
		case "p":
			hasPTag = true
			if len(tag) >= 2 {
				pubkey := tag[1]
				if len(pubkey) != 64 && (len(pubkey) != 66 || (!strings.HasPrefix(pubkey, "02") && !strings.HasPrefix(pubkey, "03"))) {
					return fmt.Errorf("p tag must contain valid pubkey (64 hex chars or 66 chars with 02/03 prefix)")
				}
			}
		case "k":
			hasKTag = true
			if len(tag) >= 2 {
				// Allow approval of different kinds, not just 1111
				validKinds := map[string]bool{
					"1111":  true, // Community posts
					"30023": true, // Long-form content
					"1":     true, // Regular notes in communities
				}
				if !validKinds[tag[1]] {
					return fmt.Errorf("k tag contains unsupported kind for approval: %s", tag[1])
				}
			}
		}
	}

	// All required tags must be present
	if !hasATag {
		return fmt.Errorf("approval event must have a tag referencing community")
	}
	// For replaceable events, we can use 'a' tags instead of 'e' tags
	// Check if we have either e tag or additional a tags (for replaceable events)
	hasEventReference := hasETag
	if !hasEventReference {
		// Check if we have additional 'a' tags that could be referencing replaceable events
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "a" && !strings.HasPrefix(tag[1], "34550:") {
				hasEventReference = true
				break
			}
		}
	}
	if !hasEventReference {
		return fmt.Errorf("approval event must have e tag (for regular events) or a tag (for replaceable events) referencing approved event")
	}
	if !hasPTag {
		return fmt.Errorf("approval event must have p tag referencing event author")
	}
	if !hasKTag {
		return fmt.Errorf("approval event must have k tag with appropriate kind")
	}

	// Content should contain the original event
	if event.Content == "" {
		return fmt.Errorf("approval event content cannot be empty")
	}

	// Validate that content is valid JSON (should contain the original event)
	var originalEvent interface{}
	if err := json.Unmarshal([]byte(event.Content), &originalEvent); err != nil {
		return fmt.Errorf("approval event content must be valid JSON: %w", err)
	}

	return nil
}

// Helper validation functions

func validateCommunityIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("community identifier cannot be empty")
	}
	if len(identifier) > 32 {
		return fmt.Errorf("community identifier too long (max 32 characters)")
	}
	return nil
}

func validateCommunityName(name string) error {
	if name == "" {
		return fmt.Errorf("community name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("community name too long (max 100 characters)")
	}
	return nil
}

func validateCommunityDescription(description string) error {
	if len(description) > 500 {
		return fmt.Errorf("community description too long (max 500 characters)")
	}
	return nil
}

// Helper function to validate community image tag (uses existing nip58 functions)
func validateCommunityImageTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("image tag must have URL")
	}

	// Validate image URL using existing function from nip58.go
	imageURL := tag[1]
	if err := validateImageURL(imageURL); err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}

	// Validate optional dimensions using existing function from nip58.go
	if len(tag) >= 3 {
		dimensions := tag[2]
		if err := validateImageDimensions(dimensions); err != nil {
			return fmt.Errorf("invalid image dimensions: %w", err)
		}
	}

	return nil
}

// Helper function to validate community relay tags
func validateCommunityRelay(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("relay tag must have URL")
	}

	// Validate relay URL
	relayURL := tag[1]
	if err := validateRelayURL(relayURL); err != nil {
		return fmt.Errorf("invalid relay URL: %w", err)
	}

	// Validate optional relay marker
	if len(tag) >= 3 {
		marker := tag[2]
		validMarkers := map[string]bool{
			"author":    true,
			"requests":  true,
			"approvals": true,
		}

		if marker != "" && !validMarkers[marker] {
			return fmt.Errorf("invalid relay marker: %s (valid: author, requests, approvals)", marker)
		}
	}

	return nil
}

// Helper function to validate community references (a/A tags)
func validateCommunityReference(reference string, expectedKind string) error {
	if reference == "" {
		return fmt.Errorf("community reference cannot be empty")
	}

	// Format: kind:pubkey:d-identifier
	parts := strings.Split(reference, ":")
	if len(parts) != 3 {
		return fmt.Errorf("community reference must be in format kind:pubkey:d-identifier")
	}

	kind, pubkey, dIdentifier := parts[0], parts[1], parts[2]

	// Validate kind
	if kind != expectedKind {
		return fmt.Errorf("expected kind %s in community reference, got: %s", expectedKind, kind)
	}

	// Validate pubkey - accept both uncompressed (64) and compressed (66) formats
	if len(pubkey) == 64 {
		if !isHexString(pubkey) {
			return fmt.Errorf("invalid pubkey in community reference: %s", pubkey)
		}
	} else if len(pubkey) == 66 && (strings.HasPrefix(pubkey, "02") || strings.HasPrefix(pubkey, "03")) {
		if !isHexString(pubkey) {
			return fmt.Errorf("invalid pubkey in community reference: %s", pubkey)
		}
	} else {
		return fmt.Errorf("invalid pubkey format in community reference (must be 64 or 66 chars): %s", pubkey)
	}

	// Validate d-identifier
	if err := validateCommunityIdentifier(dIdentifier); err != nil {
		return fmt.Errorf("invalid d-identifier in community reference: %w", err)
	}

	return nil
}

// Helper function to check if string is numeric

// Additional validation helpers for cross-posting and backwards compatibility

// ValidateCrossPost validates cross-posting scenarios (NIP-18 kind 6 or kind 16)
func ValidateCrossPost(event *nostr.Event) error {
	// Validate event kind for cross-posting
	if event.Kind != 6 && event.Kind != 16 {
		return fmt.Errorf("cross-post event must be kind 6 or 16")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid cross-post event structure: %w", err)
	}

	// Check for community references
	var hasCommunityA bool
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			hasCommunityA = true
			if err := validateCommunityReference(tag[1], "34550"); err != nil {
				return fmt.Errorf("invalid community reference in cross-post: %w", err)
			}
		}
	}

	if !hasCommunityA {
		return fmt.Errorf("cross-post must reference at least one community with a tag")
	}

	// Content must be the original event, not the approval event
	if event.Content == "" {
		return fmt.Errorf("cross-post content cannot be empty - must contain original event")
	}

	// Validate that content is valid JSON
	var originalEvent interface{}
	if err := json.Unmarshal([]byte(event.Content), &originalEvent); err != nil {
		return fmt.Errorf("cross-post content must be valid JSON: %w", err)
	}

	return nil
}

// ValidateBackwardsCompatibilityPost validates legacy kind 1 posts to communities
func ValidateBackwardsCompatibilityPost(event *nostr.Event) error {
	// This validates kind 1 events with community "a" tags for backwards compatibility
	if event.Kind != 1 {
		return fmt.Errorf("backwards compatibility post must be kind 1")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid backwards compatibility post structure: %w", err)
	}

	// Check for community reference
	var hasCommunityA bool
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "a" && strings.HasPrefix(tag[1], "34550:") {
			hasCommunityA = true
			if err := validateCommunityReference(tag[1], "34550"); err != nil {
				return fmt.Errorf("invalid community reference in backwards compatibility post: %w", err)
			}
		}
	}

	if !hasCommunityA {
		return fmt.Errorf("backwards compatibility post must reference a community with a tag")
	}

	return nil
}