package relay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/domain"
	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/Shugur-Network/relay/internal/relay/nips"
	"github.com/Shugur-Network/relay/internal/storage"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// ValidationLimits defines your limit fields
type ValidationLimits struct {
	MaxContentLength  int
	MaxTagsLength     int
	MaxTagsPerEvent   int
	MaxTagElements    int
	MaxFutureSeconds  int
	OldestEventTime   int64
	RelayStartupTime  time.Time
	MaxMetadataLength int
	AllowedKinds      map[int]bool
	RequiredTags      map[int][]string
	MaxCreatedAt      int64
	MinCreatedAt      int64
}

// PluginValidator implements EventValidator
type PluginValidator struct {
	config    *config.Config
	blacklist map[string]bool
	limits    ValidationLimits

	verifiedPubkeys map[string]time.Time
	db              *storage.DB
}

// Ensure PluginValidator implements domain.EventValidator
var _ domain.EventValidator = (*PluginValidator)(nil)

// NewPluginValidator returns a PluginValidator with default settings
func NewPluginValidator(cfg *config.Config, database *storage.DB) *PluginValidator {
	defaultLimits := ValidationLimits{
		MaxContentLength:  64000,
		MaxTagsLength:     10000,
		MaxTagsPerEvent:   256,
		MaxTagElements:    16,
		MaxFutureSeconds:  300,
		OldestEventTime:   1609459200, // Jan 1, 2021
		RelayStartupTime:  time.Now(),
		MaxMetadataLength: 10000,
		AllowedKinds: map[int]bool{
			0: true, 1: true, 2: true, 3: true, 4: true, 5: true,
			6: true, 7: true, 40: true, 41: true, 42: true, 43: true, 44: true,
			13: true, 14: true, 15: true, 1059: true, 10050: true,
			1984: true, 9734: true, 10002: true, 30023: true, 31989: true,
			1111: true, // NIP-22: Comment
			// NIP-20 Command Results
			24133: true,
			// NIP-16 Ephemeral Events (20000-29999)
			20000: true, 20001: true, // Test ephemeral kinds
			// NIP-33 Parameterized Replaceable Events
			30000: true, 30001: true, 30002: true, 30003: true,
			// NIP-15 Marketplace
			30017: true, // Stall
			30018: true, // Product
			30019: true, // Marketplace UI/UX
			30020: true, // Auction Product
			1021:  true, // Bid
			1022:  true, // Bid Confirmation
			// Other NIPs
			1040:  true, // NIP-03 OpenTimestamps attestation
			13194: true, // NIP-59 Wallet Connect events
			30078: true, // NIP-78 Application-specific Data
		},
		RequiredTags: map[int][]string{
			5:     {"e"},      // Deletion events must have an "e" tag
			7:     {"e", "p"}, // Reaction events require "e" and "p" tags
			41:    {"e"},      // NIP-28: Channel Message requires "e" tag
			43:    {"e"},      // NIP-28: Channel Hide Message requires "e" tag
			44:    {"e"},      // NIP-28: Channel Mute User requires "e" tag
			1059:  {"p"},      // Gift wrap events must have a "p" tag
			30000: {"d"},      // NIP-33: Parameterized Replaceable Events require "d" tag
			30001: {"d"},      // NIP-33: Parameterized Replaceable Events require "d" tag
			30002: {"d"},      // NIP-33: Parameterized Replaceable Events require "d" tag
			30003: {"d"},      // NIP-33: Parameterized Replaceable Events require "d" tag
			30017: {"d"},      // Stall events require "d" tag
			30018: {"d", "t"}, // Product events require "d" and at least one "t" tag
			1021:  {"e"},      // Bid events require "e" tag
			1022:  {"e"},      // Bid confirmation events require "e" tag
			1040:  {"e"},      // OpenTimestamps attestation requires "e" tag
			30078: {"p"},      // NIP-78: Application-specific Data requires "p" tag
		},
		MaxCreatedAt: time.Now().Unix() + 300,    // 5 minutes in future
		MinCreatedAt: time.Now().Unix() - 172800, // 2 days in past
	}

	return &PluginValidator{
		config:          cfg,
		blacklist:       make(map[string]bool),
		limits:          defaultLimits,
		verifiedPubkeys: make(map[string]time.Time),
		db:              database,
	}
}

