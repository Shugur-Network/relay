package nips

import (
	"encoding/base64"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-04: Encrypted Direct Message (deprecated, use NIP-17)
// https://github.com/nostr-protocol/nips/blob/master/04.md

// ValidateEncryptedDirectMessage validates kind 4 events - handles both NIP-04 and NIP-44
func ValidateEncryptedDirectMessage(evt *nostr.Event) error {
	if evt.Kind != 4 {
		return fmt.Errorf("invalid event kind for encrypted direct message: %d", evt.Kind)
	}

	// Must have at least one "p" tag with the recipient pubkey(s)
	pTags := 0
	hasRecipient := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			pTags++
			hasRecipient = true
			// Validate pubkey format (should be 64-char hex)
			if len(tag[1]) != 64 {
				return fmt.Errorf("invalid pubkey in 'p' tag: %s", tag[1])
			}
		}
	}

	if pTags == 0 {
		return fmt.Errorf("encrypted direct message must have at least one 'p' tag")
	}

	if !hasRecipient {
		return fmt.Errorf("%s", FormatDMError("missing_recipient"))
	}

	// If the content looks like NIP-44 or the event has the 'encrypted' tag,
	// it must be validated as a NIP-44 event.
	encryptedTag := evt.Tags.Find("encrypted")
	if IsNIP44Payload(evt.Content) || encryptedTag != nil {
		return ValidateNIP44Payload(*evt) // NIP-44 validation
	}

	// Fall back to NIP-04 validation (simple base64)
	if evt.Content == "" {
		return fmt.Errorf("encrypted direct message must have encrypted content")
	}

	_, err := base64.StdEncoding.DecodeString(evt.Content)
	if err != nil {
		return fmt.Errorf("%s", FormatDMError("invalid_base64"))
	}

	return nil // Valid NIP-04
}

// IsEncryptedDirectMessage checks if an event is an encrypted direct message
func IsEncryptedDirectMessage(evt *nostr.Event) bool {
	return evt.Kind == 4
}
