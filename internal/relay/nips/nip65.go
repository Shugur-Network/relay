package nips

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

const (
	// KindRelayList is the event kind for relay lists
	KindRelayList = 10002
)

// RelayListEntry represents a single relay entry in a relay list
type RelayListEntry struct {
	URL       string `json:"url"`
	Read      bool   `json:"read,omitempty"`
	Write     bool   `json:"write,omitempty"`
	Advertise bool   `json:"advertise,omitempty"`
}

// ValidateRelayListFilter validates a filter for relay list events
func ValidateRelayListFilter(f nostr.Filter) error {
	if f.Kinds == nil || len(f.Kinds) != 1 || f.Kinds[0] != KindRelayList {
		return fmt.Errorf("filter must specify kind %d for relay lists", KindRelayList)
	}

	if len(f.Authors) > 0 {
		for _, author := range f.Authors {
			if !nostr.IsValid32ByteHex(author) {
				return fmt.Errorf("invalid author pubkey: %s", author)
			}
		}
	}

	return nil
}

// ValidateRelayListEvent validates a relay list event
func ValidateRelayListEvent(evt nostr.Event) error {
	if evt.Kind != KindRelayList {
		return fmt.Errorf("invalid event kind: expected %d, got %d", KindRelayList, evt.Kind)
	}

	// Parse the content as a map of relay URLs to their read/write status
	var relayList map[string]RelayListEntry
	if err := json.Unmarshal([]byte(evt.Content), &relayList); err != nil {
		return fmt.Errorf("invalid relay list content: %v", err)
	}

	// Validate each relay URL
	for url, entry := range relayList {
		if !strings.HasPrefix(url, "wss://") && !strings.HasPrefix(url, "ws://") {
			return fmt.Errorf("invalid relay URL: %s", url)
		}

		// At least one of read or write should be true
		if !entry.Read && !entry.Write {
			return fmt.Errorf("relay %s must have at least one of read or write set to true", url)
		}
	}

	return nil
}

// ParseRelayList parses a relay list event into a map of relay URLs to their read/write status
func ParseRelayList(evt nostr.Event) (map[string]RelayListEntry, error) {
	if err := ValidateRelayListEvent(evt); err != nil {
		return nil, err
	}

	var relayList map[string]RelayListEntry
	if err := json.Unmarshal([]byte(evt.Content), &relayList); err != nil {
		return nil, fmt.Errorf("failed to parse relay list: %v", err)
	}

	return relayList, nil
}

// RelayList represents a list of relays with their read/write permissions
type RelayList struct {
	Relays map[string]RelayPermissions `json:"relays"`
}

// RelayPermissions defines read/write permissions for a relay
type RelayPermissions struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

// GetRelayPermissions returns the read/write permissions for a specific relay
func (rl *RelayList) GetRelayPermissions(relayURL string) (RelayPermissions, bool) {
	perms, exists := rl.Relays[relayURL]
	return perms, exists
}

// AddRelay adds or updates a relay's permissions
func (rl *RelayList) AddRelay(relayURL string, read, write bool) error {
	if !isValidRelayURL(relayURL) {
		return fmt.Errorf("invalid relay URL: %s", relayURL)
	}

	if rl.Relays == nil {
		rl.Relays = make(map[string]RelayPermissions)
	}

	rl.Relays[relayURL] = RelayPermissions{
		Read:  read,
		Write: write,
	}

	return nil
}

// RemoveRelay removes a relay from the list
func (rl *RelayList) RemoveRelay(relayURL string) {
	delete(rl.Relays, relayURL)
}

// ToEvent converts a RelayList to a nostr.Event
func (rl *RelayList) ToEvent(pubkey string) (*nostr.Event, error) {
	content, err := json.Marshal(rl)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal relay list: %w", err)
	}

	evt := nostr.Event{
		Kind:      10002,
		PubKey:    pubkey,
		Content:   string(content),
		CreatedAt: nostr.Now(),
	}

	// Add tags for relay URLs
	for url := range rl.Relays {
		evt.Tags = append(evt.Tags, nostr.Tag{"r", url})
	}

	return &evt, nil
}

// isValidRelayURL checks if a URL is a valid relay URL
func isValidRelayURL(url string) bool {
	if !strings.HasPrefix(url, "wss://") && !strings.HasPrefix(url, "ws://") {
		return false
	}

	// Basic URL validation
	if len(url) < 8 || len(url) > 200 {
		return false
	}

	// Check for valid characters
	for _, c := range url {
		if !isValidURLChar(c) {
			return false
		}
	}

	return true
}

// isValidURLChar checks if a character is valid in a URL
func isValidURLChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '.' || c == '-' || c == '_' || c == ':' || c == '/' || c == '?'
}