// ValidateEvent checks an event thoroughly
func (pv *PluginValidator) ValidateEvent(ctx context.Context, event nostr.Event) (bool, string) {

	// Check context cancellation at strategic points
	if ctx.Err() != nil {
		return false, "operation canceled"
	}

	// 1. Basic structure checks
	if len(event.ID) != 64 || !isHexString(event.ID) {
		return false, "invalid event ID format"
	}

	if len(event.PubKey) != 64 || !isHexString(event.PubKey) {
		return false, "invalid pubkey format"
	}

	if len(event.Sig) != 128 || !isHexString(event.Sig) {
		return false, "invalid signature format"
	}

	// 2. Check if kind is allowed
	if !pv.limits.AllowedKinds[event.Kind] {
		// Check if it's an ephemeral event (20000-29999) - these should be allowed per NIP-16
		if event.Kind >= 20000 && event.Kind < 30000 {
			// Ephemeral events are allowed but not stored
		} else {
			return false, fmt.Sprintf("unsupported event kind: %d", event.Kind)
		}
	}

	// 3. Check blacklist (case-insensitive)
	if pv.blacklist[strings.ToLower(event.PubKey)] {
		return false, "pubkey is blacklisted"
	}

	// 4. Verify event ID matches content
	computedID := event.GetID()
	if computedID != event.ID {
		return false, "event ID does not match content"
	}

	// 5. Check timestamps
	now := time.Now().Unix()
	maxFutureTime := now + int64(pv.limits.MaxFutureSeconds)

	if event.CreatedAt.Time().Unix() > maxFutureTime {
		return false, fmt.Sprintf("event timestamp is too far in the future (max %d seconds)", pv.limits.MaxFutureSeconds)
	}

	if event.CreatedAt.Time().Unix() < pv.limits.OldestEventTime {
		return false, "event timestamp is too old"
	}

	// 6. NIP-40: Check expiration timestamp
	if expTime, hasExpiration := nips.GetExpirationTime(event); hasExpiration {
		if time.Now().After(expTime) {
			return false, "event has expired"
		}
		// Validate expiration tag format
		if err := nips.ValidateExpirationTag(event); err != nil {
			return false, fmt.Sprintf("invalid expiration tag: %v", err)
		}
	}

	// 6. Content length check
	if len(event.Content) > pv.limits.MaxContentLength {
		return false, fmt.Sprintf("content exceeds maximum length of %d bytes", pv.limits.MaxContentLength)
	}

	// 7. Tags validation
	tagsSize := 0
	for _, tag := range event.Tags {
		if len(tag) > pv.limits.MaxTagElements {
			return false, "tag has too many elements"
		}
		for _, elem := range tag {
			tagsSize += len(elem)
		}
	}

	if tagsSize > pv.limits.MaxTagsLength {
		return false, "tags exceed maximum total size"
	}

	if len(event.Tags) > pv.limits.MaxTagsPerEvent {
		return false, "too many tags"
	}

	// 8. Kind-specific required tags
	if requiredTags, hasRequirements := pv.limits.RequiredTags[event.Kind]; hasRequirements {
		for _, requiredTag := range requiredTags {
			found := false
			for _, tag := range event.Tags {
				if len(tag) > 0 && tag[0] == requiredTag {
					found = true
					break
				}
			}
			if !found {
				if event.Kind == 30018 && requiredTag == "t" {
					return false, "product must have at least one category tag"
				}
				return false, fmt.Sprintf("missing required '%s' tag", requiredTag)
			}
		}
	}

	// Special handling for deletion events (kind 5)
	if event.Kind == 5 {
		// Validate deletion authorization
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "e" {
				targetEvent, err := pv.db.GetEventByID(context.Background(), tag[1])
				if err == nil && targetEvent.ID != "" && targetEvent.PubKey != event.PubKey {
					logger.Warn("Unauthorized deletion attempt",
						zap.String("deletion_pubkey", event.PubKey),
						zap.String("event_pubkey", targetEvent.PubKey),
						zap.String("event_id", tag[1]))
					return false, "unauthorized: only the event author can delete their events"
				}
			}
		}
	}

	// Special handling for NIP-03 OpenTimestamps attestation (kind 1040)
	if event.Kind == 1040 {
		// Must have at least one 'e' tag
		hasETag := false
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "e" {
				hasETag = true
				// Validate event ID format
				if len(tag[1]) != 64 || !isHexString(tag[1]) {
					return false, "invalid: invalid event ID in 'e' tag"
				}
			}
		}
		if !hasETag {
			return false, "invalid: OpenTimestamps attestation must have at least one 'e' tag"
		}

		// Optional 'alt' tag with value "opentimestamps attestation"
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "alt" && tag[1] != "opentimestamps attestation" {
				return false, "invalid: if 'alt' tag is present, it must have value 'opentimestamps attestation'"
			}
		}

		// Content must be base64 encoded OTS file data
		if event.Content == "" {
			return false, "invalid: OpenTimestamps attestation must have base64-encoded OTS file content"
		}

		// Try to decode the base64 content to verify it's valid
		_, err := base64.StdEncoding.DecodeString(event.Content)
		if err != nil {
			return false, "invalid: invalid base64 content in OpenTimestamps attestation"
		}

		// Optional: Set a size limit on the OTS file content (e.g., 2KB)
		if len(event.Content) > 2048 {
			return false, "invalid: OTS file content too large (max 2KB)"
		}
	}

	// NIP-specific validation using dedicated validators
	if err := pv.validateWithDedicatedNIPs(&event); err != nil {
		return false, fmt.Sprintf("NIP validation failed: %v", err)
	}

	return true, ""
}

