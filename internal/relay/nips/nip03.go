package nips

import (
	"encoding/base64"
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-03: OpenTimestamps Attestations for Events
// https://github.com/nostr-protocol/nips/blob/master/03.md

// ValidateOpenTimestampsAttestation validates NIP-03 OpenTimestamps attestation events (kind 1040)
func ValidateOpenTimestampsAttestation(evt *nostr.Event) error {
	if evt.Kind != 1040 {
		return fmt.Errorf("invalid event kind for OpenTimestamps attestation: %d", evt.Kind)
	}

	// Must have at least one "e" tag referencing the attested event
	hasEventTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasEventTag = true
			// Validate the event ID format (should be 64-char hex)
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid event ID in 'e' tag: %s", tag[1])
			}
			// Validate hex format
			for _, c := range tag[1] {
				if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
					return fmt.Errorf("invalid hex format in event ID: %s", tag[1])
				}
			}
		}
	}

	if !hasEventTag {
		return fmt.Errorf("OpenTimestamps attestation must reference at least one event with 'e' tag")
	}

	// Optional 'alt' tag validation
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "alt" && tag[1] != "opentimestamps attestation" {
			return fmt.Errorf("if 'alt' tag is present, it must have value 'opentimestamps attestation'")
		}
	}

	// Content must contain the OpenTimestamps proof
	if evt.Content == "" {
		return fmt.Errorf("OpenTimestamps attestation must have base64-encoded OTS file content")
	}

	// Try to decode the base64 content to verify it's valid
	_, err := base64.StdEncoding.DecodeString(evt.Content)
	if err != nil {
		return fmt.Errorf("invalid base64 content in OpenTimestamps attestation: %v", err)
	}

	// Set a size limit on the OTS file content (2KB)
	if len(evt.Content) > 2048 {
		return fmt.Errorf("OTS file content too large (max 2KB)")
	}

	return nil
}

// IsOpenTimestampsAttestation checks if an event is an OpenTimestamps attestation
func IsOpenTimestampsAttestation(evt *nostr.Event) bool {
	return evt.Kind == 1040
}
