package nips

import (
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-61: Nutzaps
// Validates nutzap info events (kind 10019) and nutzap events (kind 9321)

// ValidateNutzapInfoEvent validates a nutzap info event (kind 10019)
func ValidateNutzapInfoEvent(event *nostr.Event) error {
	// Validate event kind
	if event.Kind != 10019 {
		return fmt.Errorf("nutzap info event must be kind 10019")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid nutzap info event structure: %w", err)
	}

	// Track required tags
	var hasRelay, hasMint, hasPubkey bool

	// Validate tags
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue // Skip malformed tags
		}

		switch tag[0] {
		case "relay":
			hasRelay = true
			if err := validateRelayURL(tag[1]); err != nil {
				return fmt.Errorf("invalid relay URL in nutzap info: %w", err)
			}

		case "mint":
			hasMint = true
			if err := validateMintURL(tag[1]); err != nil {
				return fmt.Errorf("invalid mint URL in nutzap info: %w", err)
			}
			// Check optional base unit parameter
			if len(tag) >= 3 {
				if err := validateBaseUnit(tag[2]); err != nil {
					return fmt.Errorf("invalid base unit in mint tag: %w", err)
				}
			}
			// Check for additional base units
			for i := 3; i < len(tag); i++ {
				if err := validateBaseUnit(tag[i]); err != nil {
					return fmt.Errorf("invalid additional base unit in mint tag: %w", err)
				}
			}

		case "pubkey":
			hasPubkey = true
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid pubkey format in nutzap info: %s", tag[1])
			}
		}
	}

	// Check required tags
	if !hasRelay {
		return fmt.Errorf("nutzap info event must have at least one relay tag")
	}
	if !hasMint {
		return fmt.Errorf("nutzap info event must have at least one mint tag")
	}
	if !hasPubkey {
		return fmt.Errorf("nutzap info event must have a pubkey tag")
	}

	return nil
}

// ValidateNutzapEvent validates a nutzap event (kind 9321)
func ValidateNutzapEvent(event *nostr.Event) error {
	// Validate event kind
	if event.Kind != 9321 {
		return fmt.Errorf("nutzap event must be kind 9321")
	}

	// Validate basic event structure
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("invalid nutzap event structure: %w", err)
	}

	// Track required tags
	var hasProof, hasMintURL, hasRecipient bool
	var proofCount int

	// Validate tags
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue // Skip malformed tags
		}

		switch tag[0] {
		case "proof":
			hasProof = true
			proofCount++
			// Validate Cashu proof JSON
			if err := validateCashuProof(tag[1]); err != nil {
				return fmt.Errorf("invalid cashu proof in nutzap: %w", err)
			}

		case "u":
			hasMintURL = true
			if err := validateMintURL(tag[1]); err != nil {
				return fmt.Errorf("invalid mint URL in nutzap: %w", err)
			}

		case "p":
			hasRecipient = true
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid recipient pubkey in nutzap: %s", tag[1])
			}

		case "e":
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid event ID in nutzap e-tag: %s", tag[1])
			}
			// Optional relay URL in e-tag
			if len(tag) >= 3 && tag[2] != "" {
				if err := validateRelayURL(tag[2]); err != nil {
					return fmt.Errorf("invalid relay URL in nutzap e-tag: %w", err)
				}
			}

		case "k":
			if len(tag) >= 2 {
				if !isNumericString(tag[1]) {
					return fmt.Errorf("invalid kind value in nutzap k-tag: %s", tag[1])
				}
			}
		}
	}

	// Check required tags
	if !hasProof {
		return fmt.Errorf("nutzap event must have at least one proof tag")
	}
	if !hasMintURL {
		return fmt.Errorf("nutzap event must have a mint URL (u tag)")
	}
	if !hasRecipient {
		return fmt.Errorf("nutzap event must have a recipient pubkey (p tag)")
	}

	// Sanity check for proof count
	if proofCount > 100 {
		return fmt.Errorf("too many proof tags (%d), maximum 100 recommended", proofCount)
	}

	return nil
}

// Helper function to validate base units in mint tags
func validateBaseUnit(unit string) error {
	if unit == "" {
		return fmt.Errorf("base unit cannot be empty")
	}

	// Allow any string - unknown base units are allowed by NIP-61
	// but may cause UX issues in client implementations

	return nil
}