// validateWithDedicatedNIPs validates events using dedicated NIP validation functions
func (pv *PluginValidator) validateWithDedicatedNIPs(event *nostr.Event) error {
	switch event.Kind {
	case 3:
		return nips.ValidateFollowList(event)
	case 4:
		return nips.ValidateEncryptedDirectMessage(event)
	case 5:
		return nips.ValidateEventDeletion(event)
	case 7:
		return nips.ValidateReaction(event)
	case 14, 15, 1059, 10050:
		return nips.ValidatePrivateDirectMessage(event)
	case 40, 41, 42, 43, 44:
		return nips.ValidatePublicChat(event)
	case 1040:
		return nips.ValidateOpenTimestampsAttestation(event)
	case 1111:
		return nips.ValidateComment(event)
	case 24133:
		return nips.ValidateCommandResult(event)
	case 30017, 30018, 30019, 30020, 1021, 1022:
		return nips.ValidateMarketplaceEvent(event)
	case 30023:
		return nips.ValidateLongFormContent(event)
	case 30078:
		return nips.ValidateApplicationSpecificData(event)
	case 13194:
		return nips.ValidateGiftWrapEvent(event)
	default:
		// Check for NIP-16 ephemeral events
		if event.Kind >= 20000 && event.Kind < 30000 {
			return nips.ValidateEventTreatment(event)
		}
		// Check if it's a parameterized replaceable event
		if nips.IsParameterizedReplaceableKind(event.Kind) {
			return nips.ValidateParameterizedReplaceableEvent(event)
		}
		// Check for NIP-24 extra metadata
		if nips.HasExtraMetadata(event) {
			return nips.ValidateExtraMetadata(event)
		}
	}

	return nil
}

