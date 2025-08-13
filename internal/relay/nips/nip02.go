package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-02: Follow List
// https://github.com/nostr-protocol/nips/blob/master/02.md

// ValidateFollowList validates NIP-02 follow list events (kind 3)
func ValidateFollowList(evt *nostr.Event) error {
	if evt.Kind != 3 {
		return fmt.Errorf("invalid event kind for follow list: %d", evt.Kind)
	}

	// Follow lists can have any tags structure, most commonly "p" tags for pubkeys
	// No strict validation needed as the format is flexible

	return nil
}

// IsFollowListEvent checks if an event is a follow list
func IsFollowListEvent(evt *nostr.Event) bool {
	return evt.Kind == 3
}
