package nips

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-60: Cashu Wallets
// https://github.com/nostr-protocol/nips/blob/master/60.md
//
// This NIP defines operations for cashu-based wallets stored on relays.
// It includes wallet metadata (17375), token events (7375), spending history (7376),
// and quote redemption tracking (7374).

// ValidateWalletEvent validates a NIP-60 Wallet Event (kind 17375)
func ValidateWalletEvent(event *nostr.Event) error {
	// Basic event validation
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("wallet event validation failed: %w", err)
	}

	// Content must be encrypted (non-empty)
	if event.Content == "" {
		return fmt.Errorf("wallet event content cannot be empty - must contain encrypted wallet data")
	}

	// Check for required encrypted content structure
	// Note: We can't decrypt without the private key, but we can validate the structure exists
	if len(event.Content) < 10 {
		return fmt.Errorf("wallet event content too short - likely not properly encrypted")
	}

	// Validate that content appears to be NIP-44 encrypted
	if !isValidNIP44EncryptedContent(event.Content) {
		return fmt.Errorf("wallet event content does not appear to be valid NIP-44 encrypted data")
	}

	// Wallet events should not have public tags (all sensitive data should be in encrypted content)
	// But we allow some metadata tags if needed
	for _, tag := range event.Tags {
		if len(tag) < 1 {
			continue
		}
		
		// Allow specific safe tags
		switch tag[0] {
		case "expiration", "alt", "client":
			// These are acceptable for wallet events
			continue
		default:
			// Be permissive but warn about unexpected public tags
			// In a real implementation, you might want to be more strict
		}
	}

	return nil
}

// ValidateTokenEvent validates a NIP-60 Token Event (kind 7375)
func ValidateTokenEvent(event *nostr.Event) error {
	// Basic event validation
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("token event validation failed: %w", err)
	}

	// Content must be encrypted (non-empty)
	if event.Content == "" {
		return fmt.Errorf("token event content cannot be empty - must contain encrypted token data")
	}

	// Validate that content appears to be NIP-44 encrypted
	if !isValidNIP44EncryptedContent(event.Content) {
		return fmt.Errorf("token event content does not appear to be valid NIP-44 encrypted data")
	}

	// Token events should not have public tags (all sensitive data should be in encrypted content)
	for _, tag := range event.Tags {
		if len(tag) < 1 {
			continue
		}
		
		// Allow specific safe tags
		switch tag[0] {
		case "expiration", "alt", "client":
			// These are acceptable for token events
			continue
		default:
			// Be permissive but warn about unexpected public tags
		}
	}

	return nil
}

// ValidateSpendingHistoryEvent validates a NIP-60 Spending History Event (kind 7376)
func ValidateSpendingHistoryEvent(event *nostr.Event) error {
	// Basic event validation
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("spending history event validation failed: %w", err)
	}

	// Content must be encrypted (non-empty)
	if event.Content == "" {
		return fmt.Errorf("spending history event content cannot be empty - must contain encrypted spending data")
	}

	// Validate that content appears to be NIP-44 encrypted
	if !isValidNIP44EncryptedContent(event.Content) {
		return fmt.Errorf("spending history event content does not appear to be valid NIP-44 encrypted data")
	}

	// Validate e-tags in spending history
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		
		if tag[0] == "e" {
			// Validate e-tag structure
			if len(tag) < 4 {
				return fmt.Errorf("e-tag in spending history must have at least 4 elements: ['e', '<event-id>', '', '<marker>']")
			}
			
			// Validate event ID format (should be 64-char hex)
			eventID := tag[1]
			if len(eventID) != 64 || !isHexString(eventID) {
				return fmt.Errorf("invalid event ID format in e-tag: %s", eventID)
			}
			
			// Validate marker
			marker := tag[3]
			switch marker {
			case "created", "destroyed", "redeemed":
				// Valid markers
			default:
				return fmt.Errorf("invalid e-tag marker in spending history: %s (must be 'created', 'destroyed', or 'redeemed')", marker)
			}
		}
	}

	return nil
}