// ValidateFilter ensures a filter is within safe limits
func (pv *PluginValidator) ValidateFilter(f nostr.Filter) error {
	// Apply limit cap
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 500
	}

	// Validate time range
	if f.Since != nil && f.Until != nil && f.Since.Time().Unix() > f.Until.Time().Unix() {
		return fmt.Errorf("'since' timestamp is after 'until' timestamp")
	}

	// Don't allow queries too far in the future
	now := time.Now().Unix()
	maxFutureTime := now + int64(pv.limits.MaxFutureSeconds)
	if f.Until != nil && f.Until.Time().Unix() > maxFutureTime {
		return fmt.Errorf("'until' timestamp is too far in the future")
	}

	// Check IDs format
	for _, id := range f.IDs {
		if len(id) != 64 || !isHexString(id) {
			return fmt.Errorf("invalid event ID: %s", id)
		}
	}

	// Check authors format
	for _, author := range f.Authors {
		if len(author) != 64 || !isHexString(author) {
			return fmt.Errorf("invalid pubkey in authors: %s", author)
		}
	}

	// Prevent excessive tag filters
	if len(f.Tags) > 10 {
		return fmt.Errorf("too many tag filters (max 10)")
	}

	// Check tag values
	for _, values := range f.Tags {
		if len(values) > 20 {
			return fmt.Errorf("too many values in tag filter (max 20)")
		}
	}

	return nil
}

// AddBlacklistedPubkey adds a pubkey to the blacklist
func (pv *PluginValidator) AddBlacklistedPubkey(pubkey string) {
	pv.blacklist[strings.ToLower(pubkey)] = true
}

// RemoveBlacklistedPubkey removes a pubkey from the blacklist
func (pv *PluginValidator) RemoveBlacklistedPubkey(pubkey string) {
	delete(pv.blacklist, strings.ToLower(pubkey))
}

// ValidateAndProcessEvent performs validation and processing of incoming events
func (pv *PluginValidator) ValidateAndProcessEvent(ctx context.Context, event nostr.Event) (bool, string, error) {
	// Check event size
	if len(event.Content) > 100*1024 { // 100KB max content size
		return false, "invalid: event content too large", nil
	}

	// Create a timeout context for database operations
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Direct database check for duplicates with retry
	var exists bool
	var err error
	for i := 0; i < 3; i++ {
		exists, err = pv.db.EventExists(dbCtx, event.ID)
		if err == nil {
			break
		}
		if i < 2 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return false, "error checking event existence", fmt.Errorf("database error after retries: %w", err)
	}

	if exists {
		metrics.DuplicateEvents.Inc()
		return true, "duplicate: event already exists", nil
	}

	// Verify event ID matches content (prevents ID spoofing)
	computedID := event.GetID()
	if computedID != event.ID {
		return false, "invalid: event ID does not match content", nil
	}

	// Verify signature (important for security)
	valid, err := event.CheckSignature()
	if err != nil || !valid {
		return false, "invalid: signature verification failed", nil
	}

	// Perform base validation
	valid, reason := pv.ValidateEvent(dbCtx, event)
	if !valid {
		return false, reason, nil
	}

	// Special handling for specific event kinds
	switch event.Kind {
	case 5: // deletion
		if err := nips.ValidateDeletionAuth(
			event.Tags,
			event.PubKey,
			func(id string) (nostr.Event, bool) {
				evt, err := pv.db.GetEventByID(dbCtx, id)
				if err != nil {
					logger.Error("Error fetching event for deletion validation",
						zap.String("event_id", id),
						zap.Error(err))
					return nostr.Event{}, false
				}
				return evt, true
			},
		); err != nil {
			return false, err.Error(), nil
		}
	case 0: // Metadata
		if err := pv.validateMetadataEvent(event); err != nil {
			return false, err.Error(), nil
		}
	case 14, 15: // NIP-17 Chat and File messages
		if err := pv.validateNIP17Event(); err != nil {
			return false, err.Error(), nil
		}
	case 1059: // NIP-17 Gift wrap
		if err := pv.validateGiftWrapEvent(event); err != nil {
			return false, err.Error(), nil
		}
	case 10050: // NIP-17 DM relay list
		if err := pv.validateDMRelayListEvent(event); err != nil {
			return false, err.Error(), nil
		}
	case 1040: // NIP-03 OpenTimestamps attestation
		if err := pv.validateOpenTimestampsEvent(event); err != nil {
			return false, err.Error(), nil
		}
	case 30017: // NIP-15 Stall
		if ok, msg := pv.validateStallEvent(event); !ok {
			return false, msg, nil
		}
	case 30018: // NIP-15 Product
		if ok, msg := pv.validateProductEvent(event); !ok {
			return false, msg, nil
		}
	case 30019: // NIP-15 Marketplace
		if ok, msg := pv.validateMarketplaceEvent(event); !ok {
			return false, msg, nil
		}
	case 30020: // NIP-15 Auction
		if ok, msg := pv.validateAuctionEvent(event); !ok {
			return false, msg, nil
		}
	case 1021: // NIP-15 Bid
		if ok, msg := pv.validateBidEvent(event); !ok {
			return false, msg, nil
		}
	case 1022: // NIP-15 Bid Confirmation
		if ok, msg := pv.validateBidConfirmationEvent(event); !ok {
			return false, msg, nil
		}
	case 30078: // NIP-78 Application-specific Data
		if err := pv.validateNIP78Event(event); err != nil {
			return false, err.Error(), nil
		}
	}

	// Check if delegation is being used (NIP-26)
	if delegationTag := nips.ExtractDelegationTag(event); delegationTag != nil {
		if err := nips.ValidateDelegation(&event, delegationTag); err != nil {
			return false, fmt.Sprintf("invalid delegation: %s", err.Error()), nil
		}
	}

	return true, "", nil
}

