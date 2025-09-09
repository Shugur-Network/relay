package nips

import (
	"encoding/json"
	"fmt"
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

	// Must have "d" tag for addressable events
	hasDTag := false
	var dTag string
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			dTag = tag[1]
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

	// Parse and validate JSON structure
	var stall struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Currency    string `json:"currency"`
		Shipping    []struct {
			ID      string   `json:"id"`
			Name    string   `json:"name"`
			Cost    int      `json:"cost"`
			Regions []string `json:"regions"`
		} `json:"shipping,omitempty"`
	}

	if err := json.Unmarshal([]byte(evt.Content), &stall); err != nil {
		return fmt.Errorf("invalid stall JSON format: %v", err)
	}

	// Check required fields
	if stall.ID == "" {
		return fmt.Errorf("stall must have an id")
	}
	if stall.Name == "" {
		return fmt.Errorf("stall must have a name")
	}
	if stall.Currency == "" {
		return fmt.Errorf("stall must have a currency")
	}

	// Check d tag matches stall ID
	if dTag != stall.ID {
		return fmt.Errorf("stall d tag must match stall id")
	}

	// Validate shipping zones if present
	for _, zone := range stall.Shipping {
		if zone.Cost < 0 {
			return fmt.Errorf("shipping zone must have a non-negative cost")
		}
		if len(zone.Regions) == 0 {
			return fmt.Errorf("shipping zone must have at least one region")
		}
	}

	return nil
}

// validateProductEvent validates product events (kind 30018)
func validateProductEvent(evt *nostr.Event) error {
	if evt.Kind != 30018 {
		return fmt.Errorf("invalid event kind for product: %d", evt.Kind)
	}

	// Must have "d" tag for addressable events
	hasDTag := false
	hasCategoryTag := false
	var dTag string

	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			dTag = tag[1]
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

	// Parse and validate JSON structure
	var product struct {
		ID          string     `json:"id"`
		StallID     string     `json:"stall_id"`
		Name        string     `json:"name"`
		Description string     `json:"description,omitempty"`
		Currency    string     `json:"currency"`
		Price       int        `json:"price"`
		Quantity    int        `json:"quantity,omitempty"`
		Images      []string   `json:"images,omitempty"`
		Specs       [][]string `json:"specs,omitempty"`
		Shipping    []struct {
			ID   string `json:"id"`
			Cost int    `json:"cost"`
		} `json:"shipping,omitempty"`
	}

	if err := json.Unmarshal([]byte(evt.Content), &product); err != nil {
		return fmt.Errorf("invalid product JSON format: %v", err)
	}

	// Check required fields
	if product.ID == "" {
		return fmt.Errorf("product must have an id")
	}
	if product.StallID == "" {
		return fmt.Errorf("product must have a stall_id")
	}
	if product.Name == "" {
		return fmt.Errorf("product must have a name")
	}
	if product.Currency == "" {
		return fmt.Errorf("product must have a currency")
	}
	if product.Price <= 0 {
		return fmt.Errorf("product must have a positive price")
	}

	// Check d tag matches product ID
	if dTag != product.ID {
		return fmt.Errorf("product d tag must match product id")
	}

	return nil
}

// validateMarketplaceUIEvent validates marketplace UI events (kind 30019)
func validateMarketplaceUIEvent(evt *nostr.Event) error {
	if evt.Kind != 30019 {
		return fmt.Errorf("invalid event kind for marketplace UI: %d", evt.Kind)
	}

	// Must have "d" tag for addressable events
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
		var marketplace struct {
			Name  string `json:"name"`
			About string `json:"about,omitempty"`
			UI    struct {
				Picture  string `json:"picture,omitempty"`
				Banner   string `json:"banner,omitempty"`
				Theme    string `json:"theme,omitempty"`
				DarkMode bool   `json:"darkMode,omitempty"`
			} `json:"ui,omitempty"`
		}

		if err := json.Unmarshal([]byte(evt.Content), &marketplace); err != nil {
			return fmt.Errorf("invalid marketplace JSON format: %v", err)
		}

		// Check required fields
		if marketplace.Name == "" {
			return fmt.Errorf("marketplace must have a name")
		}

		// Validate URLs if present - use simple prefix check like the original
		if marketplace.UI.Picture != "" && !isValidURL(marketplace.UI.Picture) {
			return fmt.Errorf("marketplace picture must be a valid URL")
		}
		if marketplace.UI.Banner != "" && !isValidURL(marketplace.UI.Banner) {
			return fmt.Errorf("marketplace banner must be a valid URL")
		}
	}

	return nil
}

// isValidURL checks if a string is a valid URL (simple check)
func isValidURL(str string) bool {
	return str != "" && (len(str) > 7) && (str[:7] == "http://" || str[:8] == "https://")
}

// validateAuctionEvent validates auction events (kind 30020)
func validateAuctionEvent(evt *nostr.Event) error {
	if evt.Kind != 30020 {
		return fmt.Errorf("invalid event kind for auction: %d", evt.Kind)
	}

	// Must have "d" tag for addressable events
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

	// Parse and validate JSON structure
	var auction struct {
		ID          string     `json:"id"`
		StallID     string     `json:"stall_id"`
		Name        string     `json:"name"`
		Description string     `json:"description,omitempty"`
		Images      []string   `json:"images,omitempty"`
		StartingBid int        `json:"starting_bid"`
		StartDate   int64      `json:"start_date,omitempty"`
		Duration    int64      `json:"duration"`
		Specs       [][]string `json:"specs,omitempty"`
	}

	if err := json.Unmarshal([]byte(evt.Content), &auction); err != nil {
		return fmt.Errorf("invalid auction JSON format: %v", err)
	}

	// Check required fields
	if auction.ID == "" {
		return fmt.Errorf("auction must have an id")
	}
	if auction.StallID == "" {
		return fmt.Errorf("auction must have a stall_id")
	}
	if auction.Name == "" {
		return fmt.Errorf("auction must have a name")
	}
	if auction.StartingBid <= 0 {
		return fmt.Errorf("auction must have a positive starting bid")
	}
	if auction.Duration <= 0 {
		return fmt.Errorf("auction must have a positive duration")
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
		return fmt.Errorf("bid must reference an auction with e tag")
	}

	// Content should be a positive integer (bid amount)
	if evt.Content == "" {
		return fmt.Errorf("bid event must have content")
	}

	amount, err := strconv.Atoi(evt.Content)
	if err != nil || amount <= 0 {
		return fmt.Errorf("bid amount must be a positive integer")
	}

	return nil
}

// validateBidConfirmationEvent validates bid confirmation events (kind 1022)
func validateBidConfirmationEvent(evt *nostr.Event) error {
	if evt.Kind != 1022 {
		return fmt.Errorf("invalid event kind for bid confirmation: %d", evt.Kind)
	}

	// Must have "e" tag referencing the bid
	eTags := 0
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			eTags++
		}
	}

	if eTags < 2 {
		return fmt.Errorf("bid confirmation must reference both bid and auction with e tags")
	}

	// Parse and validate JSON content if present
	if evt.Content != "" {
		var confirmation struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}

		if err := json.Unmarshal([]byte(evt.Content), &confirmation); err != nil {
			return fmt.Errorf("invalid bid confirmation JSON format: %v", err)
		}

		// Check required fields
		if confirmation.Status == "" {
			return fmt.Errorf("bid confirmation must have a status")
		}
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
