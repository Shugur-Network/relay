package constants

// Time Capsules event kinds (NIP Time Capsules v1)
const (
	// KindTimeCapsule is for immutable time capsules
	KindTimeCapsule = 1995
	// KindTimeCapsuleReplaceable is for parameterized replaceable time capsules
	KindTimeCapsuleReplaceable = 31995
	// KindTimeCapsuleShareDistribution is for share distribution events (optional helper)
	KindTimeCapsuleShareDistribution = 1996
	// KindTimeCapsuleUnlockShare is for witness unlock shares
	KindTimeCapsuleUnlockShare = 1997
	
	// NIP-59 Gift Wrap (required for private share delivery)
	KindGiftWrap = 1059
)

// Time Capsules tag names (NIP Time Capsules v1)
const (
	// TagUnlock contains unlock configuration with space-delimited key/value pairs
	TagUnlock = "unlock"
	// TagWitnessCommit contains merkle root of witnesses (sha256:<hex>)
	TagWitnessCommit = "w-commit"
	// TagEncryption contains encryption metadata (must be "nip44:v2")
	TagEncryption = "enc"
	// TagLocation indicates where content is stored
	TagLocation = "loc"
	// TagURI contains external content URI
	TagURI = "uri"
	// TagSHA256 contains integrity hash for external content
	TagSHA256 = "sha256"
	// TagShareIndex contains the share index for distribution/unlock (1..n)
	TagShareIndex = "share-idx"
	// TagAAD contains auxiliary authenticated data hex
	TagAAD = "aad"
	// TagBeacon contains drand chain hash (32-byte hex)
	TagBeacon = "beacon"
	// TagRound contains drand round number
	TagRound = "round"
	// TagChainPK contains optional drand chain BLS public key (Base64)
	TagChainPK = "chain-pk"
	// TagExpiration contains expiration timestamp per NIP-40
	TagExpiration = "expiration"
	// TagAlt contains human-readable description
	TagAlt = "alt"
	// TagSkew contains UI hint for not-before tolerance
	TagSkew = "skew"
)

// Time Capsules modes (NIP Time Capsules v1)
const (
	// ModeThreshold requires t-of-n witness shares (no time constraint)
	ModeThreshold = "threshold"
	// ModeThresholdTime requires t-of-n witness shares AND drand timelock
	ModeThresholdTime = "threshold-time"
	// ModeTimelock requires only drand timelock (public after T, no witnesses)
	ModeTimelock = "timelock"
)

// Storage locations
const (
	LocationInline  = "inline"
	LocationHTTPS   = "https"
	LocationBlossom = "blossom"
	LocationIPFS    = "ipfs"
)

// Encryption algorithms (NIP Time Capsules v1)
const (
	// EncryptionNIP44v2 is the required encryption format
	EncryptionNIP44v2 = "nip44:v2"
)

// Content envelope version (NIP Time Capsules v1)
const (
	// EnvelopeVersion is the current content envelope version
	EnvelopeVersion = "1"
)

// Drand constants (NIP Time Capsules v1)
const (
	// DrandMainnetChainHash is the mainnet drand chain hash
	DrandMainnetChainHash = "8990e7a9fd29b9427d2bc63c37e2cb28e0b2d73dbb5c18f9b92b057c4c5a8580"
	// DrandTestnetChainHash is the testnet drand chain hash
	DrandTestnetChainHash = "7672797f548f3f4748ac4bf3352fc6c6b6468c9ad40ad456a397545c6e2df5bf"
	// DrandDefaultPeriod is the default period in seconds
	DrandDefaultPeriod = 30
	// DrandGenesisMainnet is the mainnet genesis time
	DrandGenesisMainnet = 1672531200 // 2023-01-01 00:00:00 UTC
	// DrandGenesisTestnet is the testnet genesis time  
	DrandGenesisTestnet = 1651677600 // 2022-05-05 00:00:00 UTC
)

// HKDF constants (NIP Time Capsules v1)
const (
	// HKDFInfo is the info parameter for HKDF key derivation
	HKDFInfo = "capsule/v1"
	// CombinerPrefix is the prefix byte for R_combined derivation
	CombinerPrefix = 0x01
)

