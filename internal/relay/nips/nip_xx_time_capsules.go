package nips

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateTimeCapsuleEvent validates time capsule events according to NIP-XX
// This is minimal validation - drand beacon verification is left to clients
func ValidateTimeCapsuleEvent(evt *nostr.Event) error {
	// Must be kind 1041
	if evt.Kind != constants.KindTimeCapsule {
		return fmt.Errorf("invalid kind: expected %d, got %d", constants.KindTimeCapsule, evt.Kind)
	}

	// Validate content format
	payload, err := base64.StdEncoding.DecodeString(evt.Content)
	if err != nil {
		return fmt.Errorf("invalid base64 content: %w", err)
	}

	if len(payload) < 1 {
		return fmt.Errorf(constants.ErrMalformedPayload)
	}

	mode := payload[0]
	if mode != constants.ModePublic && mode != constants.ModePrivate {
		return fmt.Errorf("%s: 0x%02x", constants.ErrInvalidMode, mode)
	}

	// Validate tlock tag
	tlockTag := findTlockTag(evt.Tags)
	if tlockTag == nil {
		return fmt.Errorf(constants.ErrMissingTlockTag)
	}

	if err := validateTlockTagBasic(tlockTag); err != nil {
		return fmt.Errorf("invalid tlock tag: %w", err)
	}

	// Validate mode-specific requirements
	switch mode {
	case constants.ModePublic:
		return validatePublicModeBasic(payload)
	case constants.ModePrivate:
		return validatePrivateModeBasic(payload, evt.Tags)
	}

	return nil
}

// findTlockTag finds the first tlock tag in the event tags
func findTlockTag(tags nostr.Tags) nostr.Tag {
	for _, tag := range tags {
		if len(tag) > 0 && tag[0] == constants.TagTlock {
			return tag
		}
	}
	return nil
}

// validateTlockTagBasic validates the tlock tag format (minimal validation)
func validateTlockTagBasic(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("tlock tag too short")
	}

	kv := parseTlockTagPairs(tag)
	
	// Just check that required keys exist - don't validate values deeply
	if _, exists := kv["drand_chain"]; !exists {
		return fmt.Errorf("missing drand_chain")
	}
	
	if _, exists := kv["drand_round"]; !exists {
		return fmt.Errorf("missing drand_round")
	}

	return nil
}

// parseTlockTagPairs parses "key value" string pairs from tlock tag
func parseTlockTagPairs(tag nostr.Tag) map[string]string {
	kv := make(map[string]string)
	
	for i := 1; i < len(tag); i++ {
		s := strings.TrimSpace(tag[i])
		spaceIdx := strings.Index(s, " ")
		if spaceIdx <= 0 {
			continue // skip malformed
		}
		
		key := strings.ToLower(s[:spaceIdx])
		value := s[spaceIdx+1:]
		kv[key] = value // last occurrence wins
	}
	
	return kv
}

// validatePublicModeBasic validates public mode payload format (basic structure only)
func validatePublicModeBasic(payload []byte) error {
	if len(payload) < 2 {
		return fmt.Errorf(constants.ErrMalformedPayload)
	}

	tlockBlob := payload[1:]
	if len(tlockBlob) > constants.MaxTlockBlobSize {
		return fmt.Errorf(constants.ErrTlockBlobTooLarge)
	}

	return nil
}

// validatePrivateModeBasic validates private mode payload format (basic structure only)
func validatePrivateModeBasic(payload []byte, tags nostr.Tags) error {
	// Check for recipient tag
	if !hasRecipientTag(tags) {
		return fmt.Errorf(constants.ErrMissingRecipientTag)
	}

	// Validate payload structure: 0x02 || nonce(12) || be32(tlock_len) || tlock_blob || ciphertext || mac(32)
	if len(payload) < 1+constants.MaxNonceSize+4+1+constants.HMACSize {
		return fmt.Errorf(constants.ErrMalformedPayload)
	}

	offset := 1 // Skip mode byte
	
	// Nonce (12 bytes)
	if len(payload) < offset+constants.MaxNonceSize {
		return fmt.Errorf(constants.ErrMalformedPayload)
	}
	offset += constants.MaxNonceSize

	// tlock_len (4 bytes big-endian)
	if len(payload) < offset+4 {
		return fmt.Errorf(constants.ErrMalformedPayload)
	}
	
	tlockLen := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4

	// Basic length validation
	if tlockLen > constants.MaxTlockBlobSize {
		return fmt.Errorf(constants.ErrTlockBlobTooLarge)
	}

	if len(payload) < offset+int(tlockLen)+constants.HMACSize {
		return fmt.Errorf(constants.ErrMalformedPayload)
	}

	// Total size check
	if len(payload) > constants.MaxContentSize {
		return fmt.Errorf(constants.ErrContentTooLarge)
	}

	return nil
}

