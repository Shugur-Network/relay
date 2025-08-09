package nips

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// NIP-16: Event Treatment
// https://github.com/nostr-protocol/nips/blob/master/16.md

func IsEphemeral(kind int) bool {
	// According to NIP-16, ephemeral events are 20000 <= kind < 30000
	return kind >= 20000 && kind < 30000
}

func IsReplaceable(kind int) bool {
	// NIP-16 calls certain kinds "replaceable." Commonly these are 0, 3, 41...
	// You can define a set or switch statement:
	switch kind {
	case 0, 3, 41:
		return true
	}
	return false
}

// ValidateEventTreatment validates event according to NIP-16 treatment rules
func ValidateEventTreatment(evt *nostr.Event) error {
	// For parameterized replaceable events, ensure they have a 'd' tag
	if IsParameterizedReplaceableKind(evt.Kind) {
		hasDTag := false
		for _, tag := range evt.Tags {
			if len(tag) >= 2 && tag[0] == "d" {
				hasDTag = true
				break
			}
		}
		if !hasDTag {
			return fmt.Errorf("parameterized replaceable event must have 'd' tag")
		}
	}

	// Ephemeral events are generally accepted but not stored
	// (This is handled at the storage layer, not validation)

	// Replaceable events can replace previous events of the same kind from the same author
	// (This is also handled at the storage layer)

	return nil
}
