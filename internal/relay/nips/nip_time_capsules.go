package nips

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateTimeCapsuleEvent validates time capsule events (kinds 11990, 31990)
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

	// Validate unlock time (must be in future)
	if unlockConfig.UnlockTime.Before(time.Now()) {
		return fmt.Errorf("unlock time must be in the future")
	}

	// Validate threshold mode parameters
	if unlockConfig.Mode == constants.ModeThreshold {
		if unlockConfig.Threshold < 1 || unlockConfig.WitnessCount < unlockConfig.Threshold {
			return fmt.Errorf("invalid threshold configuration: t=%d, n=%d", 
				unlockConfig.Threshold, unlockConfig.WitnessCount)
		}

		// Must have witness list
		witnesses := extractWitnesses(evt)
		if len(witnesses) != unlockConfig.WitnessCount {
			return fmt.Errorf("witness count mismatch: expected %d, got %d", 
				unlockConfig.WitnessCount, len(witnesses))
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
		if len(tag) >= 8 && tag[0] == constants.TagUnlock {
			mode := tag[1]
			if mode != constants.ModeThreshold {
				return nil, fmt.Errorf("unsupported mode: %s", mode)
			}

			// Parse ["u", "threshold", "t", "3", "n", "5", "T", "1735689600"]
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
	for _, tag := range evt.Tags {
		if len(tag) > 1 && tag[0] == constants.TagWitness {
			return tag[1:] // Return all witness npubs
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
