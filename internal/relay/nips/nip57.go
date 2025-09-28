package nips

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// isHexChar checks if a character is a valid hexadecimal character
func isHexChar(char rune) bool {
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')
}

// ValidateZapRequest validates NIP-57 zap request events (kind 9734)
func ValidateZapRequest(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-57: Validating zap request event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 9734 {
		return fmt.Errorf("invalid kind for zap request: expected 9734, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateZapRequestTags(event); err != nil {
		return fmt.Errorf("invalid zap request tags: %w", err)
	}

	logger.Debug("NIP-57: Zap request validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateZapReceipt validates NIP-57 zap receipt events (kind 9735)
func ValidateZapReceipt(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-57: Validating zap receipt event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 9735 {
		return fmt.Errorf("invalid kind for zap receipt: expected 9735, got %d", event.Kind)
	}

	// Content SHOULD be empty for zap receipts
	if event.Content != "" {
		logger.Warn("NIP-57: Zap receipt content should be empty",
			zap.String("event_id", event.ID),
			zap.String("content", event.Content))
	}

	// Validate required tags for zap receipts
	if err := validateZapReceiptTags(event); err != nil {
		return fmt.Errorf("invalid zap receipt tags: %w", err)
	}

	logger.Debug("NIP-57: Zap receipt validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// validateZapRequestTags validates the tag structure for zap request events
func validateZapRequestTags(event *nostr.Event) error {
	var hasRelaysTag bool
	var hasPTag bool
	var pTagCount int
	var eTagCount int
	var aTagCount int
	var pTagValue string

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue // Skip malformed tags
		}

		switch tag[0] {
		case "relays":
			if err := validateRelaysTag(tag); err != nil {
				return fmt.Errorf("invalid relays tag: %w", err)
			}
			hasRelaysTag = true

		case "amount":
			if err := validateAmountTag(tag); err != nil {
				return fmt.Errorf("invalid amount tag: %w", err)
			}

		case "lnurl":
			if err := validateLnurlTag(tag); err != nil {
				return fmt.Errorf("invalid lnurl tag: %w", err)
			}

		case "p":
			if err := validateZapPTag(tag); err != nil {
				return fmt.Errorf("invalid p tag: %w", err)
			}
			hasPTag = true
			pTagCount++
			pTagValue = tag[1]

		case "e":
			if err := validateZapETag(tag); err != nil {
				return fmt.Errorf("invalid e tag: %w", err)
			}
			eTagCount++

		case "a":
			if err := validateZapATag(tag); err != nil {
				return fmt.Errorf("invalid a tag: %w", err)
			}
			aTagCount++

		case "k":
			if err := validateZapKTag(tag); err != nil {
				return fmt.Errorf("invalid k tag: %w", err)
			}

		case "P":
			if err := validateZapCapitalPTag(tag); err != nil {
				return fmt.Errorf("invalid P tag: %w", err)
			}
		}
	}

	// Required validations
	if !hasPTag {
		return fmt.Errorf("zap request must include exactly one 'p' tag with recipient pubkey")
	}

	if pTagCount != 1 {
		return fmt.Errorf("zap request must have exactly one 'p' tag, found %d", pTagCount)
	}

	if eTagCount > 1 {
		return fmt.Errorf("zap request must have 0 or 1 'e' tags, found %d", eTagCount)
	}

	if aTagCount > 1 {
		return fmt.Errorf("zap request must have 0 or 1 'a' tags, found %d", aTagCount)
	}

	// Relays tag is recommended but not required
	if !hasRelaysTag {
		logger.Warn("NIP-57: Zap request missing recommended 'relays' tag",
			zap.String("event_id", event.ID),
			zap.String("recipient", pTagValue))
	}

	return nil
}

// validateZapReceiptTags validates the tag structure for zap receipt events
func validateZapReceiptTags(event *nostr.Event) error {
	var hasBolt11Tag bool
	var hasDescriptionTag bool
	var hasPTag bool

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue // Skip malformed tags
		}

		switch tag[0] {
		case "p":
			if err := validateZapPTag(tag); err != nil {
				return fmt.Errorf("invalid p tag: %w", err)
			}
			hasPTag = true

		case "P":
			if err := validateZapCapitalPTag(tag); err != nil {
				return fmt.Errorf("invalid P tag: %w", err)
			}

		case "e":
			if err := validateZapETag(tag); err != nil {
				return fmt.Errorf("invalid e tag: %w", err)
			}

		case "a":
			if err := validateZapATag(tag); err != nil {
				return fmt.Errorf("invalid a tag: %w", err)
			}

		case "k":
			if err := validateZapKTag(tag); err != nil {
				return fmt.Errorf("invalid k tag: %w", err)
			}

		case "bolt11":
			if err := validateBolt11Tag(tag); err != nil {
				return fmt.Errorf("invalid bolt11 tag: %w", err)
			}
			hasBolt11Tag = true

		case "description":
			if err := validateDescriptionTag(tag); err != nil {
				return fmt.Errorf("invalid description tag: %w", err)
			}
			hasDescriptionTag = true

		case "preimage":
			if err := validatePreimageTag(tag); err != nil {
				return fmt.Errorf("invalid preimage tag: %w", err)
			}
		}
	}

	// Required tags for zap receipts
	if !hasPTag {
		return fmt.Errorf("zap receipt must include 'p' tag with recipient pubkey")
	}

	if !hasBolt11Tag {
		return fmt.Errorf("zap receipt must include 'bolt11' tag with invoice")
	}

	if !hasDescriptionTag {
		return fmt.Errorf("zap receipt must include 'description' tag with JSON-encoded zap request")
	}

	return nil
}

// validateRelaysTag validates a relays tag for zap requests
func validateRelaysTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("relays tag must have at least one relay URL")
	}

	// Validate each relay URL
	for i := 1; i < len(tag); i++ {
		relay := tag[i]
		if len(relay) == 0 {
			return fmt.Errorf("relay URL cannot be empty")
		}

		// Basic WebSocket URL validation
		if !strings.HasPrefix(relay, "ws://") && !strings.HasPrefix(relay, "wss://") {
			return fmt.Errorf("relay URL must start with ws:// or wss://, got: %s", relay)
		}

		if len(relay) < 8 { // Minimum: "ws://x.y"
			return fmt.Errorf("relay URL too short: %s", relay)
		}
	}

	return nil
}