// Shamir constants (NIP Time Capsules v1)
const (
	// ShamirFieldPrime is the field prime for Shamir secret sharing (secp256k1 prime)
	ShamirFieldPrime = "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F" // 2^256 - 2^32 - 977
	// ShamirShareSize is the size of each share in bytes
	ShamirShareSize = 32
	// ShamirXCoordMin is the minimum x-coordinate for shares
	ShamirXCoordMin = 1
)

// Merkle tree constants (NIP Time Capsules v1) 
const (
	// WitnessMerkleLeafDomain is the domain string for witness leaf computation
	WitnessMerkleLeafDomain = "nostr:capsule:witness/v1"
	// WitnessMerkleAlgorithm is the algorithm prefix for w-commit
	WitnessMerkleAlgorithm = "sha256"
)

// Time Capsules status values
const (
	StatusLocked   = "locked"
	StatusUnlocked = "unlocked"
	StatusExpired  = "expired"
	StatusInvalid  = "invalid"
)

// Validation limits (NIP Time Capsules v1)
const (
	MaxCapsuleContentLength = 2 * 1024 * 1024 // 2MB max content
	MaxWitnessCount         = 256             // Max witnesses per capsule (per NIP DoS limits)
	MaxThresholdValue       = 256             // Max threshold value (aligned with max witnesses)
	MinThresholdValue       = 1               // Min threshold value
	MaxUnlockTimeYears      = 10              // Max years in future for unlock time
	MinClockSkewSeconds     = 60              // Min clock skew (1 minute)
	MaxClockSkewSeconds     = 3600            // Max clock skew (1 hour)
	MaxInlineBytes          = 131072          // Max inline content (128 KiB per NIP)
	MaxUnlockTagLength      = 4096            // Max unlock tag length (4 KiB per NIP)
	MaxKTlockSize           = 1024            // Max k_tlock size (1 KiB per NIP)
	MaxEnvelopeSize         = 256 * 1024      // Max envelope size (256 KiB per NIP)
	MaxTagCount             = 1024            // Max tag count per NIP DoS limits
)

// Default values
const (
	DefaultThreshold     = 3
	DefaultWitnessCount  = 5
	DefaultClockSkewSec  = 300        // 5 minutes
	DefaultMaxInlineSize = 128 * 1024 // 128 KiB
)

// Error messages (NIP Time Capsules v1)
const (
	ErrCapsuleNotFound          = "capsule not found"
	ErrCapsuleAlreadyExists     = "capsule already exists"
	ErrInvalidWitnessCommit     = "invalid witness commitment"
	ErrTooEarly                 = "shares cannot be posted before unlock time"
	ErrInvalidThreshold         = "invalid threshold configuration"
	ErrWitnessNotInCommit       = "witness not in commit"
	ErrInsufficientShares       = "insufficient shares to unlock"
	ErrInvalidEncryption        = "invalid encryption configuration"
	ErrContentTooLarge          = "content exceeds maximum size"
	ErrUnsupportedLocation      = "unsupported storage location"
	ErrInvalidEnvelopeVersion   = "invalid envelope version"
	ErrMissingAAD               = "missing AAD in envelope"
	ErrAADMismatch              = "AAD mismatch"
	ErrInvalidDrandParams       = "invalid drand parameters"
	ErrDrandVerifyFailed        = "drand beacon verification failed"
	ErrInvalidUnlockTag         = "invalid unlock tag format"
	ErrModeEnvelopeIncoherent   = "mode and envelope incoherent"
	ErrInvalidShamirShare       = "invalid Shamir share"
	ErrReconstructionFailed     = "Shamir reconstruction failed"
	ErrInvalidShareIdx          = "invalid share index"
	ErrWitnessSetMismatch       = "witness set mismatch"
	ErrURIHashMismatch          = "URI hash mismatch"
	ErrInvalidFrameSize         = "invalid frame size"
	ErrKTlockEmpty              = "k_tlock cannot be empty for time modes"
	ErrDuplicateInnerP          = "duplicate inner p tag"
	ErrLocationSchemeMismatch   = "location scheme mismatch"
	ErrBareShareNotAllowed      = "bare shares not allowed in threshold modes"
)
