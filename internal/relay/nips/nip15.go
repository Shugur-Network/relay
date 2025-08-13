package nips

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-15: Nostr Marketplace (for resilient marketplaces)
// https://github.com/nostr-protocol/nips/blob/master/15.md

// ValidateMarketplaceEvent validates NIP-15 marketplace events
func ValidateMarketplaceEvent(evt *nostr.Event) error {
	switch evt.Kind {
	case 30017:
		return validateStallEvent(evt)
	case 30018:
		return validateProductEvent(evt)
	case 30019:
		return validateMarketplaceUIEvent(evt)
	case 30020:
		return validateAuctionEvent(evt)
	case 1021:
		return validateBidEvent(evt)
	case 1022:
		return validateBidConfirmationEvent(evt)
	default:
		return fmt.Errorf("invalid event kind for marketplace event: %d", evt.Kind)
	}
}

// validateStallEvent validates stall events (kind 30017)
func validateStallEvent(evt *nostr.Event) error {
	if evt.Kind != 30017 {
		return fmt.Errorf("invalid event kind for stall: %d", evt.Kind)
	}

	// Must have "d" tag for parameterized replaceable events
	hasDTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			break
		}
	}

	if !hasDTag {
		return fmt.Errorf("stall event must have 'd' tag")
	}

	// Content should contain stall information (JSON)
	if evt.Content == "" {
		return fmt.Errorf("stall event must have content")
	}

	return nil
}

// validateProductEvent validates product events (kind 30018)
func validateProductEvent(evt *nostr.Event) error {
	if evt.Kind != 30018 {
		return fmt.Errorf("invalid event kind for product: %d", evt.Kind)
	}

	// Must have "d" tag for parameterized replaceable events
	hasDTag := false
	hasCategoryTag := false

	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
		}
		if len(tag) >= 2 && tag[0] == "t" {
			hasCategoryTag = true
		}
	}

	if !hasDTag {
		return fmt.Errorf("product event must have 'd' tag")
	}

	if !hasCategoryTag {
		return fmt.Errorf("product must have at least one category tag")
	}

	// Content should contain product information (JSON)
	if evt.Content == "" {
		return fmt.Errorf("product event must have content")
	}

	// Parse and validate JSON content
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(evt.Content), &product); err != nil {
		return fmt.Errorf("product content must be valid JSON: %v", err)
	}

	// Validate price if present
	if price, exists := product["price"]; exists {
		var priceValue float64
		switch v := price.(type) {
		case float64:
			priceValue = v
		case int:
			priceValue = float64(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				priceValue = parsed
			} else {
				return fmt.Errorf("product price must be a valid number")
			}
		default:
			return fmt.Errorf("product price must be a number")
		}

		if priceValue <= 0 {
			return fmt.Errorf("product must have a positive price")
		}
	}

	return nil
}

// validateMarketplaceUIEvent validates marketplace UI events (kind 30019)
func validateMarketplaceUIEvent(evt *nostr.Event) error {
	if evt.Kind != 30019 {
		return fmt.Errorf("invalid event kind for marketplace UI: %d", evt.Kind)
	}

	// Must have "d" tag for parameterized replaceable events
	hasDTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			break
		}
	}

	if !hasDTag {
		return fmt.Errorf("marketplace UI event must have 'd' tag")
	}

	// Parse and validate JSON content
	if evt.Content != "" {
		var marketplace map[string]interface{}
		if err := json.Unmarshal([]byte(evt.Content), &marketplace); err != nil {
			return fmt.Errorf("marketplace content must be valid JSON: %v", err)
		}

		// Validate UI field if present
		if ui, exists := marketplace["ui"]; exists {
			if uiMap, ok := ui.(map[string]interface{}); ok {
				// Validate picture URL if present
				if picture, exists := uiMap["picture"]; exists {
					if pictureStr, ok := picture.(string); ok && pictureStr != "" {
						if _, err := url.ParseRequestURI(pictureStr); err != nil {
							return fmt.Errorf("marketplace picture must be a valid URL")
						}
					}
				}
				// Validate banner URL if present
				if banner, exists := uiMap["banner"]; exists {
					if bannerStr, ok := banner.(string); ok && bannerStr != "" {
						if _, err := url.ParseRequestURI(bannerStr); err != nil {
							return fmt.Errorf("marketplace banner must be a valid URL")
						}
					}
				}
			}
		}
	}

	return nil
}

// validateAuctionEvent validates auction events (kind 30020)
func validateAuctionEvent(evt *nostr.Event) error {
	if evt.Kind != 30020 {
		return fmt.Errorf("invalid event kind for auction: %d", evt.Kind)
	}

	// Must have "d" tag for parameterized replaceable events
	hasDTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			break
		}
	}

	if !hasDTag {
		return fmt.Errorf("auction event must have 'd' tag")
	}

	// Content should contain auction information (JSON)
	if evt.Content == "" {
		return fmt.Errorf("auction event must have content")
	}

	// Parse and validate JSON content
	var auction map[string]interface{}
	if err := json.Unmarshal([]byte(evt.Content), &auction); err != nil {
		return fmt.Errorf("auction content must be valid JSON: %v", err)
	}

	// Validate duration if present
	if duration, exists := auction["duration"]; exists {
		var durationValue float64
		switch v := duration.(type) {
		case float64:
			durationValue = v
		case int:
			durationValue = float64(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				durationValue = parsed
			} else {
				return fmt.Errorf("auction duration must be a valid number")
			}
		default:
			return fmt.Errorf("auction duration must be a number")
		}

		if durationValue <= 0 {
			return fmt.Errorf("auction must have a positive duration")
		}
	}

	return nil
}

// validateBidEvent validates bid events (kind 1021)
func validateBidEvent(evt *nostr.Event) error {
	if evt.Kind != 1021 {
		return fmt.Errorf("invalid event kind for bid: %d", evt.Kind)
	}

	// Must have "e" tag referencing the auction
	hasAuctionTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasAuctionTag = true
			break
		}
	}

	if !hasAuctionTag {
		return fmt.Errorf("bid event must reference auction with 'e' tag")
	}

	// Content should contain bid information
	if evt.Content == "" {
		return fmt.Errorf("bid event must have content")
	}

	return nil
}

// validateBidConfirmationEvent validates bid confirmation events (kind 1022)
func validateBidConfirmationEvent(evt *nostr.Event) error {
	if evt.Kind != 1022 {
		return fmt.Errorf("invalid event kind for bid confirmation: %d", evt.Kind)
	}

	// Must have "e" tag referencing the bid
	hasBidTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasBidTag = true
			break
		}
	}

	if !hasBidTag {
		return fmt.Errorf("bid confirmation event must reference bid with 'e' tag")
	}

	return nil
}

// IsMarketplaceEvent checks if an event is a marketplace event
func IsMarketplaceEvent(evt *nostr.Event) bool {
	return evt.Kind == 30017 || evt.Kind == 30018 || evt.Kind == 30019 ||
		evt.Kind == 30020 || evt.Kind == 1021 || evt.Kind == 1022
}

// GetMarketplaceEventType returns a human-readable type for marketplace events
func GetMarketplaceEventType(kind int) string {
	switch kind {
	case 30017:
		return "stall"
	case 30018:
		return "product"
	case 30019:
		return "marketplace-ui"
	case 30020:
		return "auction"
	case 1021:
		return "bid"
	case 1022:
		return "bid-confirmation"
	default:
		return "unknown"
	}
}