// validateAmountTag validates an amount tag (millisats as string)
func validateAmountTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("amount tag must have exactly 2 elements")
	}

	amountStr := tag[1]
	if len(amountStr) == 0 {
		return fmt.Errorf("amount cannot be empty")
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return fmt.Errorf("amount must be a valid integer: %w", err)
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be positive, got: %d", amount)
	}

	// Reasonable limits (1 millisat to 100 million sats = 100 billion millisats)
	if amount > 100000000000 {
		return fmt.Errorf("amount too large: %d millisats", amount)
	}

	return nil
}

// validateLnurlTag validates an lnurl tag
func validateLnurlTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("lnurl tag must have exactly 2 elements")
	}

	lnurl := tag[1]
	if len(lnurl) == 0 {
		return fmt.Errorf("lnurl cannot be empty")
	}

	// Basic lnurl format validation (bech32 with lnurl prefix)
	if !strings.HasPrefix(lnurl, "lnurl1") {
		return fmt.Errorf("lnurl must start with 'lnurl1', got: %s", lnurl)
	}

	if len(lnurl) < 10 {
		return fmt.Errorf("lnurl too short: %s", lnurl)
	}

	return nil
}

// validateZapPTag validates a p tag (pubkey) for zap events
func validateZapPTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("p tag must have exactly 2 elements")
	}

	pubkey := tag[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	// Validate hex format
	for _, char := range pubkey {
		if !isHexChar(char) {
			return fmt.Errorf("pubkey must be valid hex")
		}
	}

	return nil
}

