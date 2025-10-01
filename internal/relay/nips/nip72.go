package nips

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shugur-Network/relay/internal/relay/nips/common"
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
	return common.ValidateEventWithCallback(
		event,
		"72",                   // NIP number
		34550,                  // Expected event kind
		"community definition", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Validate basic community definition structure
			return validateCommunityDefinitionTags(helper, evt)
		},
	)
}

// ValidateCommunityPost validates a community post event (kind 1111)
func ValidateCommunityPost(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"72",             // NIP number
		1111,             // Expected event kind
		"community post", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Validate basic community post structure
			return validateCommunityPostTags(helper, evt)
		},
	)
}

// ValidateApprovalEvent validates a community post approval event (kind 4550)
func ValidateApprovalEvent(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"72",             // NIP number
		4550,             // Expected event kind
		"approval event", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Validate basic approval event structure
			return validateApprovalEventTags(helper, evt)
		},
	)
}

// ValidateCrossPost validates cross-post events
func ValidateCrossPost(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"72",            // NIP number
		int(event.Kind), // Use actual event kind
		"cross post",    // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Basic cross-post validation
			return nil
		},
	)
}

// ValidateBackwardsCompatibilityPost validates backwards compatibility posts
func ValidateBackwardsCompatibilityPost(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"72",                           // NIP number
		int(event.Kind),                // Use actual event kind
		"backwards compatibility post", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Basic backwards compatibility validation
			return nil
		},
	)
}

// Helper functions for NIP-72 validation

func validateCommunityDefinitionTags(helper *common.ValidationHelper, event *nostr.Event) error {
	// Validate required tags for community definition
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
			if err := validateCommunityIdentifier72(tag[1]); err != nil {
				return helper.FormatTagError("d", "invalid community identifier: %v", err)
			}
		case "name":
			hasNameTag = true
			if err := validateCommunityName72(tag[1]); err != nil {
				return helper.FormatTagError("name", "invalid community name: %v", err)
			}
		case "description":
			if err := validateCommunityDescription72(tag[1]); err != nil {
				return helper.FormatTagError("description", "invalid community description: %v", err)
			}
		case "image":
			if err := validateCommunityImageTag72(tag); err != nil {
				return helper.FormatTagError("image", "invalid community image: %v", err)
			}
		case "relay":
			if err := validateCommunityRelay72(tag); err != nil {
				return helper.FormatTagError("relay", "invalid community relay: %v", err)
			}
		case "moderators", "p":
			hasModerators = true
			if len(tag) < 2 {
				return helper.FormatTagError(tag[0], "moderator tag must have pubkey")
			}
			pubkey := tag[1]
			// Accept both 64-char uncompressed and 66-char compressed pubkeys
			if len(pubkey) == 64 {
				if !isHexString72(pubkey) {
					return helper.FormatTagError(tag[0], "moderator pubkey must be valid hex")
				}
			} else if len(pubkey) == 66 && (strings.HasPrefix(pubkey, "02") || strings.HasPrefix(pubkey, "03")) {
				if !isHexString72(pubkey) {
					return helper.FormatTagError(tag[0], "moderator pubkey must be valid hex")
				}
			} else {
				return helper.FormatTagError(tag[0], "moderator pubkey must be 64 characters (uncompressed) or 66 characters with 02/03 prefix (compressed), got %d", len(pubkey))
			}
		}
	}

	// Required tags validation
	if !hasDTag {
		return helper.ErrorFormatter.FormatError("community definition must have d tag")
	}
	if !hasNameTag {
		return helper.ErrorFormatter.FormatError("community definition must have name tag")
	}
	if !hasModerators {
		return helper.ErrorFormatter.FormatError("community definition must have at least one moderator (p or moderators tag)")
	}

	return nil
}