// ValidateQuoteEvent validates a NIP-60 Quote Event (kind 7374)
func ValidateQuoteEvent(event *nostr.Event) error {
	// Basic event validation
	if err := validateBasicEventStructure(event); err != nil {
		return fmt.Errorf("quote event validation failed: %w", err)
	}

	// Content must be encrypted (non-empty)
	if event.Content == "" {
		return fmt.Errorf("quote event content cannot be empty - must contain encrypted quote ID")
	}

	// Validate that content appears to be NIP-44 encrypted
	if !isValidNIP44EncryptedContent(event.Content) {
		return fmt.Errorf("quote event content does not appear to be valid NIP-44 encrypted data")
	}

	// Quote events SHOULD have expiration and mint tags
	var hasExpiration, hasMint bool
	
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		
		switch tag[0] {
		case "expiration":
			hasExpiration = true
			// Validate expiration timestamp
			if _, err := strconv.ParseInt(tag[1], 10, 64); err != nil {
				return fmt.Errorf("invalid expiration timestamp: %s", tag[1])
			}
		case "mint":
			hasMint = true
			// Validate mint URL
			if err := validateMintURL(tag[1]); err != nil {
				return fmt.Errorf("invalid mint URL: %w", err)
			}
		}
	}

	if !hasExpiration {
		return fmt.Errorf("quote event must have an expiration tag")
	}
	
	if !hasMint {
		return fmt.Errorf("quote event must have a mint tag")
	}

	return nil
}

// Helper function to validate basic event structure for NIP-60 events
func validateBasicEventStructure(event *nostr.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.ID == "" {
		return fmt.Errorf("event must have an ID")
	}

	if len(event.ID) != 64 || !isHexString(event.ID) {
		return fmt.Errorf("invalid event ID format")
	}

	if event.PubKey == "" {
		return fmt.Errorf("event must have a pubkey")
	}

	if len(event.PubKey) != 64 || !isHexString(event.PubKey) {
		return fmt.Errorf("invalid pubkey format")
	}

	if event.Tags == nil {
		return fmt.Errorf("event must have tags")
	}

	return nil
}

// Helper function to validate that content appears to be NIP-44 encrypted
func isValidNIP44EncryptedContent(content string) bool {
	// Basic validation - NIP-44 encrypted content should be base64-like
	// and have a reasonable minimum length
	if len(content) < 20 {
		return false
	}
	
	// Should contain base64 characters
	validBase64 := regexp.MustCompile(`^[A-Za-z0-9+/=]+$`)
	return validBase64.MatchString(content)
}

// Helper function to validate mint URLs
func validateMintURL(mintURL string) error {
	if mintURL == "" {
		return fmt.Errorf("mint URL cannot be empty")
	}

	// Parse the URL
	u, err := url.Parse(mintURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Must be HTTP or HTTPS
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("mint URL must use http or https scheme")
	}

	// Must have a host
	if u.Host == "" {
		return fmt.Errorf("mint URL must have a valid host")
	}

	return nil
}

// Additional validation helpers for encrypted content structure
// These would be used if we wanted to validate the structure of encrypted content
// after decryption (which we can't do without the private key)

type WalletContent struct {
	PrivKey string   `json:"privkey"`
	Mints   []string `json:"mints"`
}

type TokenContent struct {
	Mint   string                 `json:"mint"`
	Proofs []CashuProof          `json:"proofs"`
	Del    []string              `json:"del,omitempty"`
}

type CashuProof struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
	Secret string `json:"secret"`
	C      string `json:"C"`
}

type SpendingHistoryContent [][]string

