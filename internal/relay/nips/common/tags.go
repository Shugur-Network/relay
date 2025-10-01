package common

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	nostr "github.com/nbd-wtf/go-nostr"
)

// TagValidator provides common tag validation utilities
type TagValidator struct{}

// NewTagValidator creates a new tag validator instance
func NewTagValidator() *TagValidator {
	return &TagValidator{}
}

// ValidateRequiredTag checks if an event has a required tag
func (tv *TagValidator) ValidateRequiredTag(event *nostr.Event, tagName string) error {
	if !tv.HasTag(event, tagName) {
		return fmt.Errorf("missing required '%s' tag", tagName)
	}
	return nil
}

// ValidateRequiredTags checks if an event has all required tags
func (tv *TagValidator) ValidateRequiredTags(event *nostr.Event, tagNames ...string) error {
	for _, tagName := range tagNames {
		if err := tv.ValidateRequiredTag(event, tagName); err != nil {
			return err
		}
	}
	return nil
}

// HasTag checks if an event contains a specific tag
func (tv *TagValidator) HasTag(event *nostr.Event, tagName string) bool {
	for _, tag := range event.Tags {
		if len(tag) > 0 && tag[0] == tagName {
			return true
		}
	}
	return false
}

// GetTagValue returns the first value (tag[1]) for a given tag key, or empty string if not found
func (tv *TagValidator) GetTagValue(event *nostr.Event, tagName string) string {
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == tagName {
			return tag[1]
		}
	}
	return ""
}

// GetAllTagValues returns all values for tags with the given name
func (tv *TagValidator) GetAllTagValues(event *nostr.Event, tagName string) []string {
	var values []string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == tagName {
			values = append(values, tag[1])
		}
	}
	return values
}

// GetTagsByName returns all tags with the given name
func (tv *TagValidator) GetTagsByName(event *nostr.Event, tagName string) [][]string {
	var tags [][]string
	for _, tag := range event.Tags {
		if len(tag) > 0 && tag[0] == tagName {
			tags = append(tags, tag)
		}
	}
	return tags
}

// ValidateTagValue validates that a tag exists and its value meets criteria
func (tv *TagValidator) ValidateTagValue(event *nostr.Event, tagName string, validator func(string) error) error {
	value := tv.GetTagValue(event, tagName)
	if value == "" {
		return fmt.Errorf("missing or empty '%s' tag", tagName)
	}
	return validator(value)
}

// ValidateOptionalTagValue validates a tag value only if the tag exists
func (tv *TagValidator) ValidateOptionalTagValue(event *nostr.Event, tagName string, validator func(string) error) error {
	value := tv.GetTagValue(event, tagName)
	if value == "" {
		return nil // Optional tag, no validation needed
	}
	return validator(value)
}

// Common validation functions

// ValidateHexString validates that a string is valid hexadecimal
func (tv *TagValidator) ValidateHexString(value string, expectedLength int) error {
	if len(value) != expectedLength {
		return fmt.Errorf("invalid hex string length: expected %d, got %d", expectedLength, len(value))
	}

	hexRegex := regexp.MustCompile("^[0-9a-fA-F]+$")
	if !hexRegex.MatchString(value) {
		return fmt.Errorf("invalid hex string format")
	}

	return nil
}

// ValidatePubkey validates a nostr public key (64-character hex)
func (tv *TagValidator) ValidatePubkey(pubkey string) error {
	if !nostr.IsValid32ByteHex(pubkey) {
		return fmt.Errorf("invalid pubkey format")
	}
	return nil
}

// ValidateEventID validates a nostr event ID (64-character hex)
func (tv *TagValidator) ValidateEventID(eventID string) error {
	return tv.ValidateHexString(eventID, 64)
}

// ValidateURL validates that a string is a valid URL
func (tv *TagValidator) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("URL must have scheme and host")
	}

	return nil
}

// ValidatePositiveInteger validates that a string represents a positive integer
func (tv *TagValidator) ValidatePositiveInteger(value string) error {
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid integer format: %w", err)
	}

	if num <= 0 {
		return fmt.Errorf("value must be positive, got: %d", num)
	}

	return nil
}

// ValidateNonEmptyString validates that a string is not empty or whitespace
func (tv *TagValidator) ValidateNonEmptyString(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty or whitespace")
	}
	return nil
}

// ValidateStringLength validates that a string is within specified length bounds
func (tv *TagValidator) ValidateStringLength(value string, minLength, maxLength int) error {
	length := len(value)
	if length < minLength {
		return fmt.Errorf("string too short: minimum %d characters, got %d", minLength, length)
	}
	if maxLength > 0 && length > maxLength {
		return fmt.Errorf("string too long: maximum %d characters, got %d", maxLength, length)
	}
	return nil
}

// ValidateTagCount validates the number of occurrences of a specific tag
func (tv *TagValidator) ValidateTagCount(event *nostr.Event, tagName string, minCount, maxCount int) error {
	tags := tv.GetTagsByName(event, tagName)
	count := len(tags)

	if count < minCount {
		return fmt.Errorf("insufficient '%s' tags: minimum %d required, got %d", tagName, minCount, count)
	}

	if maxCount > 0 && count > maxCount {
		return fmt.Errorf("too many '%s' tags: maximum %d allowed, got %d", tagName, maxCount, count)
	}

	return nil
}

// ValidateUniqueTagValues validates that all values for a tag are unique
func (tv *TagValidator) ValidateUniqueTagValues(event *nostr.Event, tagName string) error {
	values := tv.GetAllTagValues(event, tagName)
	seen := make(map[string]bool)

	for _, value := range values {
		if seen[value] {
			return fmt.Errorf("duplicate '%s' tag value: %s", tagName, value)
		}
		seen[value] = true
	}

	return nil
}