// validateNIP17Event validates NIP-17 chat and file messages
func (pv *PluginValidator) validateNIP17Event() error {
	// Direct kind 14 and 15 events must be rejected - they must be sealed and gift wrapped
	return fmt.Errorf("invalid: chat and file messages must be sealed (kind 13) and gift wrapped (kind 1059)")
}

// validateGiftWrapEvent validates a gift wrap event (kind 1059)
func (pv *PluginValidator) validateGiftWrapEvent(event nostr.Event) error {
	// Must have exactly one 'p' tag for recipient
	recipientCount := 0
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			recipientCount++
			// Validate recipient pubkey format
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid recipient pubkey format in gift wrap event")
			}
		}
	}
	if recipientCount != 1 {
		return fmt.Errorf("gift wrap events must have exactly one recipient")
	}

	// Gift wrap events must have content
	if event.Content == "" {
		return fmt.Errorf("gift wrap events must have content")
	}

	// Gift wrap events contain JSON-encoded sealed events, not NIP-44 encrypted payloads
	// Try to parse the content as JSON to ensure it's a valid sealed event
	var sealedEvent struct {
		Kind      int    `json:"kind"`
		ID        string `json:"id"`
		PubKey    string `json:"pubkey"`
		CreatedAt int64  `json:"created_at"`
		Content   string `json:"content"`
		Sig       string `json:"sig"`
	}

	if err := json.Unmarshal([]byte(event.Content), &sealedEvent); err != nil {
		return fmt.Errorf("gift wrap content must be valid JSON: %w", err)
	}

	// Validate that it looks like a sealed event (kind 13)
	if sealedEvent.Kind != 13 {
		return fmt.Errorf("gift wrap must contain a sealed event (kind 13), got kind %d", sealedEvent.Kind)
	}

	return nil
}

// validateDMRelayListEvent validates a DM relay list event (kind 10050)
func (pv *PluginValidator) validateDMRelayListEvent(event nostr.Event) error {
	// Must have at least one relay tag
	hasRelay := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			hasRelay = true
			break
		}
	}
	if !hasRelay {
		return fmt.Errorf("DM relay list must have at least one relay tag")
	}

	return nil
}

