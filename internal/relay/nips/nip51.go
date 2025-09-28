package nips

import (
	"fmt"
	"strings"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-51: Lists
// https://github.com/nostr-protocol/nips/blob/master/51.md
//
// Standard lists (single list per kind):
// - 3: Follow list (NIP-02) 
// - 10000: Mute list
// - 10001: Pinned notes
// - 10002: Read/write relays (NIP-65)
// - 10003: Bookmarks
// - 10004: Communities
// - 10005: Public chats
// - 10006: Blocked relays
// - 10007: Search relays
// - 10009: Simple groups
// - 10012: Relay feeds
// - 10015: Interests
// - 10020: Media follows
// - 10030: Emojis
// - 10050: DM relays
// - 10101: Good wiki authors
// - 10102: Good wiki relays
//
// Sets (multiple lists per kind with 'd' identifier):
// - 30000: Follow sets
// - 30001: Generic lists (deprecated)
// - 30002: Relay sets  
// - 30003: Bookmark sets
// - 30004: Curation sets (articles/notes)
// - 30005: Curation sets (videos)
// - 30007: Kind mute sets
// - 30015: Interest sets
// - 30030: Emoji sets
// - 30063: Release artifact sets
// - 30267: App curation sets
// - 31924: Calendar
// - 39089: Starter packs
// - 39092: Media starter packs

// ValidateList validates NIP-51 list events
func ValidateList(evt *nostr.Event) error {
	logger.Debug("NIP-51: Validating list event",
		zap.String("event_id", evt.ID),
		zap.Int("kind", evt.Kind))

	// Check if this is a valid list kind
	if !IsListKind(evt.Kind) {
		return fmt.Errorf("invalid event kind for list: %d", evt.Kind)
	}

	// Validate based on whether it's a standard list or set
	if IsStandardListKind(evt.Kind) {
		return validateStandardList(evt)
	} else if IsSetKind(evt.Kind) {
		return validateSet(evt)
	}

	return fmt.Errorf("unsupported list kind: %d", evt.Kind)
}

// validateStandardList validates standard list events (single per kind)
func validateStandardList(evt *nostr.Event) error {
	switch evt.Kind {
	case 3:
		// Follow list validation is handled by NIP-02
		return ValidateFollowList(evt)
	case 10000:
		return validateMuteList(evt)
	case 10001:
		return validatePinnedNotes(evt)
	case 10002:
		// Relay list validation is handled by NIP-65
		return ValidateKind10002(*evt)
	case 10003:
		return validateBookmarks(evt)
	case 10004:
		return validateCommunities(evt)
	case 10005:
		return validatePublicChats(evt)
	case 10006:
		return validateBlockedRelays(evt)
	case 10007:
		return validateSearchRelays(evt)
	case 10009:
		return validateSimpleGroups(evt)
	case 10012:
		return validateRelayFeeds(evt)
	case 10015:
		return validateInterests(evt)
	case 10020:
		return validateMediaFollows(evt)
	case 10030:
		return validateEmojis(evt)
	case 10050:
		return validateDMRelays(evt)
	case 10101:
		return validateGoodWikiAuthors(evt)
	case 10102:
		return validateGoodWikiRelays(evt)
	default:
		return fmt.Errorf("unsupported standard list kind: %d", evt.Kind)
	}
}

// validateSet validates set events (multiple per kind with 'd' identifier)
func validateSet(evt *nostr.Event) error {
	// All sets require a 'd' tag for identification
	hasDTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			hasDTag = true
			break
		}
	}

	if !hasDTag {
		return fmt.Errorf("set must have 'd' tag for identification")
	}

	// Validate specific set types
	switch evt.Kind {
	case 30000:
		return validateFollowSet(evt)
	case 30001:
		return validateGenericSet(evt) // Deprecated but still validate
	case 30002:
		return validateRelaySet(evt)
	case 30003:
		return validateBookmarkSet(evt)
	case 30004:
		return validateCurationSet(evt)
	case 30005:
		return validateVideoCurationSet(evt)
	case 30007:
		return validateKindMuteSet(evt)
	case 30015:
		return validateInterestSet(evt)
	case 30030:
		return validateEmojiSet(evt)
	case 30063:
		return validateReleaseArtifactSet(evt)
	case 30267:
		return validateAppCurationSet(evt)
	case 31924:
		return validateCalendar(evt)
	case 39089:
		return validateStarterPack(evt)
	case 39092:
		return validateMediaStarterPack(evt)
	default:
		return fmt.Errorf("unsupported set kind: %d", evt.Kind)
	}
}

