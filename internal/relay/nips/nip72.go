package nips

import (
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
	// Community post tag validation - placeholder for now
	return nil
}

func validateApprovalEventTags(helper *common.ValidationHelper, event *nostr.Event) error {
	// Approval event tag validation - placeholder for now
	return nil
}

// Basic helper functions for NIP-72 (renamed to avoid conflicts)
func validateCommunityIdentifier72(id string) error {
	return nil
}

func validateCommunityName72(name string) error {
	return nil
}

func validateCommunityDescription72(desc string) error {
	return nil
}

func validateCommunityImageTag72(tag []string) error {
	return nil
}

func validateCommunityRelay72(tag []string) error {
	return nil
}

func validateCommunityReference72(ref string, expectedKind string) error {
	return nil
}

func isHexString72(s string) bool {
	// Simple hex string validation
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
