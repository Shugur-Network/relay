package nips

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Shugur-Network/relay/internal/relay/nips/common"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateAsset validates NIP-YY Asset events (kind 1125)
// All web assets (HTML, CSS, JavaScript, fonts, etc.) use kind 1125
func ValidateAsset(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",  // NIP number
		1125,         // Expected event kind
		"web asset",  // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			return validateAssetTags(helper, evt)
		},
	)
}

// ValidatePageManifest validates NIP-YY Page Manifest events (kind 1126)
func ValidatePageManifest(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",     // NIP number
		1126,            // Expected event kind
		"page manifest", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Page manifest should have empty content
			if evt.Content != "" {
				helper.LogWarning(evt, "Page manifest content should be empty")
			}

			return validatePageManifestTags(helper, evt)
		},
	)
}

// ValidateSiteIndex validates NIP-YY Site Index events (kind 31126)
func ValidateSiteIndex(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",  // NIP number
		31126,        // Expected event kind
		"site index", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			return validateSiteIndexTags(helper, evt)
		},
	)
}

// ValidateEntrypoint validates NIP-YY Entrypoint events (kind 11126)
func ValidateEntrypoint(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web", // NIP number
		11126,       // Expected event kind
		"entrypoint", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Entrypoint should have empty content
			if evt.Content != "" {
				helper.LogWarning(evt, "Entrypoint content should be empty")
			}

			return validateEntrypointTags(helper, evt)
		},
	)
}

// validateAssetTags validates tags for asset events (kind 1125)
func validateAssetTags(helper *common.ValidationHelper, event *nostr.Event) error {
	var hasMTag bool
	var hasXTag bool
	var xTagValue string

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "m":
			if len(tag) < 2 || tag[1] == "" {
				return fmt.Errorf("m tag must have a MIME type value")
			}
			hasMTag = true
		case "x":
			if len(tag) != 2 {
				return fmt.Errorf("x tag must have exactly 2 elements")
			}
			hash := tag[1]
			if len(hash) != 64 {
				return fmt.Errorf("SHA-256 hash must be 64 hex characters, got %d", len(hash))
			}
			if !isHexString(hash) {
				return fmt.Errorf("SHA-256 hash must be valid hex")
			}
			hasXTag = true
			xTagValue = hash
		}
	}

	// Required tags validation
	if !hasMTag {
		return fmt.Errorf("asset must have an m (MIME type) tag")
	}

	if !hasXTag {
		return fmt.Errorf("asset must have an x tag with SHA-256 hash for content deduplication")
	}

	// CRITICAL SECURITY CHECK: Verify SHA-256 hash matches content
	if xTagValue != "" {
		computedHash := computeSHA256(event.Content)
		if xTagValue != computedHash {
			return fmt.Errorf("x tag value does not match content hash: expected %s, got %s", computedHash, xTagValue)
		}
	}

	return nil
}

// validatePageManifestTags validates tags for page manifest events (kind 1126)
func validatePageManifestTags(helper *common.ValidationHelper, event *nostr.Event) error {
	var hasETag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "e":
			if len(tag) < 2 {
				return fmt.Errorf("e tag must have at least 2 elements (event ID)")
			}
			// Validate event ID format (64 hex characters)
			eventID := tag[1]
			if len(eventID) != 64 || !isHexString(eventID) {
				return fmt.Errorf("invalid event ID in e tag: must be 64 hex characters")
			}
			hasETag = true
		}
	}

	// Required tags validation
	if !hasETag {
		return fmt.Errorf("page manifest must have at least one e tag referencing assets (kind 1125)")
	}

	return nil
}

// validateSiteIndexTags validates tags for site index events (kind 31126)
func validateSiteIndexTags(helper *common.ValidationHelper, event *nostr.Event) error {
	var hasDTag bool
	var hasXTag bool
	var dTagValue string
	var xTagValue string

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if len(tag) != 2 {
				return fmt.Errorf("d tag must have exactly 2 elements")
			}
			dTagValue = tag[1]
			// Validate d tag is 7-12 characters (truncated hash)
			if len(dTagValue) < 7 || len(dTagValue) > 12 {
				return fmt.Errorf("d tag must be 7-12 characters (truncated SHA-256 hash), got %d", len(dTagValue))
			}
			if !isHexString(dTagValue) {
				return fmt.Errorf("d tag must be valid hex")
			}
			hasDTag = true
		case "x":
			if len(tag) != 2 {
				return fmt.Errorf("x tag must have exactly 2 elements")
			}
			hash := tag[1]
			if len(hash) != 64 {
				return fmt.Errorf("x tag SHA-256 hash must be 64 hex characters, got %d", len(hash))
			}
			if !isHexString(hash) {
				return fmt.Errorf("x tag SHA-256 hash must be valid hex")
			}
			hasXTag = true
			xTagValue = hash
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("site index must have a d tag with truncated hash")
	}

	if !hasXTag {
		return fmt.Errorf("site index must have an x tag with full SHA-256 hash")
	}

	// Verify d tag is derived from x tag
	if xTagValue != "" && dTagValue != "" {
		if xTagValue[:len(dTagValue)] != dTagValue {
			return fmt.Errorf("d tag must be the first %d characters of x tag", len(dTagValue))
		}
	}

	// Validate content is valid JSON with required structure
	if event.Content == "" {
		return fmt.Errorf("site index content cannot be empty")
	}

	// Verify x tag matches content hash
	computedHash := computeSHA256(event.Content)
	if xTagValue != computedHash {
		return fmt.Errorf("x tag value does not match content hash: expected %s, got %s", computedHash, xTagValue)
	}

	// Parse and validate JSON structure
	var siteIndex struct {
		Routes         map[string]string `json:"routes"`
		Version        string            `json:"version,omitempty"`
		DefaultRoute   string            `json:"defaultRoute,omitempty"`
		NotFoundRoute  *string           `json:"notFoundRoute,omitempty"`
	}

	if err := json.Unmarshal([]byte(event.Content), &siteIndex); err != nil {
		return fmt.Errorf("site index content must be valid JSON: %w", err)
	}

	if len(siteIndex.Routes) == 0 {
		return fmt.Errorf("site index must contain at least one route mapping in 'routes' field")
	}

	// Validate manifest event IDs in the routes map
	for _, manifestID := range siteIndex.Routes {
		if len(manifestID) != 64 || !isHexString(manifestID) {
			return fmt.Errorf("invalid manifest event ID in site index routes: must be 64 hex characters")
		}
	}

	return nil
}

// validateEntrypointTags validates tags for entrypoint events (kind 11126)
func validateEntrypointTags(helper *common.ValidationHelper, event *nostr.Event) error {
	var hasATag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "a":
			if len(tag) < 2 {
				return fmt.Errorf("a tag must have at least 2 elements")
			}
			// Validate address coordinate format: "31126:<pubkey>:<d-tag-hash>"
			// The a tag format is: ["a", "31126:<pubkey>:<d-tag>", "relay-url (optional)"]
			address := tag[1]
			if address == "" {
				return fmt.Errorf("a tag address cannot be empty")
			}
			// Basic validation - should start with "31126:"
			if len(address) < 6 || address[:6] != "31126:" {
				return fmt.Errorf("a tag must reference a site index (31126:<pubkey>:<d-tag>)")
			}
			// tag[2] is optional relay URL hint - no validation needed
			hasATag = true
		}
	}

	// Required tags validation
	if !hasATag {
		return fmt.Errorf("entrypoint must have an a tag pointing to the current site index")
	}

	return nil
}

// computeSHA256 computes the SHA-256 hash of content
func computeSHA256(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, char := range s {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
