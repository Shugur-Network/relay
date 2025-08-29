package relay

import (
	"context"
	"encoding/json"
	"fmt"
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
	// Use configuration values for content length limits
	maxContentLength := cfg.Relay.ThrottlingConfig.MaxContentLen
	if maxContentLength == 0 {
		maxContentLength = 64000 // fallback default
	}

	defaultLimits := ValidationLimits{
		MaxContentLength:  maxContentLength,  // Use configured value
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
			// Time Capsules
			11990: true, // Time capsule (immutable)
			30095: true, // Time capsule (parameterized replaceable)
			11991: true, // Time capsule unlock share
			11992: true, // Time capsule share distribution
		},
		RequiredTags: map[int][]string{
			5:     {"e"},      // Deletion events must have an "e" tag
			7:     {"e", "p"}, // Reaction events require "e" and "p" tags
			41:    {"e"},      // NIP-28: Channel Metadata requires "e" tag
			42:    {"e"},      // NIP-28: Channel Message requires "e" tag
			43:    {"e"},      // NIP-28: Hide Message requires "e" tag
			44:    {"p"},      // NIP-28: Mute User requires "p" tag
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
			// Time Capsules
			11990: {"u", "p", "w-commit", "enc", "loc"}, // Time capsule: unlock config, witnesses, commitment, encryption, location
			30095: {"u", "p", "w-commit", "enc", "loc", "d"}, // Replaceable time capsule: + d tag
			11991: {"e", "p", "T"}, // Unlock share: capsule ref, witness, unlock time
			11992: {"e", "p", "share-idx", "enc"}, // Share distribution: capsule ref, witness, share index, encryption
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
		// Skip generic tag validation for Time Capsules kinds - they have specialized validation
		if event.Kind == 11990 || event.Kind == 30095 {
			// Time Capsules have complex validation logic that varies by mode
			// This is handled in the specialized NIP validation below
		} else {
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
	}

	// Special handling for deletion events (kind 5)
	if event.Kind == 5 {
		// Validate deletion authorization
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "e" {
				targetEvent, err := pv.db.GetEventByID(context.Background(), tag[1])
				if err == nil && targetEvent.ID != "" && targetEvent.PubKey != event.PubKey {
					logger.Warn("Unauthorized deletion attempt blocked",
						zap.String("deletion_event_id", event.ID),
						zap.String("deleter_pubkey", event.PubKey),
						zap.String("target_event_id", tag[1]),
						zap.String("target_event_pubkey", targetEvent.PubKey))
					return false, "unauthorized: only the event author can delete their events"
				}
			}
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
	case 14, 15, 10050:
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
	case 10002:
		return nips.ValidateKind10002(*event)
	case 11990, 30095:
		return nips.ValidateTimeCapsuleEvent(event)
	case 11991:
		return nips.ValidateTimeCapsuleUnlockShare(event)
	case 11992:
		return nips.ValidateTimeCapsuleShareDistribution(event)
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
	// Check event size using configured limit
	if len(event.Content) > pv.limits.MaxContentLength {
		return false, fmt.Sprintf("invalid: event content too large (max %d bytes)", pv.limits.MaxContentLength), nil
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
	case 1059: // NIP-17 Gift wrap - use NIP-17 validation (more lenient than NIP-59)
		if err := nips.ValidatePrivateDirectMessage(&event); err != nil {
			return false, err.Error(), nil
		}
	}

	// Check if delegation is being used (NIP-26)
	if delegationTag := nips.ExtractDelegationTag(event); delegationTag != nil {
		if err := nips.ValidateDelegation(&event, delegationTag); err != nil {
			return false, fmt.Sprintf("invalid delegation: %s", err.Error()), nil
		}
		logger.Debug("Event with valid delegation accepted",
			zap.String("event_id", event.ID),
			zap.String("delegator", delegationTag.MasterPubkey))
	}

	logger.Debug("Event validation successful",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey),
		zap.Int("kind", event.Kind))
	return true, "", nil
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
