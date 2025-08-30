package nips

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateTimeCapsuleEvent validates time capsule events (kinds 1990, 30095)
func ValidateTimeCapsuleEvent(evt *nostr.Event) error {
	// Extract unlock configuration
	unlockConfig, err := extractUnlockConfig(evt)
	if err != nil {
		return fmt.Errorf("invalid unlock config: %w", err)
	}

	// Validate mode-specific parameters
	switch unlockConfig.Mode {
	case constants.ModeThreshold:
		if unlockConfig.Threshold < 1 || unlockConfig.WitnessCount < unlockConfig.Threshold {
			return fmt.Errorf("invalid threshold configuration: t=%d, n=%d",
				unlockConfig.Threshold, unlockConfig.WitnessCount)
		}

		// Enforce maximum witness count
		if unlockConfig.WitnessCount > constants.MaxWitnessCount {
			return fmt.Errorf("witness count exceeds maximum: %d > %d",
				unlockConfig.WitnessCount, constants.MaxWitnessCount)
		}

		// Must have witness list (using 'p' tags as per NIP spec)
		witnesses := extractWitnesses(evt)
		if len(witnesses) == 0 {
			return fmt.Errorf("missing witnesses")
		}
		if len(witnesses) != unlockConfig.WitnessCount {
			return fmt.Errorf("witness count mismatch: expected %d, got %d",
				unlockConfig.WitnessCount, len(witnesses))
		}

		// Must have commitment (w-commit tag)
		if !hasCommitment(evt) {
			return fmt.Errorf("missing witness commitment")
		}
	case constants.ModeScheduled:
		// Scheduled mode doesn't require witnesses or commitments
		// Just validate the time is valid
		if unlockConfig.UnlockTime.IsZero() {
			return fmt.Errorf("invalid unlock time for scheduled mode")
		}
	}

	// Must have valid encryption info (enc tag)
	if err := validateEncryption(evt); err != nil {
		return fmt.Errorf("encryption validation failed: %w", err)
	}

	// Must have location info (loc tag)
	if !hasLocation(evt) {
		return fmt.Errorf("missing location info")
	}

	// Validate parameterized replaceable events (kind 30095)
	if evt.Kind == constants.KindTimeCapsuleReplaceable {
		if !hasDTag(evt) {
			return fmt.Errorf("missing 'd' tag for parameterized replaceable event")
		}
	}

	return nil
}

// ValidateTimeCapsuleUnlockShare validates unlock share events (kind 1991)
func ValidateTimeCapsuleUnlockShare(evt *nostr.Event) error {
	// Must reference a capsule event
	capsuleID := extractCapsuleReference(evt)
	if capsuleID == "" {
		return fmt.Errorf("missing capsule reference in 'e' tag")
	}

	// Must specify witness (using 'p' tag as per NIP spec)
	witness := extractWitnessFromShare(evt)
	if witness == "" {
		return fmt.Errorf("missing witness in 'p' tag")
	}

	// Extract and validate unlock time
	unlockTime, err := extractUnlockTimeFromShare(evt)
	if err != nil {
		return fmt.Errorf("invalid unlock time: %w", err)
	}

	// Share can only be posted at or after unlock time (with tolerance per NIP spec)
	tolerance := 5 * time.Minute // Allow 5 minutes early (300 seconds per spec)
	if time.Now().Before(unlockTime.Add(-tolerance)) {
		return fmt.Errorf("share posted too early")
	}

	return nil
}

// ValidateTimeCapsuleShareDistribution validates share distribution events (kind 1992)
func ValidateTimeCapsuleShareDistribution(evt *nostr.Event) error {
	// Must reference a capsule event
	capsuleID := extractCapsuleReference(evt)
	if capsuleID == "" {
		return fmt.Errorf("missing capsule reference in 'e' tag")
	}

	// Must specify recipient witness (using 'p' tag as per NIP spec)
	witness := extractWitnessFromShare(evt)
	if witness == "" {
		return fmt.Errorf("missing recipient witness in 'p' tag")
	}

	// Must have share index
	shareIdx := extractShareIndex(evt)
	if shareIdx < 0 {
		return fmt.Errorf("missing or invalid share index")
	}

	// Must have encryption info for NIP-44 v2
	if err := validateEncryption(evt); err != nil {
		return fmt.Errorf("encryption validation failed: %w", err)
	}

	// Content should be non-empty (encrypted share)
	if evt.Content == "" {
		return fmt.Errorf("missing encrypted share content")
	}

	return nil
}

