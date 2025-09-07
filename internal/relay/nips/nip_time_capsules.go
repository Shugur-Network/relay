package nips

// This file provides compatibility layer for the NIP-XX time capsule implementation
// It replaces the old witness-based time capsule system with drand-based time-lock encryption

import (
	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// ValidateTimeCapsuleEvent is now handled by the NIP-XX implementation
// in nip_xx_time_capsules.go - this file just provides compatibility

// IsTimeCapsuleEvent checks if an event is a NIP-XX time capsule
func IsTimeCapsuleEvent(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsule
}

// Legacy function compatibility - these are no longer used in NIP-XX
// but kept for any existing code that might reference them

// IsTimeCapsuleUnlockShare - deprecated, not used in NIP-XX
func IsTimeCapsuleUnlockShare(evt *nostr.Event) bool {
	return false // NIP-XX doesn't use unlock shares
}

// IsTimeCapsuleShareDistribution - deprecated, not used in NIP-XX  
func IsTimeCapsuleShareDistribution(evt *nostr.Event) bool {
	return false // NIP-XX doesn't use share distribution
}

// ValidateTimeCapsuleUnlockShare - deprecated, not used in NIP-XX
func ValidateTimeCapsuleUnlockShare(evt *nostr.Event) error {
	return nil // No-op for NIP-XX
}

// ValidateTimeCapsuleShareDistribution - deprecated, not used in NIP-XX
func ValidateTimeCapsuleShareDistribution(evt *nostr.Event) error {
	return nil // No-op for NIP-XX
}
