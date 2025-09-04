package nips

import (
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-16: Event Treatment
// https://github.com/nostr-protocol/nips/blob/master/16.md

func IsEphemeral(kind int) bool {
	// According to NIP-16, ephemeral events are 20000 <= kind < 30000
	return kind >= 20000 && kind < 30000
}

func IsReplaceable(kind int) bool {
	// According to NIP-01: events are replaceable for kind n such that:
	// 10000 <= n < 20000 || n == 0 || n == 3
	if kind >= 10000 && kind < 20000 {
		return true
	}
	switch kind {
	case 0, 3, 41: // 41 is replaceable per NIP-28: "Only the most recent kind 41 per e tag value MAY be available"
		return true
	}
	return false
}

// ValidateEventTreatment validates event according to NIP-16 treatment rules
func ValidateEventTreatment(evt *nostr.Event) error {
	// For addressable events, ensure they have a 'd' tag
	if IsParameterizedReplaceableKind(evt.Kind) {
		hasDTag := false
		for _, tag := range evt.Tags {
			if len(tag) >= 2 && tag[0] == "d" {
				hasDTag = true
				break
			}
		}
		if !hasDTag {
			return fmt.Errorf("addressable event must have 'd' tag")
		}
	}

	// Ephemeral events are generally accepted but not stored
	// (This is handled at the storage layer, not validation)

	// Replaceable events can replace previous events of the same kind from the same author
	// (This is also handled at the storage layer)

	return nil
}
