package constants

// Time Capsules event kinds (NIP-XX)
const (
	// KindTimeCapsule is for time-lock encrypted messages
	KindTimeCapsule = 1041
)

// Time Capsules tag names (NIP-XX)
const (
	// TagTlock contains time-lock parameters (drand_chain, drand_round)
	TagTlock = "tlock"
	// TagAlt contains human-readable description
	TagAlt = "alt"
	// TagP contains recipient public key (for private capsules)
	TagP = "p"
)

// Time Capsules payload modes (NIP-XX)
const (
	ModePublic  = 0x01 // Public time capsule (tlock only)
	ModePrivate = 0x02 // Private time capsule (ECDH + tlock)
)

// Validation limits (NIP-XX)
const (
	MaxTlockBlobSize = 256 * 1024 // 256 KiB for tlock_blob
	MaxContentSize   = 1024 * 1024 // 1 MiB for total content
	MaxPTags         = 10          // Max p tags per event
	MaxNonceSize     = 12          // ChaCha20 nonce size
	HMACSize         = 32          // HMAC-SHA256 size
)

// Default values (NIP-XX)
const (
	DefaultMaxTlockBlob = 256 * 1024 // 256 KiB
)

// Error messages (NIP-XX)
const (
	ErrInvalidMode              = "invalid payload mode"
	ErrMalformedPayload         = "malformed payload"
	ErrMissingTlockTag          = "missing tlock tag"
	ErrMissingRecipientTag      = "missing recipient tag for private mode"
	ErrTlockBlobTooLarge        = "tlock blob exceeds size limit"
	ErrContentTooLarge          = "content exceeds size limit"
	ErrHMACVerificationFailed   = "HMAC verification failed"
)
