package nips

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// ValidateWikiArticle validates NIP-54 wiki article events (kind 30818)
func ValidateWikiArticle(event *nostr.Event) error {
	if event.Kind != 30818 {
		return fmt.Errorf("event kind must be 30818 for wiki articles")
	}

	// Validate basic event structure
	if err := validateBasicEvent(event); err != nil {
		return fmt.Errorf("invalid basic event structure: %w", err)
	}

	// Validate tags
	if err := validateWikiArticleTags(event); err != nil {
		return fmt.Errorf("invalid wiki article tags: %w", err)
	}

	// Validate content format (should be Asciidoc with wikilinks and nostr links)
	if err := validateWikiContent(event.Content); err != nil {
		return fmt.Errorf("invalid wiki content: %w", err)
	}

	return nil
}

// ValidateMergeRequest validates NIP-54 merge request events (kind 818)
func ValidateMergeRequest(event *nostr.Event) error {
	if event.Kind != 818 {
		return fmt.Errorf("event kind must be 818 for merge requests")
	}

	// Validate basic event structure
	if err := validateBasicEvent(event); err != nil {
		return fmt.Errorf("invalid basic event structure: %w", err)
	}

	// Validate tags
	if err := validateMergeRequestTags(event); err != nil {
		return fmt.Errorf("invalid merge request tags: %w", err)
	}

	return nil
}

// ValidateWikiRedirect validates NIP-54 wiki redirect events (kind 30819)
func ValidateWikiRedirect(event *nostr.Event) error {
	if event.Kind != 30819 {
		return fmt.Errorf("event kind must be 30819 for wiki redirects")
	}

	// Validate basic event structure
	if err := validateBasicEvent(event); err != nil {
		return fmt.Errorf("invalid basic event structure: %w", err)
	}

	// Validate tags
	if err := validateWikiRedirectTags(event); err != nil {
		return fmt.Errorf("invalid wiki redirect tags: %w", err)
	}

	return nil
}

// validateWikiArticleTags validates tags for wiki article events
func validateWikiArticleTags(event *nostr.Event) error {
	var hasDTag bool
	
	for _, tag := range event.Tags {
		switch tag[0] {
		case "d":
			if err := validateWikiDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "title":
			if err := validateWikiTitleTag(tag); err != nil {
				return err
			}
		case "summary":
			if err := validateWikiSummaryTag(tag); err != nil {
				return err
			}
		case "a":
			if err := validateEventAddressTag(tag); err != nil {
				return fmt.Errorf("invalid referenced event address: %w", err)
			}
		case "e":
			if err := validateEventIDTag(tag); err != nil {
				return fmt.Errorf("invalid referenced event ID: %w", err)
			}
		case "p":
			if err := validatePubkeyTag(tag); err != nil {
				return fmt.Errorf("invalid pubkey reference: %w", err)
			}
		case "t":
			if err := validateHashtagTag(tag); err != nil {
				return fmt.Errorf("invalid hashtag: %w", err)
			}
		default:
			// Other tags are allowed but not validated
		}
	}

	// Ensure required tags are present
	if !hasDTag {
		return fmt.Errorf("wiki article must have a 'd' tag")
	}

	return nil
}

// validateMergeRequestTags validates tags for merge request events
func validateMergeRequestTags(event *nostr.Event) error {
	var hasATag, hasPTag, hasSourceETag bool
	var eTagCount int
	
	for _, tag := range event.Tags {
		switch tag[0] {
		case "a":
			if err := validateMergeRequestATag(tag); err != nil {
				return err
			}
			hasATag = true
		case "e":
			if err := validateMergeRequestETag(tag); err != nil {
				return err
			}
			eTagCount++
			// Check if this is the source e tag
			if len(tag) >= 4 && tag[3] == "source" {
				hasSourceETag = true
			}
		case "p":
			if err := validatePubkeyTag(tag); err != nil {
				return fmt.Errorf("invalid destination pubkey: %w", err)
			}
			hasPTag = true
		default:
			// Other tags are allowed but not validated
		}
	}

	// Ensure required tags are present
	if !hasATag {
		return fmt.Errorf("merge request must have an 'a' tag referencing the target article")
	}
	
	if !hasPTag {
		return fmt.Errorf("merge request must have a 'p' tag referencing the destination pubkey")
	}
	
	if !hasSourceETag {
		return fmt.Errorf("merge request must have an 'e' tag with 'source' marker")
	}

	if eTagCount < 1 {
		return fmt.Errorf("merge request must have at least one 'e' tag")
	}

	return nil
}

