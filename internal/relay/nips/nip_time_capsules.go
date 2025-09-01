package nips

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Shugur-Network/relay/internal/constants"
	nostr "github.com/nbd-wtf/go-nostr"
)

// UnlockConfig represents parsed unlock configuration
type UnlockConfig struct {
	Mode         string
	Threshold    int
	WitnessCount int
	UnlockTime   int64
	Beacon       string
	Round        int64
}

// ContentEnvelope represents the JSON envelope in content field  
type ContentEnvelope struct {
	Version string      `json:"v"`
	CT      string      `json:"ct"`
	KTlock  interface{} `json:"k_tlock"` // can be string or null
	AAD     string      `json:"aad"`
}

// ValidateTimeCapsuleEvent validates time capsule events (kinds 1995, 31995)
func ValidateTimeCapsuleEvent(evt *nostr.Event) error {
	// 1. Validate exactly one unlock tag
	unlockTags := getTagsByName(evt, constants.TagUnlock)
	if len(unlockTags) != 1 {
		if len(unlockTags) == 0 {
			return fmt.Errorf("missing unlock tag")
		}
		return fmt.Errorf("multiple unlock tags not allowed")
	}

	// 2. Parse unlock configuration
	unlockConfig, err := parseUnlockTag(unlockTags[0])
	if err != nil {
		return fmt.Errorf("invalid unlock configuration: %w", err)
	}

	// 3. Validate exactly one enc tag with correct value
	encTags := getTagsByName(evt, constants.TagEncryption)
	if len(encTags) != 1 {
		return fmt.Errorf("exactly one enc tag required")
	}
	if len(encTags[0]) < 2 || encTags[0][1] != constants.EncryptionNIP44v2 {
		return fmt.Errorf("enc tag must be exactly 'nip44:v2'")
	}

	// 4. Validate exactly one loc tag
	locTags := getTagsByName(evt, constants.TagLocation)
	if len(locTags) != 1 {
		return fmt.Errorf("exactly one loc tag required")
	}
	location := locTags[0][1]
	if !isValidLocation(location) {
		return fmt.Errorf("invalid location: %s", location)
	}

	// 5. Validate content envelope
	if err := validateContentEnvelope(evt, unlockConfig); err != nil {
		return fmt.Errorf("content envelope validation failed: %w", err)
	}

	// 6. Validate mode-specific requirements
	switch unlockConfig.Mode {
	case constants.ModeThreshold:
		if err := validateThresholdMode(evt, unlockConfig); err != nil {
			return fmt.Errorf("threshold mode validation failed: %w", err)
		}
	case constants.ModeThresholdTime:
		if err := validateThresholdTimeMode(evt, unlockConfig); err != nil {
			return fmt.Errorf("threshold-time mode validation failed: %w", err)
		}
	case constants.ModeTimelock:
		if err := validateTimelockMode(evt, unlockConfig); err != nil {
			return fmt.Errorf("timelock mode validation failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported mode: %s", unlockConfig.Mode)
	}

	// 7. Validate external storage requirements
	if location != constants.LocationInline {
		if err := validateExternalStorage(evt); err != nil {
			return fmt.Errorf("external storage validation failed: %w", err)
		}
	} else {
		// Validate inline storage restrictions
		if err := validateInlineStorage(evt); err != nil {
			return fmt.Errorf("inline storage validation failed: %w", err)
		}
	}

	// 8. Validate parameterized replaceable events (kind 31995)
	if evt.Kind == constants.KindTimeCapsuleReplaceable {
		if !hasDTag(evt) {
			return fmt.Errorf("missing 'd' tag for parameterized replaceable event")
		}
	}

	return nil
}

// ValidateTimeCapsuleUnlockShare validates unlock share events (kind 1997)
func ValidateTimeCapsuleUnlockShare(evt *nostr.Event) error {
	// 1. Must have exactly one capsule reference (e or a tag)
	eTags := getTagsByName(evt, "e")
	aTags := getTagsByName(evt, "a")
	
	if len(eTags) == 0 && len(aTags) == 0 {
		return fmt.Errorf("missing capsule reference (e or a tag)")
	}
	if len(eTags) > 1 || len(aTags) > 1 {
		return fmt.Errorf("multiple capsule references not allowed")
	}
	if len(eTags) > 0 && len(aTags) > 0 {
		// Both present - they must reference the same capsule (validation in client)
		// For now, we allow both but recommend clients validate coherence
	}

	// 2. Must have exactly one witness p tag
	pTags := getTagsByName(evt, "p")
	if len(pTags) != 1 {
		return fmt.Errorf("exactly one witness p tag required")
	}
	
	// Validate pubkey format
	if len(pTags[0]) < 2 || len(pTags[0][1]) != 64 {
		return fmt.Errorf("invalid witness pubkey format")
	}
	if !isHexString(pTags[0][1]) {
		return fmt.Errorf("witness pubkey must be lowercase hex")
	}

	// 3. Must have exactly one share-idx tag
	shareIdxTags := getTagsByName(evt, constants.TagShareIndex)
	if len(shareIdxTags) != 1 {
		return fmt.Errorf("exactly one share-idx tag required")
	}
	
	shareIdx, err := strconv.Atoi(shareIdxTags[0][1])
	if err != nil || shareIdx < 1 {
		return fmt.Errorf("invalid share index: must be >= 1")
	}

	// 4. Validate content format (Base64 share or NIP-44 v2 encrypted)
	if evt.Content == "" {
		return fmt.Errorf("missing share content")
	}

	// 5. Check for optional inner encryption
	encTags := getTagsByName(evt, constants.TagEncryption)
	if len(encTags) > 1 {
		return fmt.Errorf("multiple enc tags not allowed")
	}
	if len(encTags) == 1 {
		if encTags[0][1] != constants.EncryptionNIP44v2 {
			return fmt.Errorf("inner enc tag must be 'nip44:v2'")
		}
	}

	// 6. Author must equal the witness pubkey
	if evt.PubKey != pTags[0][1] {
		return fmt.Errorf("author must equal witness pubkey")
	}

	return nil
}

// ValidateTimeCapsuleShareDistribution validates share distribution events (kind 1996)
func ValidateTimeCapsuleShareDistribution(evt *nostr.Event) error {
	// 1. Must have exactly one capsule reference (e or a tag)
	eTags := getTagsByName(evt, "e")
	aTags := getTagsByName(evt, "a")
	
	if len(eTags) == 0 && len(aTags) == 0 {
		return fmt.Errorf("missing capsule reference (e or a tag)")
	}
	if len(eTags) > 1 || len(aTags) > 1 {
		return fmt.Errorf("multiple capsule references not allowed")
	}

	// 2. Must have exactly one witness p tag  
	pTags := getTagsByName(evt, "p")
	if len(pTags) != 1 {
		return fmt.Errorf("exactly one witness p tag required")
	}
	
	// Validate pubkey format
	if len(pTags[0]) < 2 || len(pTags[0][1]) != 64 {
		return fmt.Errorf("invalid witness pubkey format")
	}
	if !isHexString(pTags[0][1]) {
		return fmt.Errorf("witness pubkey must be lowercase hex")
	}

	// 3. Must have exactly one share-idx tag
	shareIdxTags := getTagsByName(evt, constants.TagShareIndex)
	if len(shareIdxTags) != 1 {
		return fmt.Errorf("exactly one share-idx tag required")
	}
	
	shareIdx, err := strconv.Atoi(shareIdxTags[0][1])
	if err != nil || shareIdx < 1 {
		return fmt.Errorf("invalid share index: must be >= 1")
	}

	// 4. Content validation (encrypted share)
	if evt.Content == "" {
		return fmt.Errorf("missing encrypted share content")
	}

	// 5. Optional inner encryption validation
	encTags := getTagsByName(evt, constants.TagEncryption)
	if len(encTags) > 1 {
		return fmt.Errorf("multiple enc tags not allowed")
	}
	if len(encTags) == 1 {
		if encTags[0][1] != constants.EncryptionNIP44v2 {
			return fmt.Errorf("inner enc tag must be 'nip44:v2'")
		}
	}

	return nil
}

// Helper functions

func parseUnlockTag(tag nostr.Tag) (*UnlockConfig, error) {
	if len(tag) < 2 {
		return nil, fmt.Errorf("unlock tag too short")
	}

	// Parse space-delimited key/value pairs
	pairs := strings.Fields(strings.Join(tag[1:], " "))
	if len(pairs) == 0 {
		return nil, fmt.Errorf("empty unlock configuration")
	}

	config := &UnlockConfig{}
	keyValues := make(map[string]string)

	// Parse key/value pairs
	for i := 0; i < len(pairs)-1; i += 2 {
		key := pairs[i]
		value := pairs[i+1]
		keyValues[key] = value
	}

	// Extract mode (required)
	modeValue, hasMode := keyValues["mode"]
	if !hasMode {
		return nil, fmt.Errorf("missing mode")
	}
	config.Mode = modeValue

	// Validate mode and extract mode-specific parameters
	switch config.Mode {
	case constants.ModeThreshold:
		// Required: t, n
		// Forbidden: T, beacon, round
		if err := extractThresholdParams(keyValues, config); err != nil {
			return nil, err
		}
		if hasTimeParams(keyValues) {
			return nil, fmt.Errorf("threshold mode must not have time parameters")
		}

	case constants.ModeThresholdTime:
		// Required: t, n, T, beacon, round
		if err := extractThresholdParams(keyValues, config); err != nil {
			return nil, err
		}
		if err := extractTimeParams(keyValues, config); err != nil {
			return nil, err
		}

	case constants.ModeTimelock:
		// Required: T, beacon, round
		// Forbidden: t, n
		if hasThresholdParams(keyValues) {
			return nil, fmt.Errorf("timelock mode must not have threshold parameters")
		}
		if err := extractTimeParams(keyValues, config); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("invalid mode: %s", config.Mode)
	}

	return config, nil
}

func extractThresholdParams(kv map[string]string, config *UnlockConfig) error {
	tStr, hasT := kv["t"]
	nStr, hasN := kv["n"]

	if !hasT || !hasN {
		return fmt.Errorf("threshold mode requires t and n parameters")
	}

	t, err := strconv.Atoi(tStr)
	if err != nil || t < 1 {
		return fmt.Errorf("invalid threshold t: %s", tStr)
	}

	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 {
		return fmt.Errorf("invalid witness count n: %s", nStr)
	}

	if t > n {
		return fmt.Errorf("threshold t cannot exceed witness count n")
	}

	if t > constants.MaxThresholdValue || n > constants.MaxWitnessCount {
		return fmt.Errorf("threshold or witness count exceeds maximum limits")
	}

	config.Threshold = t
	config.WitnessCount = n
	return nil
}

func extractTimeParams(kv map[string]string, config *UnlockConfig) error {
	TStr, hasT := kv["T"]
	beaconStr, hasBeacon := kv["beacon"]
	roundStr, hasRound := kv["round"]

	if !hasT || !hasBeacon || !hasRound {
		return fmt.Errorf("time mode requires T, beacon, and round parameters")
	}

	T, err := strconv.ParseInt(TStr, 10, 64)
	if err != nil || T <= 0 {
		return fmt.Errorf("invalid unlock time T: %s", TStr)
	}

	round, err := strconv.ParseInt(roundStr, 10, 64)
	if err != nil || round < 1 {
		return fmt.Errorf("invalid round: %s", roundStr)
	}

	// Validate beacon format (32-byte hex)
	if len(beaconStr) != 64 || !isHexString(beaconStr) {
		return fmt.Errorf("invalid beacon format: must be 64 lowercase hex chars")
	}

	config.UnlockTime = T
	config.Beacon = beaconStr
	config.Round = round
	return nil
}

func hasThresholdParams(kv map[string]string) bool {
	_, hasT := kv["t"]
	_, hasN := kv["n"]
	return hasT || hasN
}

func hasTimeParams(kv map[string]string) bool {
	_, hasT := kv["T"]
	_, hasBeacon := kv["beacon"]
	_, hasRound := kv["round"]
	return hasT || hasBeacon || hasRound
}

func validateContentEnvelope(evt *nostr.Event, config *UnlockConfig) error {
	// Parse envelope JSON
	var envelope ContentEnvelope
	if err := json.Unmarshal([]byte(evt.Content), &envelope); err != nil {
		return fmt.Errorf("content must be valid JSON envelope")
	}

	// 1. Validate version
	if envelope.Version != constants.EnvelopeVersion {
		return fmt.Errorf("invalid envelope version: %s", envelope.Version)
	}

	// 2. Validate AAD
	if envelope.AAD == "" {
		return fmt.Errorf("missing AAD in envelope")
	}
	if len(envelope.AAD) != 64 || !isHexString(envelope.AAD) {
		return fmt.Errorf("AAD must be 64 lowercase hex chars")
	}

	// 3. Validate mode/envelope coherence
	switch config.Mode {
	case constants.ModeThreshold:
		// k_tlock must be null
		if envelope.KTlock != nil {
			return fmt.Errorf("threshold mode: k_tlock must be null")
		}
	case constants.ModeThresholdTime, constants.ModeTimelock:
		// k_tlock must not be null and must decode to non-empty bytes
		if envelope.KTlock == nil {
			return fmt.Errorf("time mode: k_tlock cannot be null")
		}
		ktlockStr, ok := envelope.KTlock.(string)
		if !ok {
			return fmt.Errorf("k_tlock must be string")
		}
		if ktlockStr == "" {
			return fmt.Errorf("k_tlock cannot be empty string")
		}
		// Validate it's valid Base64 and decodes to non-empty
		if !isValidBase64(ktlockStr) {
			return fmt.Errorf("k_tlock must be valid Base64")
		}
	}

	return nil
}

func validateThresholdMode(evt *nostr.Event, config *UnlockConfig) error {
	// 1. Must have witnesses
	witnesses := getTagsByName(evt, "p")
	if len(witnesses) == 0 {
		return fmt.Errorf("threshold mode requires witnesses")
	}
	if len(witnesses) != config.WitnessCount {
		return fmt.Errorf("witness count mismatch: expected %d, got %d", config.WitnessCount, len(witnesses))
	}

	// 2. Must have witness commitment
	commitTags := getTagsByName(evt, constants.TagWitnessCommit)
	if len(commitTags) != 1 {
		return fmt.Errorf("exactly one w-commit tag required")
	}
	if err := validateWitnessCommit(commitTags[0]); err != nil {
		return fmt.Errorf("invalid witness commitment: %w", err)
	}

	// 3. Must not have time-related tags
	if hasAnyTimeTag(evt) {
		return fmt.Errorf("threshold mode must not have time-related tags")
	}

	return nil
}

func validateThresholdTimeMode(evt *nostr.Event, config *UnlockConfig) error {
	// 1. Must have witnesses (same as threshold)
	witnesses := getTagsByName(evt, "p")
	if len(witnesses) == 0 {
		return fmt.Errorf("threshold-time mode requires witnesses")
	}
	if len(witnesses) != config.WitnessCount {
		return fmt.Errorf("witness count mismatch: expected %d, got %d", config.WitnessCount, len(witnesses))
	}

	// 2. Must have witness commitment
	commitTags := getTagsByName(evt, constants.TagWitnessCommit)
	if len(commitTags) != 1 {
		return fmt.Errorf("exactly one w-commit tag required")
	}
	if err := validateWitnessCommit(commitTags[0]); err != nil {
		return fmt.Errorf("invalid witness commitment: %w", err)
	}

	// 3. Time validation is done in parseUnlockTag

	return nil
}

func validateTimelockMode(evt *nostr.Event, config *UnlockConfig) error {
	// 1. Must not have witnesses
	witnesses := getTagsByName(evt, "p")
	if len(witnesses) > 0 {
		return fmt.Errorf("timelock mode must not have witnesses")
	}

	// 2. Must not have witness commitment
	commitTags := getTagsByName(evt, constants.TagWitnessCommit)
	if len(commitTags) > 0 {
		return fmt.Errorf("timelock mode must not have witness commitment")
	}

	// 3. Time validation is done in parseUnlockTag

	return nil
}

func validateExternalStorage(evt *nostr.Event) error {
	// 1. Must have URI tag
	uriTags := getTagsByName(evt, constants.TagURI)
	if len(uriTags) != 1 {
		return fmt.Errorf("exactly one uri tag required for external storage")
	}
	uri := uriTags[0][1]
	if uri == "" {
		return fmt.Errorf("uri cannot be empty")
	}

	// 2. Must have SHA256 tag
	sha256Tags := getTagsByName(evt, constants.TagSHA256)
	if len(sha256Tags) != 1 {
		return fmt.Errorf("exactly one sha256 tag required for external storage")
	}
	hash := sha256Tags[0][1]
	if len(hash) != 64 || !isHexString(hash) {
		return fmt.Errorf("sha256 must be 64 lowercase hex chars")
	}

	// 3. Validate URI scheme matches location
	locTags := getTagsByName(evt, constants.TagLocation)
	location := locTags[0][1]
	if !validateURIScheme(uri, location) {
		return fmt.Errorf("URI scheme must match location")
	}

	// 4. Content envelope ct must be empty string
	var envelope ContentEnvelope
	if err := json.Unmarshal([]byte(evt.Content), &envelope); err == nil {
		if envelope.CT != "" {
			return fmt.Errorf("external storage: envelope ct must be empty string")
		}
	}

	return nil
}

func validateInlineStorage(evt *nostr.Event) error {
	// 1. Must not have URI or SHA256 tags
	uriTags := getTagsByName(evt, constants.TagURI)
	sha256Tags := getTagsByName(evt, constants.TagSHA256)
	
	if len(uriTags) > 0 {
		return fmt.Errorf("inline storage must not have uri tag")
	}
	if len(sha256Tags) > 0 {
		return fmt.Errorf("inline storage must not have sha256 tag")
	}

	// 2. Content envelope ct must be non-empty
	var envelope ContentEnvelope
	if err := json.Unmarshal([]byte(evt.Content), &envelope); err == nil {
		if envelope.CT == "" {
			return fmt.Errorf("inline storage: envelope ct cannot be empty")
		}
	}

	return nil
}

func validateWitnessCommit(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("w-commit tag too short")
	}
	
	commit := tag[1]
	if !strings.HasPrefix(commit, "sha256:") {
		return fmt.Errorf("w-commit must have sha256: prefix")
	}
	
	hashHex := strings.TrimPrefix(commit, "sha256:")
	if len(hashHex) != 64 || !isHexString(hashHex) {
		return fmt.Errorf("w-commit hash must be 64 lowercase hex chars")
	}
	
	return nil
}