// Helper function to validate P2PK proof JSON structure for NIP-61
func validateCashuProof(proofJSON string) error {
	var proof CashuProof
	if err := json.Unmarshal([]byte(proofJSON), &proof); err != nil {
		return fmt.Errorf("invalid proof JSON: %w", err)
	}

	// Validate proof fields
	if proof.Amount <= 0 {
		return fmt.Errorf("proof amount must be positive")
	}

	if proof.ID == "" {
		return fmt.Errorf("proof must have a keyset ID")
	}

	if proof.C == "" {
		return fmt.Errorf("proof must have a commitment (C)")
	}

	if proof.Secret == "" {
		return fmt.Errorf("proof must have a secret")
	}

	// For NIP-61, we specifically need P2PK proofs
	// Validate P2PK secret structure for nutzaps
	if err := validateP2PKSecret(proof.Secret); err != nil {
		return fmt.Errorf("invalid P2PK secret in proof: %w", err)
	}

	return nil
}

// Helper function to validate P2PK secret structure
func validateP2PKSecret(secret string) error {
	// P2PK secrets should be JSON arrays starting with "P2PK"
	var secretArray []interface{}
	if err := json.Unmarshal([]byte(secret), &secretArray); err != nil {
		return fmt.Errorf("P2PK secret must be valid JSON array: %w", err)
	}

	if len(secretArray) < 2 {
		return fmt.Errorf("P2PK secret must have at least 2 elements")
	}

	// First element should be "P2PK"
	if secretArray[0] != "P2PK" {
		return fmt.Errorf("P2PK secret must start with 'P2PK'")
	}

	// Second element should be an object with nonce and data
	secretData, ok := secretArray[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("P2PK secret second element must be an object")
	}

	// Check for required fields
	if _, hasNonce := secretData["nonce"]; !hasNonce {
		return fmt.Errorf("P2PK secret must have a nonce field")
	}
	if _, hasData := secretData["data"]; !hasData {
		return fmt.Errorf("P2PK secret must have a data field")
	}

	return nil
}

// Helper function to check if string is numeric
func isNumericString(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Additional validation helpers for extended functionality

// ValidateNutzapRedemption validates a NIP-60 spending history event that redeems nutzaps
func ValidateNutzapRedemption(event *nostr.Event) error {
	// This would validate kind:7376 events that redeem nutzaps
	// Basic validation for spending history events with nutzap redemption markers

	if event.Kind != 7376 {
		return fmt.Errorf("nutzap redemption must be kind 7376")
	}

	// Content must be encrypted (non-empty)
	if event.Content == "" {
		return fmt.Errorf("nutzap redemption content cannot be empty - must contain encrypted spending data")
	}

	// Look for redeemed e-tags pointing to nutzap events
	var hasRedeemedNutzap bool
	for _, tag := range event.Tags {
		if len(tag) >= 4 && tag[0] == "e" && tag[3] == "redeemed" {
			hasRedeemedNutzap = true
			// Validate event ID format
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid nutzap event ID in redemption e-tag: %s", tag[1])
			}
		}
	}

	if !hasRedeemedNutzap {
		return fmt.Errorf("nutzap redemption must have at least one e-tag with 'redeemed' marker")
	}

	return nil
}

// ValidateNutzapOfflineVerification validates offline verification requirements
func ValidateNutzapOfflineVerification(nutzapEvent *nostr.Event, recipientInfo *nostr.Event) error {
	// This validates that a nutzap can be verified offline according to NIP-61 requirements

	if nutzapEvent.Kind != 9321 {
		return fmt.Errorf("nutzap event must be kind 9321")
	}

	if recipientInfo != nil && recipientInfo.Kind != 10019 {
		return fmt.Errorf("recipient info must be kind 10019")
	}

	// Extract recipient pubkey from nutzap
	var recipientPubkey string
	for _, tag := range nutzapEvent.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			recipientPubkey = tag[1]
			break
		}
	}

	if recipientPubkey == "" {
		return fmt.Errorf("nutzap event must have a recipient pubkey for offline verification")
	}

	// If recipient info is provided, verify it matches
	if recipientInfo != nil && recipientInfo.PubKey != recipientPubkey {
		return fmt.Errorf("recipient info pubkey does not match nutzap recipient")
	}

	// Extract mint URL from nutzap
	var mintURL string
	for _, tag := range nutzapEvent.Tags {
		if len(tag) >= 2 && tag[0] == "u" {
			mintURL = tag[1]
			break
		}
	}

	if mintURL == "" {
		return fmt.Errorf("nutzap event must have a mint URL for offline verification")
	}

	// If recipient info is provided, verify mint is supported
	if recipientInfo != nil {
		var mintSupported bool
		for _, tag := range recipientInfo.Tags {
			if len(tag) >= 2 && tag[0] == "mint" && tag[1] == mintURL {
				mintSupported = true
				break
			}
		}
		if !mintSupported {
			return fmt.Errorf("mint %s is not supported by recipient according to their nutzap info", mintURL)
		}
	}

	return nil
}