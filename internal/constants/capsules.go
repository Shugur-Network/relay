package constants

// Time Capsules event kinds
const (
	// KindTimeCapsule is for immutable time capsules (NIP-01 compatible)
	KindTimeCapsule = 11990
	// KindTimeCapsuleReplaceable is for parameterized replaceable time capsules
	KindTimeCapsuleReplaceable = 30095
	// KindTimeCapsuleUnlockShare is for witness unlock shares
	KindTimeCapsuleUnlockShare = 11991
	// KindTimeCapsuleShareDistribution is for share distribution events
	KindTimeCapsuleShareDistribution = 11992
)

// Time Capsules tag names
const (
	// TagUnlock contains unlock configuration (mode, t, n, T)
	TagUnlock = "u"
	// TagWitnessCommit contains merkle root of witnesses
	TagWitnessCommit = "w-commit"
	// TagEncryption contains encryption metadata
	TagEncryption = "enc"
	// TagLocation indicates where content is stored
	TagLocation = "loc"
	// TagURI contains external content URI
	TagURI = "uri"
	// TagProof contains auxiliary proof data
	TagProof = "proof"
	// TagShareIndex contains the share index for distribution
	TagShareIndex = "share-idx"
)

// Time Capsules modes
const (
	ModeThreshold = "threshold"
	ModeScheduled = "scheduled"
)

// Storage locations
const (
	LocationInline  = "inline"
	LocationHTTPS   = "https"
	LocationBlossom = "blossom"
	LocationIPFS    = "ipfs"
)

// Encryption algorithms
const (
	EncryptionNIP44v2 = "nip44:v2"
)

// Time Capsules status values
const (
	StatusLocked   = "locked"
	StatusUnlocked = "unlocked"
	StatusExpired  = "expired"
	StatusInvalid  = "invalid"
)

// Validation limits
const (
	MaxCapsuleContentLength = 2 * 1024 * 1024 // 2MB max content
	MaxWitnessCount         = 10              // Max witnesses per capsule (aligned with config.yaml)
	MaxThresholdValue       = 10              // Max threshold value (aligned with max witnesses)
	MinThresholdValue       = 1               // Min threshold value
	MaxUnlockTimeYears      = 10              // Max years in future for unlock time
	MinClockSkewSeconds     = 60              // Min clock skew (1 minute)
	MaxClockSkewSeconds     = 3600            // Max clock skew (1 hour)
)

// Default values
const (
	DefaultThreshold     = 3
	DefaultWitnessCount  = 5
	DefaultClockSkewSec  = 300        // 5 minutes
	DefaultMaxInlineSize = 128 * 1024 // 128 KiB
)

// Error messages
const (
	ErrCapsuleNotFound      = "capsule not found"
	ErrCapsuleAlreadyExists = "capsule already exists"
	ErrInvalidWitnessCommit = "invalid witness commit"
	ErrTooEarly             = "shares cannot be posted before unlock time"
	ErrInvalidThreshold     = "invalid threshold configuration"
	ErrWitnessNotInCommit   = "witness not in commit"
	ErrInsufficientShares   = "insufficient shares to unlock"
	ErrInvalidEncryption    = "invalid encryption configuration"
	ErrContentTooLarge      = "content exceeds maximum size"
	ErrUnsupportedLocation  = "unsupported storage location"
)