func validateCommunityPostTags(helper *common.ValidationHelper, event *nostr.Event) error {
	// Validate community post tags
	var hasATag bool
	var kTagValues []string

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "A", "a":
			hasATag = true
			if len(tag) < 2 {
				return helper.FormatTagError(tag[0], "community reference tag must have value")
			}
			// Validate community reference format (34550:pubkey:identifier)
			if err := validateCommunityReference72(tag[1], "34550"); err != nil {
				return helper.FormatTagError(tag[0], "invalid community reference: %v", err)
			}
		case "K", "k":
			if len(tag) >= 2 {
				// K tag can reference the community kind (34550) or other referenced event kinds
				// According to NIP-72, this is flexible and depends on what's being referenced
				// Allow any numeric value
				if _, err := strconv.Atoi(tag[1]); err != nil {
					return helper.FormatTagError(tag[0], "K tag value must be numeric, got '%s'", tag[1])
				}
				kTagValues = append(kTagValues, tag[1])
			}
		}
	}

	if !hasATag {
		return helper.ErrorFormatter.FormatError("community post must have community reference (A or a tag)")
	}

	// If both K and k tags are present, they should have the same value
	if len(kTagValues) > 1 {
		firstValue := kTagValues[0]
		for _, value := range kTagValues[1:] {
			if value != firstValue {
				return helper.ErrorFormatter.FormatError("conflicting K tag values: all K/k tags must have the same value")
			}
		}
	}

	return nil
}

func validateApprovalEventTags(helper *common.ValidationHelper, event *nostr.Event) error {
	// Approval event validation
	var hasATag bool
	var hasPostReference bool
	var hasAuthorReference bool
	var hasKindReference bool
	var hasReplaceableReference bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "a":
			if len(tag) < 2 {
				return helper.FormatTagError("a", "community reference tag must have value")
			}
			// Check if this is a community reference (34550:...)
			if strings.HasPrefix(tag[1], "34550:") {
				hasATag = true
				// Validate community reference format (34550:pubkey:identifier)
				if err := validateCommunityReference72(tag[1], "34550"); err != nil {
					return helper.FormatTagError("a", "invalid community reference: %v", err)
				}
			} else {
				// This could be a reference to a replaceable event being approved
				hasReplaceableReference = true
			}
		case "e":
			hasPostReference = true
			if len(tag) < 2 || len(tag[1]) != 64 || !isHexString72(tag[1]) {
				return helper.FormatTagError("e", "invalid event ID format")
			}
		case "p":
			hasAuthorReference = true
			if len(tag) < 2 {
				return helper.FormatTagError("p", "pubkey tag must have value")
			}
			pubkey := tag[1]
			// Accept both 64-char uncompressed and 66-char compressed pubkeys
			if len(pubkey) == 64 {
				if !isHexString72(pubkey) {
					return helper.FormatTagError("p", "pubkey must be valid hex")
				}
			} else if len(pubkey) == 66 && (strings.HasPrefix(pubkey, "02") || strings.HasPrefix(pubkey, "03")) {
				if !isHexString72(pubkey) {
					return helper.FormatTagError("p", "pubkey must be valid hex")
				}
			} else {
				return helper.FormatTagError("p", "pubkey must be 64 characters (uncompressed) or 66 characters with 02/03 prefix (compressed), got %d", len(pubkey))
			}
		case "k":
			hasKindReference = true
			if len(tag) < 2 {
				return helper.FormatTagError("k", "kind tag must have value")
			}
		}
	}

	// Required tags validation
	if !hasATag {
		return helper.ErrorFormatter.FormatError("approval event must have community reference (a tag)")
	}
	// Either an "e" tag (for regular events) or an "a" tag (for replaceable events) is required
	if !hasPostReference && !hasReplaceableReference {
		return helper.ErrorFormatter.FormatError("approval event must reference a post (e tag) or replaceable event (a tag)")
	}
	if !hasAuthorReference {
		return helper.ErrorFormatter.FormatError("approval event must reference post author (p tag)")
	}
	if !hasKindReference {
		return helper.ErrorFormatter.FormatError("approval event must specify post kind (k tag)")
	}

	// Validate content is valid JSON
	if event.Content == "" {
		return helper.ErrorFormatter.FormatError("approval event content cannot be empty")
	}

	// Try to parse as JSON
	if !isValidJSON(event.Content) {
		return helper.ErrorFormatter.FormatError("approval event content must be valid JSON")
	}

	return nil
}