// validateWikiRedirectTags validates tags for wiki redirect events
func validateWikiRedirectTags(event *nostr.Event) error {
	var hasDTag, hasRedirectTag bool
	
	for _, tag := range event.Tags {
		switch tag[0] {
		case "d":
			if err := validateWikiDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "redirect":
			if err := validateWikiRedirectTag(tag); err != nil {
				return err
			}
			hasRedirectTag = true
		case "title":
			if err := validateWikiTitleTag(tag); err != nil {
				return err
			}
		case "summary":
			if err := validateWikiSummaryTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed but not validated
		}
	}

	// Ensure required tags are present
	if !hasDTag {
		return fmt.Errorf("wiki redirect must have a 'd' tag")
	}
	
	if !hasRedirectTag {
		return fmt.Errorf("wiki redirect must have a 'redirect' tag")
	}

	return nil
}

// validateWikiDTag validates the d tag for wiki articles and redirects
func validateWikiDTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("d tag must have exactly 2 elements")
	}

	dValue := tag[1]
	if dValue == "" {
		return fmt.Errorf("d tag value cannot be empty")
	}

	// Validate normalization rules:
	// - All letters must be lowercase
	// - Any non-letter character must be converted to a dash
	normalizedValue := normalizeDTag(dValue)
	if dValue != normalizedValue {
		return fmt.Errorf("d tag value '%s' is not properly normalized (should be '%s')", dValue, normalizedValue)
	}

	// Additional validation: must contain at least one letter
	if !regexp.MustCompile(`[a-z]`).MatchString(dValue) {
		return fmt.Errorf("d tag value must contain at least one letter")
	}

	// Must not start or end with dashes
	if strings.HasPrefix(dValue, "-") || strings.HasSuffix(dValue, "-") {
		return fmt.Errorf("d tag value cannot start or end with dashes")
	}

	// Must not have consecutive dashes
	if strings.Contains(dValue, "--") {
		return fmt.Errorf("d tag value cannot contain consecutive dashes")
	}

	return nil
}

// validateWikiTitleTag validates the title tag for wiki articles
func validateWikiTitleTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("title tag must have exactly 2 elements")
	}

	title := tag[1]
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Title should not be excessively long
	if len(title) > 200 {
		return fmt.Errorf("title too long (max 200 characters)")
	}

	// Title should not contain newlines
	if strings.Contains(title, "\n") || strings.Contains(title, "\r") {
		return fmt.Errorf("title cannot contain newlines")
	}

	return nil
}

// validateWikiSummaryTag validates the summary tag for wiki articles
func validateWikiSummaryTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("summary tag must have exactly 2 elements")
	}

	summary := tag[1]
	if summary == "" {
		return fmt.Errorf("summary cannot be empty")
	}

	// Summary should not be excessively long
	if len(summary) > 500 {
		return fmt.Errorf("summary too long (max 500 characters)")
	}

	return nil
}

// validateMergeRequestATag validates the 'a' tag in merge requests
func validateMergeRequestATag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("merge request 'a' tag must have 2-3 elements")
	}

	addressStr := tag[1]
	if addressStr == "" {
		return fmt.Errorf("merge request target address cannot be empty")
	}

	// Validate event address format: kind:pubkey:dtag
	parts := strings.Split(addressStr, ":")
	if len(parts) != 3 {
		return fmt.Errorf("target address must be in format 'kind:pubkey:dtag'")
	}

	// Validate kind (should be 30818 for wiki articles)
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in target address: %s", parts[0])
	}
	
	if kind != 30818 {
		return fmt.Errorf("merge request target must be a wiki article (kind 30818), got %d", kind)
	}

	// Validate pubkey
	if len(parts[1]) != 64 || !isHexChar64(parts[1]) {
		return fmt.Errorf("invalid pubkey in target address")
	}

	// Validate dtag (should follow normalization rules)
	normalizedDTag := normalizeDTag(parts[2])
	if parts[2] != normalizedDTag {
		return fmt.Errorf("dtag in target address is not properly normalized")
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid relay hint in merge request: %w", err)
		}
	}

	return nil
}