// Standard List Validators

func validateMuteList(evt *nostr.Event) error {
	// Can contain public and private items
	// Public: "p" (pubkeys), "t" (hashtags), "word" (lowercase strings), "e" (threads)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "p":
				if len(tag[1]) != 64 || !isHexString(tag[1]) {
					return fmt.Errorf("invalid pubkey in mute list: %s", tag[1])
				}
			case "t":
				if tag[1] == "" {
					return fmt.Errorf("empty hashtag in mute list")
				}
			case "word":
				if tag[1] == "" {
					return fmt.Errorf("empty word in mute list")
				}
				// Validate lowercase requirement
				if strings.ToLower(tag[1]) != tag[1] {
					return fmt.Errorf("word must be lowercase: %s", tag[1])
				}
			case "e":
				if len(tag[1]) != 64 || !isHexString(tag[1]) {
					return fmt.Errorf("invalid event ID in mute list: %s", tag[1])
				}
			}
		}
	}

	// Validate encrypted content if present
	if evt.Content != "" {
		if err := validateEncryptedListContent(evt); err != nil {
			return fmt.Errorf("invalid encrypted content in mute list: %v", err)
		}
	}

	return nil
}

func validatePinnedNotes(evt *nostr.Event) error {
	// Must contain "e" tags referencing kind:1 notes
	hasEventTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			hasEventTag = true
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid event ID in pinned notes: %s", tag[1])
			}
		}
	}

	if !hasEventTag && evt.Content == "" {
		return fmt.Errorf("pinned notes must have event references or encrypted content")
	}

	return nil
}

func validateBookmarks(evt *nostr.Event) error {
	// Can contain "e" (kind:1 notes), "a" (kind:30023 articles), "t" (hashtags), "r" (URLs)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e":
				if len(tag[1]) != 64 || !isHexString(tag[1]) {
					return fmt.Errorf("invalid event ID in bookmarks: %s", tag[1])
				}
			case "a":
				if err := validateAddressableReference(tag[1]); err != nil {
					return fmt.Errorf("invalid addressable reference in bookmarks: %v", err)
				}
			case "t":
				if tag[1] == "" {
					return fmt.Errorf("empty hashtag in bookmarks")
				}
			case "r":
				if !strings.HasPrefix(tag[1], "http://") && !strings.HasPrefix(tag[1], "https://") {
					return fmt.Errorf("invalid URL in bookmarks: %s", tag[1])
				}
			}
		}
	}
	return nil
}

func validateCommunities(evt *nostr.Event) error {
	// Must contain "a" tags referencing kind:34550 community definitions
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "a" {
			if err := validateAddressableReference(tag[1]); err != nil {
				return fmt.Errorf("invalid community reference: %v", err)
			}
			// Should reference kind 34550
			parts := strings.Split(tag[1], ":")
			if len(parts) >= 1 && parts[0] != "34550" {
				return fmt.Errorf("community reference must be kind 34550, got: %s", parts[0])
			}
		}
	}
	return nil
}

func validatePublicChats(evt *nostr.Event) error {
	// Must contain "e" tags referencing kind:40 channel definitions
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid channel reference: %s", tag[1])
			}
		}
	}
	return nil
}

func validateBlockedRelays(evt *nostr.Event) error {
	// Must contain "relay" tags with relay URLs
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			if !isWebSocketURL(tag[1]) {
				return fmt.Errorf("invalid relay URL: %s", tag[1])
			}
		}
	}
	return nil
}

func validateSearchRelays(evt *nostr.Event) error {
	// Must contain "relay" tags with relay URLs
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			if !isWebSocketURL(tag[1]) {
				return fmt.Errorf("invalid relay URL in search relays: %s", tag[1])
			}
		}
	}
	return nil
}