// Basic helper functions for NIP-72 (renamed to avoid conflicts)
func validateCommunityIdentifier72(id string) error {
	if id == "" {
		return fmt.Errorf("community identifier cannot be empty")
	}
	// Validate length (reasonable limit for community identifiers)
	if len(id) > 64 {
		return fmt.Errorf("community identifier too long (max 64 characters), got %d", len(id))
	}
	// Allow alphanumeric, hyphens, and underscores
	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return fmt.Errorf("community identifier contains invalid character: %c", r)
		}
	}
	return nil
}

func validateCommunityName72(name string) error {
	if name == "" {
		return fmt.Errorf("community name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("community name too long (max 100 characters), got %d", len(name))
	}
	return nil
}

func validateCommunityDescription72(desc string) error {
	if len(desc) > 500 {
		return fmt.Errorf("community description too long (max 500 characters), got %d", len(desc))
	}
	return nil
}

func validateCommunityImageTag72(tag []string) error {
	if len(tag) < 2 {
		return fmt.Errorf("image tag must have URL")
	}
	
	// Basic URL validation
	if !strings.HasPrefix(tag[1], "http://") && !strings.HasPrefix(tag[1], "https://") {
		return fmt.Errorf("image URL must start with http:// or https://")
	}
	
	// Check for dimensions parameter if present
	if len(tag) >= 3 && tag[2] != "" {
		// Validate dimensions format (e.g., "200x200")
		if !isValidDimensions(tag[2]) {
			return fmt.Errorf("invalid dimensions format: %s (expected format: WxH like 200x200)", tag[2])
		}
	}
	
	return nil
}

func validateCommunityRelay72(tag []string) error {
	if len(tag) < 2 {
		return fmt.Errorf("relay tag must have URL")
	}
	
	// Basic relay URL validation
	if !strings.HasPrefix(tag[1], "ws://") && !strings.HasPrefix(tag[1], "wss://") {
		return fmt.Errorf("relay URL must start with ws:// or wss://")
	}
	
	// Check for marker parameter if present
	if len(tag) >= 3 && tag[2] != "" {
		// Define allowed markers for NIP-72
		allowedMarkers := map[string]bool{
			"read":      true,
			"write":     true,
			"author":    true,
			"requests":  true,
			"approvals": true,
		}
		
		if !allowedMarkers[tag[2]] {
			return fmt.Errorf("invalid relay marker: %s (allowed: read, write, author, requests, approvals)", tag[2])
		}
	}
	
	return nil
}

func validateCommunityReference72(ref string, expectedKind string) error {
	if ref == "" {
		return fmt.Errorf("community reference cannot be empty")
	}
	
	// Format should be: kind:pubkey:identifier
	parts := strings.Split(ref, ":")
	if len(parts) != 3 {
		return fmt.Errorf("community reference must have format 'kind:pubkey:identifier', got %d parts", len(parts))
	}
	
	// Validate kind
	if parts[0] != expectedKind {
		return fmt.Errorf("expected kind %s, got %s", expectedKind, parts[0])
	}
	
	// Validate pubkey format
	pubkey := parts[1]
	if len(pubkey) == 64 {
		if !isHexString72(pubkey) {
			return fmt.Errorf("invalid pubkey format in community reference")
		}
	} else if len(pubkey) == 66 && (strings.HasPrefix(pubkey, "02") || strings.HasPrefix(pubkey, "03")) {
		if !isHexString72(pubkey) {
			return fmt.Errorf("invalid pubkey format in community reference")
		}
	} else {
		return fmt.Errorf("invalid pubkey format in community reference: must be 64 or 66 characters, got %d", len(pubkey))
	}
	
	// Validate identifier
	if err := validateCommunityIdentifier72(parts[2]); err != nil {
		return fmt.Errorf("invalid community identifier in reference: %w", err)
	}
	
	return nil
}

func isHexString72(s string) bool {
	// Simple hex string validation
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// isValidDimensions validates dimension format like "200x200"
func isValidDimensions(dims string) bool {
	parts := strings.Split(dims, "x")
	if len(parts) != 2 {
		return false
	}
	
	// Both parts should be positive integers
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	
	return true
}

// isValidJSON checks if content is valid JSON
func isValidJSON(content string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(content), &js) == nil
}