// IsTimeCapsuleEvent checks if an event is a time capsule
func IsTimeCapsuleEvent(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsule ||
		evt.Kind == constants.KindTimeCapsuleReplaceable
}

// IsTimeCapsuleUnlockShare checks if an event is an unlock share
func IsTimeCapsuleUnlockShare(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsuleUnlockShare
}

// IsTimeCapsuleShareDistribution checks if an event is a share distribution
func IsTimeCapsuleShareDistribution(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsuleShareDistribution
}

// Helper types
type UnlockConfig struct {
	Mode         string
	Threshold    int
	WitnessCount int
	UnlockTime   time.Time
}

// Helper functions

func extractUnlockConfig(evt *nostr.Event) (*UnlockConfig, error) {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagUnlock {
			// NIP format: ["u","threshold","t","<t>","n","<n>","T","<unix>"] for threshold
			// NIP format: ["u","scheduled","T","<unix>"] for scheduled
			mode := tag[1]
			if mode != constants.ModeThreshold && mode != constants.ModeScheduled {
				return nil, fmt.Errorf("unsupported mode: %s", mode)
			}

			// For scheduled mode, only need unlock time
			if mode == constants.ModeScheduled {
				if len(tag) < 4 || tag[2] != "T" {
					return nil, fmt.Errorf("invalid scheduled mode format")
				}
				unlockTimestamp, err := strconv.ParseInt(tag[3], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid unlock timestamp: %s", tag[3])
				}
				return &UnlockConfig{
					Mode:         mode,
					Threshold:    0,
					WitnessCount: 0,
					UnlockTime:   time.Unix(unlockTimestamp, 0),
				}, nil
			}

			// For threshold mode, need all parameters
			if len(tag) < 8 || tag[2] != "t" || tag[4] != "n" || tag[6] != "T" {
				return nil, fmt.Errorf("invalid threshold mode format")
			}

			threshold, err := strconv.Atoi(tag[3])
			if err != nil {
				return nil, fmt.Errorf("invalid threshold: %s", tag[3])
			}

			witnessCount, err := strconv.Atoi(tag[5])
			if err != nil {
				return nil, fmt.Errorf("invalid witness count: %s", tag[5])
			}

			unlockTimestamp, err := strconv.ParseInt(tag[7], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid unlock timestamp: %s", tag[7])
			}

			return &UnlockConfig{
				Mode:         mode,
				Threshold:    threshold,
				WitnessCount: witnessCount,
				UnlockTime:   time.Unix(unlockTimestamp, 0),
			}, nil
		}
	}
	return nil, fmt.Errorf("missing unlock configuration")
}

func extractWitnesses(evt *nostr.Event) []string {
	var witnesses []string

	// Use 'p' tags for witnesses as per NIP specification
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			witnesses = append(witnesses, tag[1])
		}
	}

	return witnesses
}

func extractCapsuleReference(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			return tag[1]
		}
	}
	return ""
}

func extractWitnessFromShare(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			return tag[1]
		}
	}
	return ""
}

func extractUnlockTimeFromShare(evt *nostr.Event) (time.Time, error) {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "T" {
			timestamp, err := strconv.ParseInt(tag[1], 10, 64)
			if err != nil {
				return time.Time{}, err
			}
			return time.Unix(timestamp, 0), nil
		}
	}
	return time.Time{}, fmt.Errorf("missing unlock time")
}

func extractShareIndex(evt *nostr.Event) int {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagShareIndex {
			if idx, err := strconv.Atoi(tag[1]); err == nil {
				return idx
			}
		}
	}
	return -1
}

func hasCommitment(evt *nostr.Event) bool {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagWitnessCommit {
			return true
		}
	}
	return false
}

func validateEncryption(evt *nostr.Event) error {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagEncryption {
			encType := tag[1]
			// Validate encryption format
			if !isValidEncryptionFormat(encType) {
				return fmt.Errorf("invalid encryption format: %s", encType)
			}
			return nil
		}
	}
	return fmt.Errorf("missing encryption info")
}

func isValidEncryptionFormat(encType string) bool {
	validFormats := []string{
		"nip44:v2",
		"nip44:v1", // Legacy support
	}

	for _, format := range validFormats {
		if encType == format {
			return true
		}
	}
	return false
}

func hasLocation(evt *nostr.Event) bool {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagLocation {
			return true
		}
	}
	return false
}

func hasDTag(evt *nostr.Event) bool {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			return true
		}
	}
	return false
}