func validateSimpleGroups(evt *nostr.Event) error {
	// Must contain "group" tags (NIP-29 group id + relay URL + optional name)
	// and "r" tags for each relay in use
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "group":
				// Basic validation - should contain group identifier
				if tag[1] == "" {
					return fmt.Errorf("empty group identifier")
				}
			case "r":
				if !isWebSocketURL(tag[1]) {
					return fmt.Errorf("invalid relay URL in simple groups: %s", tag[1])
				}
			}
		}
	}
	return nil
}

func validateRelayFeeds(evt *nostr.Event) error {
	// Can contain "relay" tags and "a" tags referencing kind:30002 relay sets
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "relay":
				if !isWebSocketURL(tag[1]) {
					return fmt.Errorf("invalid relay URL in relay feeds: %s", tag[1])
				}
			case "a":
				if err := validateAddressableReference(tag[1]); err != nil {
					return fmt.Errorf("invalid relay set reference: %v", err)
				}
			}
		}
	}
	return nil
}

func validateInterests(evt *nostr.Event) error {
	// Can contain "t" (hashtags) and "a" (kind:30015 interest sets)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "t":
				if tag[1] == "" {
					return fmt.Errorf("empty hashtag in interests")
				}
			case "a":
				if err := validateAddressableReference(tag[1]); err != nil {
					return fmt.Errorf("invalid interest set reference: %v", err)
				}
			}
		}
	}
	return nil
}

func validateMediaFollows(evt *nostr.Event) error {
	// Same format as follow list but for multimedia content
	return validateFollowListFormat(evt)
}

func validateEmojis(evt *nostr.Event) error {
	// Can contain "emoji" tags and "a" tags referencing kind:30030 emoji sets
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "emoji":
				// Basic emoji validation - should not be empty
				if tag[1] == "" {
					return fmt.Errorf("empty emoji reference")
				}
			case "a":
				if err := validateAddressableReference(tag[1]); err != nil {
					return fmt.Errorf("invalid emoji set reference: %v", err)
				}
			}
		}
	}
	return nil
}

func validateDMRelays(evt *nostr.Event) error {
	// Must contain "relay" tags for NIP-17 direct messages
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			if !isWebSocketURL(tag[1]) {
				return fmt.Errorf("invalid DM relay URL: %s", tag[1])
			}
		}
	}
	return nil
}

func validateGoodWikiAuthors(evt *nostr.Event) error {
	// Must contain "p" tags referencing pubkeys
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid pubkey in wiki authors: %s", tag[1])
			}
		}
	}
	return nil
}

func validateGoodWikiRelays(evt *nostr.Event) error {
	// Must contain "relay" tags
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			if !isWebSocketURL(tag[1]) {
				return fmt.Errorf("invalid wiki relay URL: %s", tag[1])
			}
		}
	}
	return nil
}

// Set Validators

func validateFollowSet(evt *nostr.Event) error {
	// Must contain "p" tags with pubkeys
	hasPTag := false
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			hasPTag = true
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid pubkey in follow set: %s", tag[1])
			}
		}
	}

	if !hasPTag && evt.Content == "" {
		return fmt.Errorf("follow set must have pubkey references or encrypted content")
	}

	return validateSetMetadata(evt)
}

func validateGenericSet(evt *nostr.Event) error {
	// Deprecated format - just ensure basic set requirements
	logger.Warn("NIP-51: Using deprecated generic set format (kind 30001)",
		zap.String("event_id", evt.ID))
	return validateSetMetadata(evt)
}