// validateMergeRequestETag validates the 'e' tag in merge requests
func validateMergeRequestETag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 4 {
		return fmt.Errorf("merge request 'e' tag must have 2-4 elements")
	}

	eventID := tag[1]
	if len(eventID) != 64 || !isHexChar64(eventID) {
		return fmt.Errorf("invalid event ID in merge request")
	}

	// Optional relay hint validation
	if len(tag) >= 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid relay hint in merge request e tag: %w", err)
		}
	}

	// Optional marker validation
	if len(tag) == 4 && tag[3] != "" {
		validMarkers := map[string]bool{
			"source": true,
			"fork":   true,
			"defer":  true,
		}
		
		if !validMarkers[tag[3]] {
			return fmt.Errorf("invalid e tag marker '%s', allowed: source, fork, defer", tag[3])
		}
	}

	return nil
}

// validateWikiRedirectTag validates the redirect tag for wiki redirects
func validateWikiRedirectTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("redirect tag must have exactly 2 elements")
	}

	redirectTarget := tag[1]
	if redirectTarget == "" {
		return fmt.Errorf("redirect target cannot be empty")
	}

	// Parse the redirect target (should be in format kind:pubkey:dtag)
	parts := strings.Split(redirectTarget, ":")
	if len(parts) != 3 {
		return fmt.Errorf("redirect target must be in format 'kind:pubkey:dtag'")
	}

	// Validate kind
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return fmt.Errorf("invalid kind in redirect target: %s", parts[0])
	}

	// Validate pubkey (should be 64 hex characters)
	if len(parts[1]) != 64 || !isHexChar64(parts[1]) {
		return fmt.Errorf("invalid pubkey in redirect target")
	}

	// Validate that the dtag part follows normalization rules
	dTag := parts[2]
	normalizedDTag := normalizeDTag(dTag)
	if dTag != normalizedDTag {
		return fmt.Errorf("redirect target dtag '%s' is not properly normalized (should be '%s')", dTag, normalizedDTag)
	}

	// Additional validation: dtag must contain at least one letter
	if !regexp.MustCompile(`[a-z]`).MatchString(dTag) {
		return fmt.Errorf("redirect target dtag must contain at least one letter")
	}

	return nil
}

// validateEventAddressTag validates event address tags (a tags)
func validateEventAddressTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 4 {
		return fmt.Errorf("event address tag must have 2-4 elements")
	}

	addressStr := tag[1]
	if addressStr == "" {
		return fmt.Errorf("event address cannot be empty")
	}

	// Validate event address format: kind:pubkey:dtag
	parts := strings.Split(addressStr, ":")
	if len(parts) != 3 {
		return fmt.Errorf("event address must be in format 'kind:pubkey:dtag'")
	}

	// Validate kind
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return fmt.Errorf("invalid kind in event address: %s", parts[0])
	}

	// Validate pubkey
	if len(parts[1]) != 64 || !isHexChar64(parts[1]) {
		return fmt.Errorf("invalid pubkey in event address")
	}

	// Optional relay hint validation
	if len(tag) >= 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid relay hint: %w", err)
		}
	}

	// Optional marker validation
	if len(tag) == 4 && tag[3] != "" {
		validMarkers := map[string]bool{
			"fork":  true,
			"defer": true,
		}
		
		if !validMarkers[tag[3]] {
			return fmt.Errorf("invalid a tag marker '%s', allowed: fork, defer", tag[3])
		}
	}

	return nil
}

// validateEventIDTag validates event ID tags (e tags)
func validateEventIDTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 4 {
		return fmt.Errorf("event ID tag must have 2-4 elements")
	}

	eventID := tag[1]
	if len(eventID) != 64 || !isHexChar64(eventID) {
		return fmt.Errorf("invalid event ID")
	}

	// Optional relay hint validation
	if len(tag) >= 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid relay hint: %w", err)
		}
	}

	// Optional marker validation
	if len(tag) == 4 && tag[3] != "" {
		validMarkers := map[string]bool{
			"fork":  true,
			"defer": true,
		}
		
		if !validMarkers[tag[3]] {
			return fmt.Errorf("invalid e tag marker '%s', allowed: fork, defer", tag[3])
		}
	}

	return nil
}

// validatePubkeyTag validates pubkey tags (p tags)
func validatePubkeyTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("pubkey tag must have 2-3 elements")
	}

	pubkey := tag[1]
	if len(pubkey) != 64 || !isHexChar64(pubkey) {
		return fmt.Errorf("invalid pubkey")
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid relay hint: %w", err)
		}
	}

	return nil
}