// hasRecipientTag checks if the event has a recipient (p) tag
func hasRecipientTag(tags nostr.Tags) bool {
	count := 0
	for _, tag := range tags {
		if len(tag) >= 2 && tag[0] == constants.TagP {
			count++
			if count > constants.MaxPTags {
				return false // Too many p tags
			}
		}
	}
	return count > 0
}

// Helper functions for clients (optional to use)

// GetTlockKV extracts a specific key-value from tlock tag
func GetTlockKV(tag nostr.Tag, key string) string {
	kv := parseTlockTagPairs(tag)
	return kv[strings.ToLower(key)]
}

// ExtractDrandParameters extracts drand chain hash and round from tlock tag
func ExtractDrandParameters(evt *nostr.Event) (chainHash string, round int64, err error) {
	tlockTag := findTlockTag(evt.Tags)
	if tlockTag == nil {
		return "", 0, fmt.Errorf(constants.ErrMissingTlockTag)
	}

	kv := parseTlockTagPairs(tlockTag)
	
	chainHash = kv["drand_chain"]
	if chainHash == "" {
		return "", 0, fmt.Errorf("missing drand_chain")
	}

	roundStr := kv["drand_round"]
	if roundStr == "" {
		return "", 0, fmt.Errorf("missing drand_round")
	}

	round, err = strconv.ParseInt(roundStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid drand_round: %w", err)
	}

	return chainHash, round, nil
}

// GetPayloadMode extracts the mode byte from the content
func GetPayloadMode(evt *nostr.Event) (byte, error) {
	payload, err := base64.StdEncoding.DecodeString(evt.Content)
	if err != nil {
		return 0, fmt.Errorf("invalid base64 content: %w", err)
	}

	if len(payload) < 1 {
		return 0, fmt.Errorf(constants.ErrMalformedPayload)
	}

	return payload[0], nil
}

// ParsePrivatePayload parses a private mode payload
func ParsePrivatePayload(payload []byte) (nonce []byte, tlockBlob []byte, ciphertext []byte, mac []byte, err error) {
	if len(payload) < 1 {
		return nil, nil, nil, nil, fmt.Errorf(constants.ErrMalformedPayload)
	}

	if payload[0] != constants.ModePrivate {
		return nil, nil, nil, nil, fmt.Errorf("not a private mode payload")
	}

	offset := 1 // Skip mode byte

	// Extract nonce (12 bytes)
	if len(payload) < offset+constants.MaxNonceSize {
		return nil, nil, nil, nil, fmt.Errorf(constants.ErrMalformedPayload)
	}
	nonce = payload[offset : offset+constants.MaxNonceSize]
	offset += constants.MaxNonceSize

	// Extract tlock_len (4 bytes big-endian)
	if len(payload) < offset+4 {
		return nil, nil, nil, nil, fmt.Errorf(constants.ErrMalformedPayload)
	}
	tlockLen := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4

	// Extract tlock_blob
	if len(payload) < offset+int(tlockLen) {
		return nil, nil, nil, nil, fmt.Errorf(constants.ErrMalformedPayload)
	}
	tlockBlob = payload[offset : offset+int(tlockLen)]
	offset += int(tlockLen)

	// Extract ciphertext (everything except last 32 bytes for MAC)
	if len(payload) < offset+constants.HMACSize {
		return nil, nil, nil, nil, fmt.Errorf(constants.ErrMalformedPayload)
	}
	ciphertext = payload[offset : len(payload)-constants.HMACSize]
	
	// Extract MAC (last 32 bytes)
	mac = payload[len(payload)-constants.HMACSize:]

	return nonce, tlockBlob, ciphertext, mac, nil
}

// GetFirstRecipientPubkey extracts the first recipient pubkey from p tags
func GetFirstRecipientPubkey(tags nostr.Tags) string {
	for _, tag := range tags {
		if len(tag) >= 2 && tag[0] == constants.TagP {
			return tag[1]
		}
	}
	return ""
}
