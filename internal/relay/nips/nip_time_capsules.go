package nips

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateTimeCapsuleEvent validates time capsule events (kinds 11990, 30095)
func ValidateTimeCapsuleEvent(evt *nostr.Event) error {
	// Must have vendor tag
	if !hasVendorTag(evt) {
		return fmt.Errorf("missing vendor tag [\"x-cap\", \"v1\"]")
	}

	// Extract unlock configuration
	unlockConfig, err := extractUnlockConfig(evt)
	if err != nil {
		return fmt.Errorf("invalid unlock config: %w", err)
	}

	// Validate unlock time (allow some tolerance for testing and clock skew)
	tolerance := 60 * time.Second // Allow 1 minute tolerance for clock skew and testing
	if unlockConfig.UnlockTime.Before(time.Now().Add(-tolerance)) {
		return fmt.Errorf("unlock time is too far in the past")
	}

	// Validate threshold mode parameters
	if unlockConfig.Mode == constants.ModeThreshold {
		if unlockConfig.Threshold < 1 || unlockConfig.WitnessCount < unlockConfig.Threshold {
			return fmt.Errorf("invalid threshold configuration: t=%d, n=%d", 
				unlockConfig.Threshold, unlockConfig.WitnessCount)
		}

		// Enforce maximum witness count
		if unlockConfig.WitnessCount > constants.MaxWitnessCount {
			return fmt.Errorf("witness count exceeds maximum: %d > %d", 
				unlockConfig.WitnessCount, constants.MaxWitnessCount)
		}

		// Must have witness list
		witnesses := extractWitnesses(evt)
		if len(witnesses) == 0 {
			return fmt.Errorf("missing witnesses")
		}
		if len(witnesses) != unlockConfig.WitnessCount {
			return fmt.Errorf("witness count mismatch: expected %d, got %d", 
				unlockConfig.WitnessCount, len(witnesses))
		}
	}

	// Must have commitment (w-commit tag)
	if !hasCommitment(evt) {
		return fmt.Errorf("missing commitment")
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

// ValidateTimeCapsuleUnlockShare validates unlock share events (kind 11991)
func ValidateTimeCapsuleUnlockShare(evt *nostr.Event) error {
	// Must have vendor tag
	if !hasVendorTag(evt) {
		return fmt.Errorf("missing vendor tag [\"x-cap\", \"v1\"]")
	}

	// Must reference a capsule event
	capsuleID := extractCapsuleReference(evt)
	if capsuleID == "" {
		return fmt.Errorf("missing capsule reference in 'e' tag")
	}

	// Must specify witness
	witness := extractWitnessFromShare(evt)
	if witness == "" {
		return fmt.Errorf("missing witness in 'w' tag")
	}

	// Extract and validate unlock time
	unlockTime, err := extractUnlockTimeFromShare(evt)
	if err != nil {
		return fmt.Errorf("invalid unlock time: %w", err)
	}

	// Share can only be posted at or after unlock time (with small tolerance)
	tolerance := 5 * time.Minute // Allow 5 minutes early
	if time.Now().Before(unlockTime.Add(-tolerance)) {
		return fmt.Errorf("share posted too early")
	}

	return nil
}

// IsTimeCapsuleEvent checks if an event is a time capsule
func IsTimeCapsuleEvent(evt *nostr.Event) bool {
	return (evt.Kind == constants.KindTimeCapsule || 
			evt.Kind == constants.KindTimeCapsuleReplaceable) && 
		   hasVendorTag(evt)
}

// IsTimeCapsuleUnlockShare checks if an event is an unlock share
func IsTimeCapsuleUnlockShare(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsuleUnlockShare && hasVendorTag(evt)
}

// Helper types
type UnlockConfig struct {
	Mode         string
	Threshold    int
	WitnessCount int
	UnlockTime   time.Time
}

// Helper functions

func hasVendorTag(evt *nostr.Event) bool {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagXCap && tag[1] == "v1" {
			return true
		}
	}
	return false
}

func extractUnlockConfig(evt *nostr.Event) (*UnlockConfig, error) {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == constants.TagUnlock {
			// Handle both formats:
			// Format 1: ["u", "threshold", "t", "3", "n", "5", "T", "1735689600"]
			// Format 2: ["u", "threshold,t,3,n,3,T,1756311650"]
			
			var parts []string
			if len(tag) >= 8 {
				// Format 1: separate elements
				parts = tag[1:]
			} else if len(tag) == 2 {
				// Format 2: comma-separated string
				parts = strings.Split(tag[1], ",")
			} else {
				return nil, fmt.Errorf("invalid unlock tag format")
			}
			
			if len(parts) < 7 {
				return nil, fmt.Errorf("insufficient unlock config parameters")
			}
			
			mode := parts[0]
			if mode != constants.ModeThreshold {
				return nil, fmt.Errorf("unsupported mode: %s", mode)
			}

			// Parse threshold and witness count
			if parts[1] != "t" || parts[3] != "n" || parts[5] != "T" {
				return nil, fmt.Errorf("invalid unlock config format")
			}

			threshold, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid threshold: %s", parts[2])
			}

			witnessCount, err := strconv.Atoi(parts[4])
			if err != nil {
				return nil, fmt.Errorf("invalid witness count: %s", parts[4])
			}

			unlockTimestamp, err := strconv.ParseInt(parts[6], 10, 64)
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
	for _, tag := range evt.Tags {
		if len(tag) > 1 && tag[0] == constants.TagWitness {
			// Handle both formats:
			// Format 1: ["w", "npub1", "npub2", "npub3"]
			// Format 2: ["w", "npub1,npub2,npub3"]
			
			if len(tag) > 2 {
				// Format 1: multiple elements
				return tag[1:]
			} else if len(tag) == 2 {
				// Format 2: comma-separated string
				witnesses := strings.Split(tag[1], ",")
				// Trim whitespace from each witness
				for i, w := range witnesses {
					witnesses[i] = strings.TrimSpace(w)
				}
				return witnesses
			}
		}
	}
	return nil
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
		if len(tag) >= 2 && tag[0] == constants.TagWitness {
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

func hasCommitment(evt *nostr.Event) bool {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "w-commit" {
			return true
		}
	}
	return false
}

func hasEncryption(evt *nostr.Event) bool {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "enc" {
			return true
		}
	}
	return false
}

func validateEncryption(evt *nostr.Event) error {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "enc" {
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
		if len(tag) >= 2 && tag[0] == "loc" {
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
