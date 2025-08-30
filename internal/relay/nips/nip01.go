package nips

import (
	"encoding/json"
	"fmt"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// VerifyEventJSON parses the raw JSON into a go-nostr Event
// and checks the BIP-340 signature. It returns nil if valid, or an error otherwise.
func VerifyEventJSON(rawEvent []byte) error {
	// Unmarshal into the standard Nostr event struct
	var evt nostr.Event
	if err := json.Unmarshal(rawEvent, &evt); err != nil {
		logger.Debug("NIP-01: Failed to parse JSON into nostr.Event",
			zap.Error(err),
			zap.ByteString("raw_event", rawEvent))
		return fmt.Errorf("failed to parse JSON into nostr.Event: %w", err)
	}

	// CheckSignature() does:
	//   1) Recompute event ID from [0, pubkey, created_at, kind, tags, content]
	//   2) Parse x-only pubkey
	//   3) Parse 64-byte Schnorr signature
	//   4) BIP-340 verification
	_, err := evt.CheckSignature()
	if err != nil {
		logger.Warn("NIP-01: Signature verification failed",
			zap.String("event_id", evt.ID),
			zap.String("pubkey", evt.PubKey),
			zap.Int("kind", evt.Kind),
			zap.Error(err))
		return fmt.Errorf("signature check failed: %w", err)
	}

	logger.Debug("NIP-01: Event signature verified successfully",
		zap.String("event_id", evt.ID),
		zap.String("pubkey", evt.PubKey),
		zap.Int("kind", evt.Kind))
	return nil
}

func IsAddressable(evt nostr.Event) bool {
	return evt.Kind >= 30000 && evt.Kind < 40000 && GetTagValue(evt, "d") != ""
}

// GetTagValue returns the first t[1] found for the given key, or "" if not found
func GetTagValue(evt nostr.Event, key string) string {
	for _, t := range evt.Tags {
		if len(t) >= 2 && t[0] == key {
			return t[1]
		}
	}
	return ""
}
