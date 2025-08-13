package nips

import (
	"strconv"
	"time"

	nostr "github.com/nbd-wtf/go-nostr"
)

// GetExpirationTime extracts the expiration timestamp from an event
// Returns the expiration time and true if found, or zero time and false if not found
func GetExpirationTime(evt nostr.Event) (time.Time, bool) {
	for _, t := range evt.Tags {
		if len(t) >= 2 && t[0] == "expiration" {
			if timestamp, err := strconv.ParseInt(t[1], 10, 64); err == nil {
				return time.Unix(timestamp, 0), true
			}
		}
	}
	return time.Time{}, false
}

// IsExpired checks if an event has expired based on its expiration tag
func IsExpired(evt nostr.Event) bool {
	if expTime, hasExpiration := GetExpirationTime(evt); hasExpiration {
		return time.Now().After(expTime)
	}
	return false
}

// ShouldAcceptExpiredEvent checks if a relay should accept an expired event
// According to NIP-40, relays SHOULD drop expired events that are published
func ShouldAcceptExpiredEvent(evt nostr.Event) bool {
	return !IsExpired(evt)
}

// ValidateExpirationTag validates the expiration tag format
func ValidateExpirationTag(evt nostr.Event) error {
	for _, t := range evt.Tags {
		if len(t) >= 2 && t[0] == "expiration" {
			if _, err := strconv.ParseInt(t[1], 10, 64); err != nil {
				return err
			}
		}
	}
	return nil
}