// validateWikiContent validates the content of wiki articles
func validateWikiContent(content string) error {
	// Content should not be empty for wiki articles
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("wiki content cannot be empty")
	}

	// Validate wikilink syntax
	if err := validateWikilinks(content); err != nil {
		return fmt.Errorf("invalid wikilinks: %w", err)
	}

	// Validate nostr links
	if err := validateNostrLinks(content); err != nil {
		return fmt.Errorf("invalid nostr links: %w", err)
	}

	// Basic Asciidoc validation - check for common syntax errors
	if err := validateAsciidocSyntax(content); err != nil {
		return fmt.Errorf("invalid Asciidoc syntax: %w", err)
	}

	return nil
}

// validateWikilinks validates wikilink syntax in content
func validateWikilinks(content string) error {
	// Find all wikilinks: [[...]] patterns
	wikilinkPattern := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	matches := wikilinkPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		linkContent := match[1]
		
		// Check for pipe syntax: [[target|display]]
		if strings.Contains(linkContent, "|") {
			parts := strings.SplitN(linkContent, "|", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid wikilink syntax: %s", match[0])
			}
			
			target := strings.TrimSpace(parts[0])
			display := strings.TrimSpace(parts[1])
			
			if target == "" {
				return fmt.Errorf("wikilink target cannot be empty: %s", match[0])
			}
			
			if display == "" {
				return fmt.Errorf("wikilink display text cannot be empty: %s", match[0])
			}
		} else {
			// Simple wikilink: [[Target Page]]
			target := strings.TrimSpace(linkContent)
			if target == "" {
				return fmt.Errorf("wikilink target cannot be empty: %s", match[0])
			}
		}
	}

	return nil
}

// validateNostrLinks validates nostr: links in content
func validateNostrLinks(content string) error {
	// Find all nostr: links
	nostrLinkPattern := regexp.MustCompile(`nostr:([a-zA-Z0-9]+)`)
	matches := nostrLinkPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		identifier := match[1]
		
		// Basic validation - should be bech32-like format
		if len(identifier) < 10 {
			return fmt.Errorf("invalid nostr identifier too short: %s", match[0])
		}

		// Check for valid bech32 characters
		if !regexp.MustCompile(`^[a-z0-9]+$`).MatchString(identifier) {
			return fmt.Errorf("invalid nostr identifier contains invalid characters: %s", match[0])
		}
	}

	return nil
}

// validateAsciidocSyntax performs basic Asciidoc syntax validation
func validateAsciidocSyntax(content string) error {
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		// Check for unbalanced brackets in links
		if strings.Count(line, "[") != strings.Count(line, "]") {
			// Allow wikilinks which use double brackets
			wikilinkCount := strings.Count(line, "[[") * 2
			if strings.Count(line, "[") - wikilinkCount != strings.Count(line, "]") - wikilinkCount {
				return fmt.Errorf("unbalanced brackets on line %d: %s", i+1, line)
			}
		}

		// Check for common Asciidoc syntax issues
		if strings.Contains(line, "<<") && !strings.Contains(line, ">>") {
			return fmt.Errorf("unclosed cross-reference on line %d: %s", i+1, line)
		}
	}

	return nil
}

// normalizeDTag normalizes a d tag value according to NIP-54 rules
func normalizeDTag(value string) string {
	// Convert to lowercase
	normalized := strings.ToLower(value)
	
	// Replace any non-letter character with dash
	result := make([]rune, 0, len(normalized))
	
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') {
			result = append(result, r)
		} else {
			result = append(result, '-')
		}
	}
	
	// Convert consecutive dashes to single dash
	normalizedStr := regexp.MustCompile(`-+`).ReplaceAllString(string(result), "-")
	
	// Remove leading and trailing dashes
	normalizedStr = strings.Trim(normalizedStr, "-")
	
	return normalizedStr
}

// validateBasicEvent performs basic validation common to all NIP-54 events
func validateBasicEvent(event *nostr.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if event.PubKey == "" {
		return fmt.Errorf("event must have a pubkey")
	}

	if len(event.PubKey) != 64 || !isHexChar64(event.PubKey) {
		return fmt.Errorf("invalid pubkey format")
	}

	if event.Tags == nil {
		return fmt.Errorf("event must have tags")
	}

	return nil
}