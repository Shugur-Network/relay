package constants

import (
	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/identity"
	nip11 "github.com/nbd-wtf/go-nostr/nip11"
)

// Database constants
const (
	DatabaseName = "shugur"
)

// Default relay metadata constants
const (
	DefaultRelayDescription = "High-performance, reliable, scalable Nostr relay for decentralized communication."
	DefaultRelayContact     = "support@shugur.com"
	DefaultRelaySoftware    = "shugur"
	DefaultRelayVersion     = "2.0.0"
	DefaultRelayIcon        = "https://avatars.githubusercontent.com/u/198367099?s=400&u=2bc76d4fe6f57a1c39ef00fd784dd0bf85d79bda&v=4"
)

// DefaultSupportedNIPs lists the NIPs supported by the relay
var DefaultSupportedNIPs = []interface{}{
	1,  // NIP-01: Basic protocol flow description
	2,  // NIP-02: Follow List
	3,  // NIP-03: OpenTimestamps Attestations for Events
	4,  // NIP-04: Encrypted Direct Message (deprecated, use NIP-17)
	9,  // NIP-09: Event Deletion Request
	11, // NIP-11: Relay Information Document
	15, // NIP-15: Nostr Marketplace (for resilient marketplaces)
	16, // NIP-16: Event Treatment
	17, // NIP-17: Private Direct Messages
	20, // NIP-20: Command Results
	22, // NIP-22: Comment
	23, // NIP-23: Long-form Content
	24, // NIP-24: Extra metadata fields and tags
	25, // NIP-25: Reactions
	28, // NIP-28: Public Chat
	33, // NIP-33: Parameterized Replaceable Events
	40, // NIP-40: Expiration Timestamp
	44, // NIP-44: Encrypted Payloads (Versioned)
	45, // NIP-45: Counting Events
	50, // NIP-50: Search Capability
	59, // NIP-59: Gift Wrap
	65, // NIP-65: Relay List Metadata
	78, // NIP-78: Application-specific data
}

// Relay limitations and settings
const (
	MaxMessageLength = 2048
	MaxSubscriptions = 100
	MaxFilters       = 100
	MaxLimit         = 100
	MaxSubIDLength   = 100
	MaxEventTags     = 100
	MaxContentLength = 2048
	MinPowDifficulty = 0
	AuthRequired     = false
	PaymentRequired  = false
	RestrictedWrites = false
)

// DefaultRelayMetadata returns the default relay metadata document
func DefaultRelayMetadata(cfg *config.Config) nip11.RelayInformationDocument {
	// Get or create relay identity
	relayIdentity, err := identity.GetOrCreateRelayIdentity()
	if err != nil {
		// Fallback to default if identity system fails
		relayIdentity = &identity.RelayIdentity{
			RelayID:   "relay-unknown",
			PublicKey: "unknown",
		}
	}

	// Use relay name from config, fallback to "shugur-relay" if empty
	relayName := cfg.Relay.Name
	if relayName == "" {
		relayName = "shugur-relay"
	}

	// Use relay description from config, fallback to default if empty
	relayDescription := cfg.Relay.Description
	if relayDescription == "" {
		relayDescription = DefaultRelayDescription
	}

	// Use relay contact from config, fallback to default if empty
	relayContact := cfg.Relay.Contact
	if relayContact == "" {
		relayContact = DefaultRelayContact
	}

	// Use relay icon from config, fallback to default if empty
	relayIcon := cfg.Relay.Icon
	if relayIcon == "" {
		relayIcon = DefaultRelayIcon
	}

	// Use relay banner from config if provided
	relayBanner := cfg.Relay.Banner

	// Use actual configuration values for limitations instead of hardcoded constants
	maxContentLength := cfg.Relay.ThrottlingConfig.MaxContentLen
	if maxContentLength == 0 {
		maxContentLength = MaxContentLength // fallback to default constant
	}

	return nip11.RelayInformationDocument{
		Name:          relayName,
		Description:   relayDescription,
		Contact:       relayContact,
		PubKey:        relayIdentity.PublicKey,
		SupportedNIPs: DefaultSupportedNIPs,
		Software:      DefaultRelaySoftware,
		Version:       config.Version,
		Icon:          relayIcon,
		Banner:        relayBanner,
		Limitation: &nip11.RelayLimitationDocument{
			MaxMessageLength: maxContentLength,      // Use actual configured content length
			MaxSubscriptions: MaxSubscriptions,      // Keep default for now (could be made configurable)
			MaxFilters:       MaxFilters,            // Keep default for now (could be made configurable)
			MaxLimit:         MaxLimit,              // Keep default for now (could be made configurable)
			MaxSubidLength:   MaxSubIDLength,        // Keep default for now (could be made configurable)
			MaxEventTags:     MaxEventTags,          // Keep default for now (could be made configurable)
			MaxContentLength: maxContentLength,      // Use actual configured content length
			MinPowDifficulty: MinPowDifficulty,      // Keep default for now (could be made configurable)
			AuthRequired:     AuthRequired,          // Keep default for now (could be made configurable)
			PaymentRequired:  PaymentRequired,       // Keep default for now (could be made configurable)
			RestrictedWrites: RestrictedWrites,      // Keep default for now (could be made configurable)
		},
	}
}
