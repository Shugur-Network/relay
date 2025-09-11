package constants

// Time Capsules event kinds (NIP-XX)
const (
	// KindTimeCapsule is for time-lock encrypted messages
	KindTimeCapsule = 1041
	// KindSeal is for NIP-59 sealed events (rumor wrapped in NIP-44 encryption)
	KindSeal = 13
	// KindGiftWrap is for NIP-59 gift wrapped events (seal wrapped in ephemeral encryption)
	KindGiftWrap = 1059
)

// Time Capsules tag names (NIP-XX)
const (
	// TagTlock contains time-lock parameters in format: ["tlock", "<drand_chain_hex64>", "<drand_round_uint>"]
	TagTlock = "tlock"
	// TagAlt contains human-readable description
	TagAlt = "alt"
	// TagP contains recipient public key (for routing gift wraps)
	TagP = "p"
)

// NIP-44 constants
const (
	// NIP44Version is the current NIP-44 version
	NIP44Version = 0x02
	// NIP44NonceSize is the nonce size for NIP-44 v2 (32 bytes)
	NIP44NonceSize = 32
	// NIP44MacSize is the MAC size for NIP-44 v2 (32 bytes HMAC-SHA256)
	NIP44MacSize = 32
	// NIP44ChaChaKeySize is the ChaCha20 key size (32 bytes)
	NIP44ChaChaKeySize = 32
	// NIP44ChachaNonceSize is the ChaCha20 nonce size (12 bytes)
	NIP44ChachaNonceSize = 12
)

// Validation limits (NIP-XX)
const (
	// MaxContentSize is the maximum size for decoded time capsule content (64 KiB per spec)
	MaxContentSize = 64 * 1024
	// MaxRelayContentSize is the relay limit for encoded content (256 KiB)
	MaxRelayContentSize = 256 * 1024
	// DrandChainHashLength is the expected length of drand chain hash (64 hex chars = 32 bytes)
	DrandChainHashLength = 64
	// MaxDrandRound is the maximum drand round number (2^63-1 for safety)
	MaxDrandRound = 9223372036854775807
)

// Error messages (NIP-XX)
const (
	ErrMalformedPayload         = "malformed payload"
	ErrMissingTlockTag          = "missing tlock tag"
	ErrInvalidTlockFormat       = "invalid tlock tag format"
	ErrInvalidDrandChain        = "invalid drand chain hash format"
	ErrInvalidDrandRound        = "invalid drand round format"
	ErrContentTooLarge          = "content exceeds size limit"
	ErrInvalidBase64            = "invalid base64 content"
	ErrInvalidNIP44Version      = "invalid NIP-44 version"
	ErrNIP44PayloadTooSmall     = "NIP-44 payload too small"
	ErrEmptyTags                = "seal event must have empty tags"
	ErrMissingSealContent       = "seal event missing content"
	ErrMissingGiftWrapRecipient = "gift wrap missing recipient tag"
)