// validateOpenTimestampsEvent validates a NIP-03 OpenTimestamps attestation event
func (pv *PluginValidator) validateOpenTimestampsEvent(event nostr.Event) error {
	// Must have at least one 'e' tag
	hasETag := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasETag = true
			// Validate event ID format
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid: invalid event ID in 'e' tag")
			}
		}
	}
	if !hasETag {
		return fmt.Errorf("invalid: OpenTimestamps attestation must have at least one 'e' tag")
	}

	// Optional 'alt' tag with value "opentimestamps attestation"
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "alt" && tag[1] != "opentimestamps attestation" {
			return fmt.Errorf("invalid: if 'alt' tag is present, it must have value 'opentimestamps attestation'")
		}
	}

	// Content must be base64 encoded OTS file data
	if event.Content == "" {
		return fmt.Errorf("invalid: OpenTimestamps attestation must have base64-encoded OTS file content")
	}

	// Try to decode the base64 content to verify it's valid
	_, err := base64.StdEncoding.DecodeString(event.Content)
	if err != nil {
		return fmt.Errorf("invalid: invalid base64 content in OpenTimestamps attestation")
	}

	// Optional: Set a size limit on the OTS file content (e.g., 2KB)
	if len(event.Content) > 2048 {
		return fmt.Errorf("invalid: OTS file content too large (max 2KB)")
	}

	return nil
}

// validateMetadataEvent validates a metadata event (kind 0)
func (pv *PluginValidator) validateMetadataEvent(event nostr.Event) error {
	// Ensure content is valid JSON
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(event.Content), &metadata); err != nil {
		return fmt.Errorf("metadata must be valid JSON: %w", err)
	}

	// Validate common metadata fields
	if name, ok := metadata["name"].(string); ok && len(name) > 100 {
		return fmt.Errorf("name field too long (max 100 characters)")
	}

	if about, ok := metadata["about"].(string); ok && len(about) > 500 {
		return fmt.Errorf("about field too long (max 500 characters)")
	}

	return nil
}

// validateStallEvent validates a NIP-15 stall event
func (pv *PluginValidator) validateStallEvent(event nostr.Event) (bool, string) {
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

	if err := json.Unmarshal([]byte(event.Content), &stall); err != nil {
		return false, "invalid stall JSON format"
	}

	// Check required fields
	if stall.ID == "" {
		return false, "stall must have an id"
	}
	if stall.Name == "" {
		return false, "stall must have a name"
	}
	if stall.Currency == "" {
		return false, "stall must have a currency"
	}

	// Check d tag matches stall ID
	var dTag string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			dTag = tag[1]
			break
		}
	}
	if dTag != stall.ID {
		return false, "stall d tag must match stall id"
	}

	// Validate shipping zones if present
	for _, zone := range stall.Shipping {
		if zone.Cost < 0 {
			return false, "shipping zone must have a non-negative cost"
		}
		if len(zone.Regions) == 0 {
			return false, "shipping zone must have at least one region"
		}
	}

	return true, ""
}

// validateProductEvent validates a NIP-15 product event
func (pv *PluginValidator) validateProductEvent(event nostr.Event) (bool, string) {
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

	if err := json.Unmarshal([]byte(event.Content), &product); err != nil {
		return false, "invalid product JSON format"
	}

	// Check required fields
	if product.ID == "" {
		return false, "product must have an id"
	}
	if product.StallID == "" {
		return false, "product must have a stall_id"
	}
	if product.Name == "" {
		return false, "product must have a name"
	}
	if product.Currency == "" {
		return false, "product must have a currency"
	}
	if product.Price <= 0 {
		return false, "product must have a positive price"
	}

	// Check d tag matches product ID
	var dTag string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			dTag = tag[1]
			break
		}
	}
	if dTag != product.ID {
		return false, "product d tag must match product id"
	}

	// Check for at least one category tag
	hasCategory := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "t" {
			hasCategory = true
			break
		}
	}
	if !hasCategory {
		return false, "product must have at least one category tag"
	}

	return true, ""
}

