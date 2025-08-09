package nips

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
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
			break
		}
	}

	if !hasEventTag {
		return fmt.Errorf("OpenTimestamps attestation must reference at least one event with 'e' tag")
	}

	// Content should contain the OpenTimestamps proof
	if evt.Content == "" {
		return fmt.Errorf("OpenTimestamps attestation must have content with the proof")
	}

	return nil
}

// IsOpenTimestampsAttestation checks if an event is an OpenTimestamps attestation
func IsOpenTimestampsAttestation(evt *nostr.Event) bool {
	return evt.Kind == 1040
}