func validateRelaySet(evt *nostr.Event) error {
	// Must contain "relay" tags
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "relay" {
			if !isWebSocketURL(tag[1]) {
				return fmt.Errorf("invalid relay URL in set: %s", tag[1])
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateBookmarkSet(evt *nostr.Event) error {
	// Same validation as standard bookmarks
	if err := validateBookmarks(evt); err != nil {
		return err
	}
	return validateSetMetadata(evt)
}

func validateCurationSet(evt *nostr.Event) error {
	// Must contain "a" (kind:30023 articles) and/or "e" (kind:1 notes)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "a":
				if err := validateAddressableReference(tag[1]); err != nil {
					return fmt.Errorf("invalid article reference: %v", err)
				}
			case "e":
				if len(tag[1]) != 64 || !isHexString(tag[1]) {
					return fmt.Errorf("invalid note reference: %s", tag[1])
				}
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateVideoCurationSet(evt *nostr.Event) error {
	// Must contain "e" tags referencing kind:21 videos  
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid video reference: %s", tag[1])
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateKindMuteSet(evt *nostr.Event) error {
	// 'd' tag MUST be the kind string
	dTagValue := ""
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			dTagValue = tag[1]
			break
		}
	}

	if dTagValue == "" {
		return fmt.Errorf("kind mute set must have 'd' tag with kind string")
	}

	// Must contain "p" tags for pubkeys
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid pubkey in kind mute set: %s", tag[1])
			}
		}
	}

	return validateSetMetadata(evt)
}