// validateMarketplaceEvent validates a NIP-15 marketplace event
func (pv *PluginValidator) validateMarketplaceEvent(event nostr.Event) (bool, string) {
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

	if err := json.Unmarshal([]byte(event.Content), &marketplace); err != nil {
		return false, "invalid marketplace JSON format"
	}

	// Check required fields
	if marketplace.Name == "" {
		return false, "marketplace must have a name"
	}

	// Validate URLs if present
	if marketplace.UI.Picture != "" && !strings.HasPrefix(marketplace.UI.Picture, "http") {
		return false, "marketplace picture must be a valid URL"
	}
	if marketplace.UI.Banner != "" && !strings.HasPrefix(marketplace.UI.Banner, "http") {
		return false, "marketplace banner must be a valid URL"
	}

	return true, ""
}

// validateAuctionEvent validates a NIP-15 auction event
func (pv *PluginValidator) validateAuctionEvent(event nostr.Event) (bool, string) {
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

	if err := json.Unmarshal([]byte(event.Content), &auction); err != nil {
		return false, "invalid auction JSON format"
	}

	// Check required fields
	if auction.ID == "" {
		return false, "auction must have an id"
	}
	if auction.StallID == "" {
		return false, "auction must have a stall_id"
	}
	if auction.Name == "" {
		return false, "auction must have a name"
	}
	if auction.StartingBid <= 0 {
		return false, "auction must have a positive starting bid"
	}
	if auction.Duration <= 0 {
		return false, "auction must have a positive duration"
	}

	return true, ""
}

// validateBidEvent validates a NIP-15 bid event
func (pv *PluginValidator) validateBidEvent(event nostr.Event) (bool, string) {
	// Check for e tag referencing auction
	hasETag := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasETag = true
			break
		}
	}
	if !hasETag {
		return false, "bid must reference an auction with e tag"
	}

	// Content should be a positive integer (bid amount)
	amount, err := strconv.Atoi(event.Content)
	if err != nil || amount <= 0 {
		return false, "bid amount must be a positive integer"
	}

	return true, ""
}

// validateBidConfirmationEvent validates a NIP-15 bid confirmation event
func (pv *PluginValidator) validateBidConfirmationEvent(event nostr.Event) (bool, string) {
	var confirmation struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal([]byte(event.Content), &confirmation); err != nil {
		return false, "invalid bid confirmation JSON format"
	}

	// Check required fields
	if confirmation.Status == "" {
		return false, "bid confirmation must have a status"
	}

	// Check for e tags referencing both bid and auction
	hasBidETag := false
	hasAuctionETag := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			if !hasBidETag {
				hasBidETag = true
			} else {
				hasAuctionETag = true
			}
		}
	}
	if !hasBidETag || !hasAuctionETag {
		return false, "bid confirmation must reference both bid and auction with e tags"
	}

	return true, ""
}

// validateNIP78Event validates a NIP-78 application-specific data event
func (pv *PluginValidator) validateNIP78Event(event nostr.Event) error {
	// Content must be valid JSON
	var appData struct {
		Name string      `json:"name"`
		Data interface{} `json:"data"`
	}

	if err := json.Unmarshal([]byte(event.Content), &appData); err != nil {
		return fmt.Errorf("invalid: application-specific data must be valid JSON: %w", err)
	}

	// Check required fields
	if appData.Name == "" {
		return fmt.Errorf("invalid: application-specific data must have a 'name' field")
	}

	// Check that data field exists (can be empty object but must be present)
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(event.Content), &rawData); err != nil {
		return fmt.Errorf("invalid: failed to parse application-specific data: %w", err)
	}

	if _, hasData := rawData["data"]; !hasData {
		return fmt.Errorf("invalid: application-specific data must have a 'data' field")
	}

	// Validate that data field is an object, not a string or other primitive
	if dataValue, ok := rawData["data"]; ok {
		if _, isObject := dataValue.(map[string]interface{}); !isObject {
			return fmt.Errorf("invalid: application-specific data 'data' field must be an object")
		}
	}

	// Validate name field constraints
	if len(appData.Name) > 100 {
		return fmt.Errorf("invalid: application-specific data 'name' field too long (max 100 characters)")
	}

	return nil
}
