package nips

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

const (
	NIP44Version1 = 1
	NIP44Version2 = 2
	V2NonceLength = 24 // bytes (XChaCha20-Poly1305 standard)
)

// NIP44PayloadV1 represents v1 structure
type NIP44PayloadV1 struct {
	V          int    `json:"v"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

// ValidateNIP44Payload validates a NIP-44 encrypted event for v1 and v2
func ValidateNIP44Payload(event nostr.Event) error {
	// NIP-44 events must have an "encrypted" tag.
	if event.Tags.Find("encrypted") == nil {
		return errors.New("missing 'encrypted' tag for NIP-44 event")
	}

	// Recipient tag check
	if len(event.Tags) == 0 {
		return fmt.Errorf("%s", FormatDMError("missing_recipient"))
	}
	hasRecipient := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			if !nostr.IsValid32ByteHex(tag[1]) {
				return fmt.Errorf("%s", FormatDMError("invalid_pubkey"))
			}
			hasRecipient = true
			break
		}
	}
	if !hasRecipient {
		return fmt.Errorf("%s", FormatDMError("missing_recipient"))
	}

	// Empty content is allowed for NIP-44, representing a placeholder or signal.
	if event.Content == "" {
		return nil
	}

	// If content is not empty, it must be a valid base64 string.
	decoded, err := base64.StdEncoding.DecodeString(event.Content)
	if err != nil {
		return fmt.Errorf("invalid: must be base64 encoded")
	}

	// Try unmarshal as v1 JSON
	var payloadV1 NIP44PayloadV1
	if err := json.Unmarshal(decoded, &payloadV1); err == nil {
		// Check version field
		if payloadV1.V != NIP44Version1 {
			return fmt.Errorf("unsupported NIP-44 version: %d", payloadV1.V)
		}
		// Nonce & ciphertext fields must be present and base64
		if payloadV1.Nonce == "" {
			return fmt.Errorf("missing nonce field")
		}
		if _, err := base64.StdEncoding.DecodeString(payloadV1.Nonce); err != nil {
			return fmt.Errorf("invalid nonce base64 encoding: %w", err)
		}
		if payloadV1.Ciphertext == "" {
			return fmt.Errorf("missing ciphertext field")
		}
		if _, err := base64.StdEncoding.DecodeString(payloadV1.Ciphertext); err != nil {
			return fmt.Errorf("invalid ciphertext base64 encoding: %w", err)
		}
		return nil // v1 valid
	}

	// Try v2: binary envelope ([2][24B nonce][N ciphertext])
	if len(decoded) < 1+V2NonceLength+1 {
		return fmt.Errorf("invalid NIP-44 v2 envelope: too short")
	}
	if decoded[0] != NIP44Version2 {
		return fmt.Errorf("unsupported NIP-44 version: %d", int(decoded[0]))
	}
	// nonce := decoded[1 : 1+V2NonceLength]
	ciphertext := decoded[1+V2NonceLength:]

	if len(ciphertext) == 0 {
		return errors.New("invalid NIP-44 v2 envelope: missing ciphertext")
	}
	// (Optional: add more checks, e.g., nonce not all zeros, min ciphertext length, etc.)

	return nil // v2 valid
}

// IsNIP44Payload checks if a content string is likely v1 or v2 NIP-44
func IsNIP44Payload(content string) bool {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return false
	}
	// v1: try JSON
	var payloadV1 NIP44PayloadV1
	if err := json.Unmarshal(decoded, &payloadV1); err == nil {
		return payloadV1.V == NIP44Version1 && payloadV1.Nonce != "" && payloadV1.Ciphertext != ""
	}
	// v2: version byte, correct minimum length
	if len(decoded) >= 1+V2NonceLength+1 && decoded[0] == NIP44Version2 {
		return true
	}
	return false
}