func validateURIScheme(uri, location string) bool {
	switch location {
	case constants.LocationHTTPS:
		return strings.HasPrefix(uri, "https://")
	case constants.LocationIPFS:
		return strings.HasPrefix(uri, "ipfs://")
	case constants.LocationBlossom:
		return strings.HasPrefix(uri, "blossom://")
	default:
		return false
	}
}

func isValidLocation(loc string) bool {
	validLocations := []string{
		constants.LocationInline,
		constants.LocationHTTPS,
		constants.LocationIPFS,
		constants.LocationBlossom,
	}
	
	for _, valid := range validLocations {
		if loc == valid {
			return true
		}
	}
	return false
}

func hasAnyTimeTag(evt *nostr.Event) bool {
	timeTagNames := []string{constants.TagBeacon, constants.TagRound, "T"}
	for _, tagName := range timeTagNames {
		if len(getTagsByName(evt, tagName)) > 0 {
			return true
		}
	}
	return false
}

func getTagsByName(evt *nostr.Event, name string) []nostr.Tag {
	var tags []nostr.Tag
	for _, tag := range evt.Tags {
		if len(tag) > 0 && tag[0] == name {
			tags = append(tags, tag)
		}
	}
	return tags
}

func hasDTag(evt *nostr.Event) bool {
	return len(getTagsByName(evt, "d")) > 0
}

func isHexString(s string) bool {
	if len(s)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func isValidBase64(s string) bool {
	// Basic Base64 validation - just check it can decode
	// More thorough validation would check RFC 4648 compliance
	matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]*={0,2}$`, s)
	return matched
}

// IsTimeCapsuleEvent checks if an event is a time capsule
func IsTimeCapsuleEvent(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsule ||
		evt.Kind == constants.KindTimeCapsuleReplaceable
}

// IsTimeCapsuleUnlockShare checks if an event is an unlock share
func IsTimeCapsuleUnlockShare(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsuleUnlockShare
}

// IsTimeCapsuleShareDistribution checks if an event is a share distribution
func IsTimeCapsuleShareDistribution(evt *nostr.Event) bool {
	return evt.Kind == constants.KindTimeCapsuleShareDistribution
}