// validateZapCapitalPTag validates a P tag (zap sender pubkey) for zap receipts
func validateZapCapitalPTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("p tag must have exactly 2 elements")
	}

	pubkey := tag[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("sender pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	// Validate hex format
	for _, char := range pubkey {
		if !isHexChar(char) {
			return fmt.Errorf("sender pubkey must be valid hex")
		}
	}

	return nil
}

// validateZapETag validates an e tag (event id) for zap events
func validateZapETag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("e tag must have exactly 2 elements")
	}

	eventId := tag[1]
	if len(eventId) != 64 {
		return fmt.Errorf("event id must be 64 hex characters, got %d", len(eventId))
	}

	// Validate hex format
	for _, char := range eventId {
		if !isHexChar(char) {
			return fmt.Errorf("event id must be valid hex")
		}
	}

	return nil
}

// validateZapATag validates an a tag (event coordinate) for zap events
func validateZapATag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("a tag must have exactly 2 elements")
	}

	coordinate := tag[1]
	if len(coordinate) == 0 {
		return fmt.Errorf("event coordinate cannot be empty")
	}

	// Basic event coordinate format validation (kind:pubkey:d-tag)
	parts := strings.Split(coordinate, ":")
	if len(parts) != 3 {
		return fmt.Errorf("event coordinate must have format 'kind:pubkey:d-tag', got: %s", coordinate)
	}

	// Validate kind is numeric
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return fmt.Errorf("event coordinate kind must be numeric: %w", err)
	}

	// Validate pubkey is 64 hex chars
	if len(parts[1]) != 64 {
		return fmt.Errorf("event coordinate pubkey must be 64 hex characters")
	}

	// d-tag can be any string, so no validation needed for parts[2]

	return nil
}

// validateZapKTag validates a k tag (kind) for zap events
func validateZapKTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("k tag must have exactly 2 elements")
	}

	kindStr := tag[1]
	if len(kindStr) == 0 {
		return fmt.Errorf("kind cannot be empty")
	}

	kind, err := strconv.Atoi(kindStr)
	if err != nil {
		return fmt.Errorf("kind must be a valid integer: %w", err)
	}

	if kind < 0 || kind > 65535 {
		return fmt.Errorf("kind must be between 0 and 65535, got: %d", kind)
	}

	return nil
}

// validateBolt11Tag validates a bolt11 tag (lightning invoice)
func validateBolt11Tag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("bolt11 tag must have exactly 2 elements")
	}

	bolt11 := tag[1]
	if len(bolt11) == 0 {
		return fmt.Errorf("bolt11 invoice cannot be empty")
	}

	// Basic bolt11 format validation
	if !strings.HasPrefix(strings.ToLower(bolt11), "ln") {
		return fmt.Errorf("bolt11 invoice must start with 'ln', got: %s", bolt11[:2])
	}

	if len(bolt11) < 20 {
		return fmt.Errorf("bolt11 invoice too short: %d characters", len(bolt11))
	}

	return nil
}

// validateDescriptionTag validates a description tag (JSON-encoded zap request)
func validateDescriptionTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("description tag must have exactly 2 elements")
	}

	description := tag[1]
	if len(description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}

	// Validate that description is valid JSON
	var zapRequest interface{}
	if err := json.Unmarshal([]byte(description), &zapRequest); err != nil {
		return fmt.Errorf("description must be valid JSON: %w", err)
	}

	// Try to parse as nostr event
	var event nostr.Event
	if err := json.Unmarshal([]byte(description), &event); err != nil {
		return fmt.Errorf("description must be a valid nostr event: %w", err)
	}

	// Validate it's a zap request event
	if event.Kind != 9734 {
		return fmt.Errorf("description must contain a kind 9734 zap request, got kind %d", event.Kind)
	}

	return nil
}

// validatePreimageTag validates a preimage tag (optional payment proof)
func validatePreimageTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("preimage tag must have exactly 2 elements")
	}

	preimage := tag[1]
	if len(preimage) == 0 {
		return fmt.Errorf("preimage cannot be empty")
	}

	// Preimage should be 64 hex characters (32 bytes)
	if len(preimage) != 64 {
		return fmt.Errorf("preimage must be 64 hex characters, got %d", len(preimage))
	}

	// Validate hex format
	for _, char := range preimage {
		if !isHexChar(char) {
			return fmt.Errorf("preimage must be valid hex")
		}
	}

	return nil
}