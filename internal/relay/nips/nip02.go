package nips

import (
	"github.com/Shugur-Network/relay/internal/relay/nips/common"
	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-02: Follow List
// https://github.com/nostr-protocol/nips/blob/master/02.md

// ValidateFollowList validates NIP-02 follow list events (kind 3)
func ValidateFollowList(evt *nostr.Event) error {
	return common.ValidateEventWithCallback(
		evt,
		"02",          // NIP number
		3,             // Expected event kind
		"follow list", // Event name for logging
		func(helper *common.ValidationHelper, event *nostr.Event) error {
			// Follow lists can have any tags structure, most commonly "p" tags for pubkeys
			// Validate pubkey format in any "p" tags if they exist
			for _, tag := range event.Tags {
				if len(tag) >= 2 && tag[0] == "p" {
					if err := helper.ValidatePubkey(tag[1]); err != nil {
						return helper.FormatTagError("p", "invalid pubkey: %v", err)
					}
				}
			}

			// No strict validation needed as the format is flexible
			return nil
		},
	)
}

// IsFollowListEvent checks if an event is a follow list
func IsFollowListEvent(evt *nostr.Event) bool {
	return evt.Kind == 3
}
