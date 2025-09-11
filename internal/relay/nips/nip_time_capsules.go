package nips

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"

	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateTimeCapsuleEvent validates time capsule events according to NIP-XX
// Public time capsules: content is Base64 of binary age v1 ciphertext with exactly one tlock recipient
// Private time capsules are delivered via NIP-59 (kinds 13, 1059) and validated separately
func ValidateTimeCapsuleEvent(evt *nostr.Event) error {
	// Must be kind 1041
	if evt.Kind != constants.KindTimeCapsule {
		return fmt.Errorf("invalid kind: expected %d, got %d", constants.KindTimeCapsule, evt.Kind)
	}

	// Validate content is valid base64
	_, err := base64.StdEncoding.DecodeString(evt.Content)
	if err != nil {
		return fmt.Errorf("%s: %w", constants.ErrInvalidBase64, err)
	}

	// Check content size limits (decoded)
	decodedSize := base64.StdEncoding.DecodedLen(len(evt.Content))
	if decodedSize > constants.MaxContentSize {
		return fmt.Errorf("%s: %d bytes exceeds %d limit", constants.ErrContentTooLarge, decodedSize, constants.MaxContentSize)
	}

	// Validate tlock tag
	tlockTag := findTlockTag(evt.Tags)
	if tlockTag == nil {
		return fmt.Errorf(constants.ErrMissingTlockTag)
	}

	if err := validateTlockTag(tlockTag); err != nil {
		return fmt.Errorf("invalid tlock tag: %w", err)
	}

	// For public capsules, inner 1041 MUST NOT contain p tags
	if hasRecipientTag(evt.Tags) {
		return fmt.Errorf("public time capsule must not contain p tags (use NIP-59 for private capsules)")
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

// validateTlockTag validates the new tlock tag format: ["tlock", "<drand_chain_hex64>", "<drand_round_uint>"]
func validateTlockTag(tag nostr.Tag) error {
	// Must be exactly 3 elements: ["tlock", "<drand_chain_hex64>", "<drand_round_uint>"]
	if len(tag) != 3 {
		return fmt.Errorf("%s: expected 3 elements, got %d", constants.ErrInvalidTlockFormat, len(tag))
	}

	// Validate drand chain hash format (64 lowercase hex characters)
	drandChain := tag[1]
	if err := validateDrandChainHash(drandChain); err != nil {
		return err
	}

	// Validate drand round format (positive integer)
	drandRoundStr := tag[2]
	if err := validateDrandRound(drandRoundStr); err != nil {
		return err
	}

	return nil
}

// validateDrandChainHash validates drand chain hash format (64 lowercase hex chars)
func validateDrandChainHash(chainHash string) error {
	if len(chainHash) != constants.DrandChainHashLength {
		return fmt.Errorf("%s: expected %d characters, got %d", constants.ErrInvalidDrandChain, constants.DrandChainHashLength, len(chainHash))
	}

	// Check if it's valid lowercase hex
	matched, _ := regexp.MatchString("^[0-9a-f]{64}$", chainHash)
	if !matched {
		return fmt.Errorf("%s: must be 64 lowercase hex characters", constants.ErrInvalidDrandChain)
	}

	return nil
}

// validateDrandRound validates drand round format (positive integer, 64-bit safe)
func validateDrandRound(roundStr string) error {
	// Check format: positive integer matching ^[1-9][0-9]{0,18}$
	matched, _ := regexp.MatchString("^[1-9][0-9]{0,18}$", roundStr)
	if !matched {
		return fmt.Errorf("%s: must be positive integer", constants.ErrInvalidDrandRound)
	}

	// Parse to ensure it's within int64 range
	round, err := strconv.ParseInt(roundStr, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: %w", constants.ErrInvalidDrandRound, err)
	}

	if round <= 0 || round > constants.MaxDrandRound {
		return fmt.Errorf("%s: round %d out of valid range", constants.ErrInvalidDrandRound, round)
	}

	return nil
}

// hasRecipientTag checks if the event has a recipient (p) tag
func hasRecipientTag(tags nostr.Tags) bool {
	for _, tag := range tags {
		if len(tag) >= 2 && tag[0] == constants.TagP {
			return true
		}
	}
	return false
}

// Helper functions for clients (optional to use)

// ExtractDrandParameters extracts drand chain hash and round from tlock tag (new format)
func ExtractDrandParameters(evt *nostr.Event) (chainHash string, round int64, err error) {
	tlockTag := findTlockTag(evt.Tags)
	if tlockTag == nil {
		return "", 0, fmt.Errorf(constants.ErrMissingTlockTag)
	}

	// New format: ["tlock", "<drand_chain_hex64>", "<drand_round_uint>"]
	if len(tlockTag) != 3 {
		return "", 0, fmt.Errorf("%s: expected 3 elements", constants.ErrInvalidTlockFormat)
	}

	chainHash = tlockTag[1]
	roundStr := tlockTag[2]

	round, err = strconv.ParseInt(roundStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid drand_round: %w", err)
	}

	return chainHash, round, nil
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