func validateInterestSet(evt *nostr.Event) error {
	// Must contain "t" tags with hashtags
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "t" {
			if tag[1] == "" {
				return fmt.Errorf("empty hashtag in interest set")
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateEmojiSet(evt *nostr.Event) error {
	// Must contain "emoji" tags
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "emoji" {
			if tag[1] == "" {
				return fmt.Errorf("empty emoji in set")
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateReleaseArtifactSet(evt *nostr.Event) error {
	// Must contain "e" tags (kind:1063 file metadata) and/or "a" tags (software application)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e":
				if len(tag[1]) != 64 || !isHexString(tag[1]) {
					return fmt.Errorf("invalid file metadata reference: %s", tag[1])
				}
			case "a":
				if err := validateAddressableReference(tag[1]); err != nil {
					return fmt.Errorf("invalid software application reference: %v", err)
				}
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateAppCurationSet(evt *nostr.Event) error {
	// Must contain "a" tags referencing software application events
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "a" {
			if err := validateAddressableReference(tag[1]); err != nil {
				return fmt.Errorf("invalid app reference: %v", err)
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateCalendar(evt *nostr.Event) error {
	// Must contain "a" tags referencing calendar event events
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "a" {
			if err := validateAddressableReference(tag[1]); err != nil {
				return fmt.Errorf("invalid calendar event reference: %v", err)
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateStarterPack(evt *nostr.Event) error {
	// Must contain "p" tags with pubkeys
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid pubkey in starter pack: %s", tag[1])
			}
		}
	}
	return validateSetMetadata(evt)
}

func validateMediaStarterPack(evt *nostr.Event) error {
	// Same as starter pack but for multimedia clients
	return validateStarterPack(evt)
}

// Helper Functions

func validateSetMetadata(evt *nostr.Event) error {
	// Sets can have optional metadata tags
	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "title":
				if len(tag[1]) > 200 {
					return fmt.Errorf("set title too long (max 200 characters)")
				}
			case "description":
				if len(tag[1]) > 1000 {
					return fmt.Errorf("set description too long (max 1000 characters)")
				}
			case "image":
				if !strings.HasPrefix(tag[1], "http://") && !strings.HasPrefix(tag[1], "https://") {
					return fmt.Errorf("invalid image URL in set metadata: %s", tag[1])
				}
			}
		}
	}
	return nil
}

func validateEncryptedListContent(evt *nostr.Event) error {
	// Basic validation that content could be encrypted
	if evt.Content == "" {
		return nil
	}

	// Check if it looks like NIP-44 encrypted content (base64)
	if len(evt.Content) < 20 {
		return fmt.Errorf("encrypted content too short")
	}

	// We can't decrypt here without private key, so just validate format
	// The content should be base64 encoded encrypted data
	return nil
}

func validateFollowListFormat(evt *nostr.Event) error {
	// Validate "p" tags format for follow lists
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			if len(tag[1]) != 64 || !isHexString(tag[1]) {
				return fmt.Errorf("invalid pubkey format: %s", tag[1])
			}
			// Optional relay hint and petname in positions 2 and 3
			if len(tag) >= 3 && tag[2] != "" && !isWebSocketURL(tag[2]) {
				return fmt.Errorf("invalid relay hint: %s", tag[2])
			}
		}
	}
	return nil
}

func validateAddressableReference(ref string) error {
	// Format: kind:pubkey:dvalue
	parts := strings.Split(ref, ":")
	if len(parts) != 3 {
		return fmt.Errorf("addressable reference must have format kind:pubkey:dvalue")
	}

	// Validate kind is numeric
	if parts[0] == "" {
		return fmt.Errorf("kind cannot be empty")
	}

	// Validate pubkey format
	if len(parts[1]) != 64 || !isHexString(parts[1]) {
		return fmt.Errorf("invalid pubkey in addressable reference")
	}

	// dvalue can be empty or any string
	return nil
}



// Helper function for validating WebSocket URLs
func isWebSocketURL(url string) bool {
	return strings.HasPrefix(url, "ws://") || strings.HasPrefix(url, "wss://")
}

func IsListKind(kind int) bool {
	return IsStandardListKind(kind) || IsSetKind(kind)
}

func IsStandardListKind(kind int) bool {
	standardKinds := map[int]bool{
		3:     true, // Follow list (NIP-02)
		10000: true, // Mute list
		10001: true, // Pinned notes
		10002: true, // Read/write relays (NIP-65)
		10003: true, // Bookmarks
		10004: true, // Communities
		10005: true, // Public chats
		10006: true, // Blocked relays
		10007: true, // Search relays
		10009: true, // Simple groups
		10012: true, // Relay feeds
		10015: true, // Interests
		10020: true, // Media follows
		10030: true, // Emojis
		10050: true, // DM relays
		10101: true, // Good wiki authors
		10102: true, // Good wiki relays
	}
	return standardKinds[kind]
}

func IsSetKind(kind int) bool {
	setKinds := map[int]bool{
		30000: true, // Follow sets
		30001: true, // Generic lists (deprecated)
		30002: true, // Relay sets
		30003: true, // Bookmark sets
		30004: true, // Curation sets (articles/notes)
		30005: true, // Curation sets (videos)
		30007: true, // Kind mute sets
		30015: true, // Interest sets
		30030: true, // Emoji sets
		30063: true, // Release artifact sets
		30267: true, // App curation sets
		31924: true, // Calendar
		39089: true, // Starter packs
		39092: true, // Media starter packs
	}
	return setKinds[kind]
}

func GetListType(kind int) string {
	switch kind {
	case 3:
		return "follow_list"
	case 10000:
		return "mute_list"
	case 10001:
		return "pinned_notes"
	case 10002:
		return "relay_list"
	case 10003:
		return "bookmarks"
	case 10004:
		return "communities"
	case 10005:
		return "public_chats"
	case 10006:
		return "blocked_relays"
	case 10007:
		return "search_relays"
	case 10009:
		return "simple_groups"
	case 10012:
		return "relay_feeds"
	case 10015:
		return "interests"
	case 10020:
		return "media_follows"
	case 10030:
		return "emojis"
	case 10050:
		return "dm_relays"
	case 10101:
		return "good_wiki_authors"
	case 10102:
		return "good_wiki_relays"
	case 30000:
		return "follow_set"
	case 30001:
		return "generic_set"
	case 30002:
		return "relay_set"
	case 30003:
		return "bookmark_set"
	case 30004:
		return "curation_set"
	case 30005:
		return "video_curation_set"
	case 30007:
		return "kind_mute_set"
	case 30015:
		return "interest_set"
	case 30030:
		return "emoji_set"
	case 30063:
		return "release_artifact_set"
	case 30267:
		return "app_curation_set"
	case 31924:
		return "calendar"
	case 39089:
		return "starter_pack"
	case 39092:
		return "media_starter_pack"
	default:
		return "unknown"
	}
}

// IsListEvent checks if an event is a list event
func IsListEvent(evt *nostr.Event) bool {
	return IsListKind(evt.Kind)
}

func isHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}