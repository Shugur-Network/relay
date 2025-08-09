package nips

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-24: Extra metadata fields and tags
// https://github.com/nostr-protocol/nips/blob/master/24.md

// ValidateExtraMetadata validates NIP-24 extra metadata in events
func ValidateExtraMetadata(evt *nostr.Event) error {
	// NIP-24 allows additional metadata fields and tags
	// Most validation is permissive as it extends existing event kinds

	// Validate any "nonce" tags if present (for proof of work)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "nonce" {
			// Nonce should be numeric
			if len(tag[1]) == 0 {
				return fmt.Errorf("invalid nonce tag: empty value")
			}
		}
	}

	return nil
}

// HasExtraMetadata checks if an event has NIP-24 extra metadata
func HasExtraMetadata(evt *nostr.Event) bool {
	// Check for common NIP-24 tags
	for _, tag := range evt.Tags {
		if len(tag) >= 1 {
			switch tag[0] {
			case "nonce", "subject", "client", "relays":
				return true
			}
		}
	}
	return false
}
