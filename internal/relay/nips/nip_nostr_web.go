package nips

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Shugur-Network/relay/internal/relay/nips/common"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateHTMLContent validates NIP-YY HTML content events (kind 40000)
func ValidateHTMLContent(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",           // NIP number
		40000,          // Expected event kind
		"HTML content", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			return validateImmutableAsset(helper, evt, "text/html")
		},
	)
}

// ValidateCSSStylesheet validates NIP-YY CSS stylesheet events (kind 40001)
func ValidateCSSStylesheet(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",            // NIP number
		40001,           // Expected event kind
		"CSS stylesheet", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			return validateImmutableAsset(helper, evt, "text/css")
		},
	)
}

// ValidateJavaScriptModule validates NIP-YY JavaScript module events (kind 40002)
func ValidateJavaScriptModule(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",                 // NIP number
		40002,                // Expected event kind
		"JavaScript module", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			return validateImmutableAsset(helper, evt, "text/javascript")
		},
	)
}

// ValidateComponentFragment validates NIP-YY Component/Fragment events (kind 40003)
func ValidateComponentFragment(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",                  // NIP number
		40003,                 // Expected event kind
		"component/fragment", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			// Components can have various MIME types, so we validate tags but don't enforce specific MIME
			return validateImmutableAssetTags(helper, evt, false)
		},
	)
}

// ValidatePageManifest validates NIP-YY Page Manifest events (kind 34235)
func ValidatePageManifest(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",            // NIP number
		34235,           // Expected event kind
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

// ValidateSiteIndex validates NIP-YY Site Index events (kind 34236)
func ValidateSiteIndex(event *nostr.Event) error {
	return common.ValidateEventWithCallback(
		event,
		"Nostr Web",          // NIP number
		34236,         // Expected event kind
		"site index", // Event name for logging
		func(helper *common.ValidationHelper, evt *nostr.Event) error {
			return validateSiteIndexTags(helper, evt)
		},
	)
}

// validateImmutableAsset validates immutable asset events (kinds 40000-40003)
func validateImmutableAsset(helper *common.ValidationHelper, event *nostr.Event, expectedMimeType string) error {
	return validateImmutableAssetTags(helper, event, true)
}

// validateImmutableAssetTags validates tags for immutable asset events
func validateImmutableAssetTags(helper *common.ValidationHelper, event *nostr.Event, requireSpecificMime bool) error {
	var hasMTag bool
	var hasSHA256Tag bool
	var sha256Value string

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
		case "sha256":
			if len(tag) != 2 {
				return fmt.Errorf("sha256 tag must have exactly 2 elements")
			}
			hash := tag[1]
			if len(hash) != 64 {
				return fmt.Errorf("SHA-256 hash must be 64 hex characters, got %d", len(hash))
			}
			if !isHexString(hash) {
				return fmt.Errorf("SHA-256 hash must be valid hex")
			}
			hasSHA256Tag = true
			sha256Value = hash
		}
	}

	// Required tags validation
	if !hasMTag {
		return fmt.Errorf("immutable asset must have an m (MIME type) tag")
	}

	if !hasSHA256Tag {
		return fmt.Errorf("immutable asset must have a sha256 tag for Subresource Integrity")
	}

	// CRITICAL SECURITY CHECK: Verify SHA-256 hash matches content
	if sha256Value != "" {
		computedHash := computeSHA256(event.Content)
		if sha256Value != computedHash {
			return fmt.Errorf("sha256 tag value does not match content hash: expected %s, got %s", computedHash, sha256Value)
		}
	}

	return nil
}

// validatePageManifestTags validates tags for page manifest events
func validatePageManifestTags(helper *common.ValidationHelper, event *nostr.Event) error {
	var hasDTag bool
	var hasETag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if len(tag) < 2 || tag[1] == "" {
				return fmt.Errorf("d tag must have a route value")
			}
			hasDTag = true
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
	if !hasDTag {
		return fmt.Errorf("page manifest must have a d tag with route path")
	}

	if !hasETag {
		return fmt.Errorf("page manifest must have at least one e tag referencing assets")
	}

	return nil
}

// validateSiteIndexTags validates tags for site index events
func validateSiteIndexTags(helper *common.ValidationHelper, event *nostr.Event) error {
	var hasDTag bool
	var hasDefaultRoute bool
	var dTagValue string

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
			hasDTag = true
		case "default_route":
			if len(tag) < 2 || tag[1] == "" {
				return fmt.Errorf("default_route tag must have a route value")
			}
			hasDefaultRoute = true
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("site index must have a d tag")
	}

	if dTagValue != "site-index" {
		return fmt.Errorf("site index d tag must have value 'site-index', got '%s'", dTagValue)
	}

	if !hasDefaultRoute {
		return fmt.Errorf("site index must have a default_route tag")
	}

	// Validate content is valid JSON with basic structure check
	if event.Content == "" {
		return fmt.Errorf("site index content cannot be empty")
	}

	var routeMap map[string]string
	if err := json.Unmarshal([]byte(event.Content), &routeMap); err != nil {
		return fmt.Errorf("site index content must be valid JSON: %w", err)
	}

	if len(routeMap) == 0 {
		return fmt.Errorf("site index must contain at least one route mapping")
	}

	// Validate event IDs in the map are properly formatted
	for _, manifestID := range routeMap {
		if len(manifestID) != 64 || !isHexString(manifestID) {
			return fmt.Errorf("invalid manifest event ID in site index: must be 64 hex characters")
		}
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
