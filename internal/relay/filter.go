package relay

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// parseFilterFromRaw merges any "#p", "#e", etc. keys into Filter.Tags.
// Then Filter.Matches() can correctly check them.
func parseFilterFromRaw(raw interface{}) (nostr.Filter, error) {
	var f nostr.Filter

	// Step 1: Marshal the raw interface{} to JSON
	data, err := json.Marshal(raw)
	if err != nil {
		return f, fmt.Errorf("failed to encode filter: %w", err)
	}

	// Step 2: Unmarshal into standard Filter struct
	if err = json.Unmarshal(data, &f); err != nil {
		return f, fmt.Errorf("failed to decode filter: %w", err)
	}

	// Step 3: Unmarshal into a map to extract #tag fields
	var partial map[string]json.RawMessage
	if err = json.Unmarshal(data, &partial); err != nil {
		return f, fmt.Errorf("failed to decode partial: %w", err)
	}

	// Initialize Tags map if not already present
	if f.Tags == nil {
		f.Tags = make(map[string][]string)
	}

	// Step 4: Process each key that starts with "#" (e.g., "#p", "#e")
	for k, v := range partial {
		if len(k) > 1 && k[0] == '#' {
			tagKey := k[1:] // Extract tag name (e.g., "#p" -> "p")
			var arrVals []string
			if err2 := json.Unmarshal(v, &arrVals); err2 == nil {
				// Add tag values to the filter's Tags map
				f.Tags[tagKey] = arrVals
			}
		}
	}

	// Step 5: Apply filter normalization
	normalizeFilter(&f)

	return f, nil
}

// normalizeFilter applies normalization rules to ensure filter consistency
func normalizeFilter(f *nostr.Filter) {
	// Cap result limit to reasonable values
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 500
	}

	// Normalize IDs and Authors to lowercase if needed
	for i, id := range f.IDs {
		if len(id) < 64 {
			// Pad shorter IDs with prefix matching
			f.IDs[i] = id + strings.Repeat("0", 64-len(id))
		}
	}

	// Ensure search terms are properly formatted
	if f.Search != "" {
		f.Search = strings.TrimSpace(f.Search)
	}
}

// ValidateFilter ensures a filter is within safe limits to prevent DoS
// This is a lightweight validation that can be called before expensive operations
func ValidateFilter(f nostr.Filter) error {
	// Check for empty filter (no conditions)
	if len(f.IDs) == 0 &&
		len(f.Authors) == 0 &&
		len(f.Kinds) == 0 &&
		len(f.Tags) == 0 &&
		f.Since == nil &&
		f.Until == nil &&
		f.Search == "" {
		return fmt.Errorf("filter must have at least one condition")
	}

	// Validate IDs format
	for _, id := range f.IDs {
		if len(id) > 0 && len(id) <= 64 && !isHexString(id) {
			return fmt.Errorf("invalid ID format: %s", id)
		}
	}

	// Validate Authors format
	for _, author := range f.Authors {
		if len(author) > 0 && !nostr.IsValid32ByteHex(author) {
			return fmt.Errorf("invalid author pubkey: %s", author)
		}
	}

	// Validate kinds are within valid ranges
	for _, kind := range f.Kinds {
		if kind < 0 || kind > 40000 {
			return fmt.Errorf("invalid event kind: %d", kind)
		}
	}

	// Validate tag filters
	if len(f.Tags) > 10 {
		return fmt.Errorf("too many tag filters (max 10)")
	}

	// Check tag values length
	for tagName, values := range f.Tags {
		if len(values) > 20 {
			return fmt.Errorf("too many values for tag '%s' (max 20)", tagName)
		}
	}

	// Validate time range
	if f.Since != nil && f.Until != nil && f.Since.Time().Unix() > f.Until.Time().Unix() {
		return fmt.Errorf("invalid time range: 'since' is after 'until'")
	}

	// Validate search parameter if present
	if f.Search != "" {
		if len(strings.Fields(f.Search)) > 10 {
			return fmt.Errorf("search query has too many terms (max 10)")
		}
		if len(f.Search) > 200 {
			return fmt.Errorf("search query too long (max 200 chars)")
		}
	}

	return nil
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