// ValidateDecryptedWalletContent validates the structure of decrypted wallet content
// This is primarily for testing or when the client has access to decrypt the content
func ValidateDecryptedWalletContent(decryptedJSON string) error {
	var content [][]string
	if err := json.Unmarshal([]byte(decryptedJSON), &content); err != nil {
		return fmt.Errorf("wallet content must be valid JSON array: %w", err)
	}

	var hasPrivKey, hasMint bool
	
	for _, item := range content {
		if len(item) < 2 {
			return fmt.Errorf("wallet content items must have at least 2 elements")
		}
		
		switch item[0] {
		case "privkey":
			hasPrivKey = true
			// Validate private key format (64-char hex)
			if len(item[1]) != 64 || !isHexString(item[1]) {
				return fmt.Errorf("invalid private key format in wallet content")
			}
		case "mint":
			hasMint = true
			// Validate mint URL
			if err := validateMintURL(item[1]); err != nil {
				return fmt.Errorf("invalid mint URL in wallet content: %w", err)
			}
		}
	}

	if !hasPrivKey {
		return fmt.Errorf("wallet content must include a private key")
	}
	
	if !hasMint {
		return fmt.Errorf("wallet content must include at least one mint")
	}

	return nil
}

// ValidateDecryptedTokenContent validates the structure of decrypted token content
func ValidateDecryptedTokenContent(decryptedJSON string) error {
	var content TokenContent
	if err := json.Unmarshal([]byte(decryptedJSON), &content); err != nil {
		return fmt.Errorf("token content must be valid JSON object: %w", err)
	}

	// Validate mint URL
	if err := validateMintURL(content.Mint); err != nil {
		return fmt.Errorf("invalid mint URL in token content: %w", err)
	}

	// Must have at least one proof
	if len(content.Proofs) == 0 {
		return fmt.Errorf("token content must have at least one proof")
	}

	// Validate each proof
	for i, proof := range content.Proofs {
		if proof.ID == "" {
			return fmt.Errorf("proof %d must have an ID", i)
		}
		if proof.Amount <= 0 {
			return fmt.Errorf("proof %d must have a positive amount", i)
		}
		if proof.Secret == "" {
			return fmt.Errorf("proof %d must have a secret", i)
		}
		if proof.C == "" {
			return fmt.Errorf("proof %d must have a C value", i)
		}
		// Validate C value is valid hex (it's a public key point)
		if len(proof.C) != 66 || !strings.HasPrefix(proof.C, "02") && !strings.HasPrefix(proof.C, "03") {
			return fmt.Errorf("proof %d has invalid C value format", i)
		}
	}

	// Validate del array if present
	for i, eventID := range content.Del {
		if len(eventID) != 64 || !isHexString(eventID) {
			return fmt.Errorf("del[%d] has invalid event ID format: %s", i, eventID)
		}
	}

	return nil
}

// ValidateDecryptedSpendingContent validates the structure of decrypted spending history content
func ValidateDecryptedSpendingContent(decryptedJSON string) error {
	var content SpendingHistoryContent
	if err := json.Unmarshal([]byte(decryptedJSON), &content); err != nil {
		return fmt.Errorf("spending history content must be valid JSON array: %w", err)
	}

	var hasDirection, hasAmount bool
	
	for _, item := range content {
		if len(item) < 2 {
			return fmt.Errorf("spending history items must have at least 2 elements")
		}
		
		switch item[0] {
		case "direction":
			hasDirection = true
			if item[1] != "in" && item[1] != "out" {
				return fmt.Errorf("spending direction must be 'in' or 'out', got: %s", item[1])
			}
		case "amount":
			hasAmount = true
			// Validate amount is a positive number
			amount, err := strconv.ParseInt(item[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount format: %s", item[1])
			}
			if amount <= 0 {
				return fmt.Errorf("amount must be positive, got: %d", amount)
			}
		case "e":
			// Validate e-tag reference structure
			if len(item) < 4 {
				return fmt.Errorf("e-tag in spending content must have at least 4 elements")
			}
			// Validate event ID
			if len(item[1]) != 64 || !isHexString(item[1]) {
				return fmt.Errorf("invalid event ID in e-tag: %s", item[1])
			}
			// Validate marker
			marker := item[3]
			if marker != "created" && marker != "destroyed" {
				return fmt.Errorf("invalid e-tag marker in spending content: %s", marker)
			}
		}
	}

	if !hasDirection {
		return fmt.Errorf("spending history content must include direction")
	}
	
	if !hasAmount {
		return fmt.Errorf("spending history content must include amount")
	}

	return nil
}