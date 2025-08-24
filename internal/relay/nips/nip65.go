package nips

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-65: Relay List Metadata
// https://github.com/nostr-protocol/nips/blob/master/65.md

const (
	// KindRelayList is the event kind for relay lists
	KindRelayList = 10002
)

// ValidateKind10002 validates a kind 10002 relay list metadata event according to NIP-65
func ValidateKind10002(evt nostr.Event) error {
	if evt.Kind != KindRelayList {
		return fmt.Errorf("invalid event kind: expected %d, got %d", KindRelayList, evt.Kind)
	}

	// According to NIP-65:
	// - kind 10002 events SHOULD have empty content
	// - They contain "r" tags with relay URLs and optional markers

	// Content should be empty (though we allow it for forward compatibility)
	if strings.TrimSpace(evt.Content) != "" {
		// This is a SHOULD requirement, so we'll log but not reject
		// In the future, we might want to make this more strict
	}

	// Validate all r tags
	for _, tag := range evt.Tags {
		if len(tag) == 0 {
			continue
		}
		
		if tag[0] == "r" {
			if err := validateRelayTag(tag); err != nil {
				return fmt.Errorf("invalid r tag: %w", err)
			}
		}
	}

	return nil
}

// validateRelayTag validates an individual "r" tag
func validateRelayTag(tag []string) error {
	if len(tag) < 2 {
		return fmt.Errorf("r tag must have at least 2 elements: ['r', 'relay_url']")
	}

	relayURL := tag[1]
	
	// Validate URL format
	if err := validateRelayURL(relayURL); err != nil {
		return fmt.Errorf("invalid relay URL '%s': %w", relayURL, err)
	}

	// If there's a third element, it should be a valid marker
	if len(tag) >= 3 {
		marker := tag[2]
		if err := validateRelayMarker(marker); err != nil {
			return fmt.Errorf("invalid relay marker '%s': %w", marker, err)
		}
	}

	return nil
}

// validateRelayURL validates that a URL is a valid WebSocket relay URL
func validateRelayURL(rawURL string) error {
	// Parse the URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	// Check scheme - should be ws or wss
	switch u.Scheme {
	case "ws", "wss":
		// Valid WebSocket schemes
	default:
		return fmt.Errorf("invalid scheme '%s', expected 'ws' or 'wss'", u.Scheme)
	}

	// Check that host is present
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}

	// Basic hostname validation
	if err := validateHostname(u.Host); err != nil {
		return fmt.Errorf("invalid host: %w", err)
	}

	return nil
}

// validateHostname performs basic hostname validation
func validateHostname(host string) error {
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Basic checks
	if host == "" {
		return fmt.Errorf("empty hostname")
	}

	if len(host) > 253 {
		return fmt.Errorf("hostname too long")
	}

	// Check for valid hostname characters
	// Allow alphanumeric, hyphens, dots, and underscores
	validHostname := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validHostname.MatchString(host) {
		return fmt.Errorf("invalid hostname characters")
	}

	return nil
}

// validateRelayMarker validates the optional marker in r tags
func validateRelayMarker(marker string) error {
	// According to NIP-65, valid markers are "read" and "write"
	// If no marker is present, the relay is used for both read and write
	switch marker {
	case "read", "write":
		return nil
	default:
		return fmt.Errorf("invalid marker '%s', expected 'read' or 'write'", marker)
	}
}

// ExtractRelayList extracts relay information from a kind 10002 event
func ExtractRelayList(evt nostr.Event) map[string]string {
	relays := make(map[string]string)
	
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "r" {
			relayURL := tag[1]
			marker := "read,write" // default if no marker specified
			
			if len(tag) >= 3 {
				marker = tag[2]
			}
			
			relays[relayURL] = marker
		}
	}
	
	return relays
}

// RelayListEntry represents a single relay entry in a relay list
type RelayListEntry struct {
	URL    string `json:"url"`
	Marker string `json:"marker,omitempty"` // "read", "write", or empty for both
}

// ValidateRelayListFilter validates a filter for relay list events
func ValidateRelayListFilter(f nostr.Filter) error {
	// Allow filters that include kind 10002
	if f.Kinds != nil {
		hasKind10002 := false
		for _, kind := range f.Kinds {
			if kind == KindRelayList {
				hasKind10002 = true
				break
			}
		}
		if !hasKind10002 {
			return fmt.Errorf("filter must include kind %d for relay lists", KindRelayList)
		}
	}

	// Validate authors if present
	if len(f.Authors) > 0 {
		for _, author := range f.Authors {
			if !nostr.IsValid32ByteHex(author) {
				return fmt.Errorf("invalid author pubkey: %s", author)
			}
		}
	}

	return nil
}
