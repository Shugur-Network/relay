package nips

import (
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// ValidateReport validates NIP-56 report events (kind 1984)
func ValidateReport(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	if event.Kind != 1984 {
		return fmt.Errorf("invalid kind for report: expected 1984, got %d", event.Kind)
	}

	logger.Debug("NIP-56: Validating report event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	// Validate required tags
	if err := validateReportTags(event); err != nil {
		return fmt.Errorf("invalid report tags: %w", err)
	}

	// Validate content length (optional but reasonable)
	if len(event.Content) > 2000 {
		return fmt.Errorf("report content too long: maximum 2000 characters")
	}

	logger.Debug("NIP-56: Report validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// validateReportTags validates the tag structure for NIP-56 reports
func validateReportTags(event *nostr.Event) error {
	var hasPTag bool
	var xTags []nostr.Tag
	var eTags []nostr.Tag
	var lTags []nostr.Tag
	var lNamespaces []string

	// Process all tags
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue // Skip malformed tags
		}

		switch tag[0] {
		case "p":
			if err := validatePTag(tag); err != nil {
				return fmt.Errorf("invalid p tag: %w", err)
			}
			hasPTag = true

		case "e":
			if err := validateETag(tag); err != nil {
				return fmt.Errorf("invalid e tag: %w", err)
			}
			eTags = append(eTags, tag)

		case "x":
			if err := validateXTag(tag); err != nil {
				return fmt.Errorf("invalid x tag: %w", err)
			}
			xTags = append(xTags, tag)

		case "server":
			if err := validateServerTag(tag); err != nil {
				return fmt.Errorf("invalid server tag: %w", err)
			}

		case "L":
			if err := validateLabelNamespaceTag(tag); err != nil {
				return fmt.Errorf("invalid L tag: %w", err)
			}
			lNamespaces = append(lNamespaces, tag[1])

		case "l":
			if err := validateLabelTag(tag); err != nil {
				return fmt.Errorf("invalid l tag: %w", err)
			}
			lTags = append(lTags, tag)
		}
	}

	// Required: Must have at least one p tag
	if !hasPTag {
		return fmt.Errorf("report must include at least one 'p' tag referencing the pubkey being reported")
	}

	// Validate x tag constraints
	if err := validateXTagConstraints(xTags, eTags); err != nil {
		return fmt.Errorf("x tag constraint violation: %w", err)
	}

	// Validate NIP-32 label relationships (l tags require L tags)
	if err := validateNIP32LabelRelationships(lTags, lNamespaces); err != nil {
		return fmt.Errorf("NIP-32 label validation failed: %w", err)
	}

	return nil
}

// validatePTag validates a p (pubkey) tag with report type
func validatePTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("p tag must have at least 2 elements")
	}

	// Validate pubkey (element 1)
	pubkey := tag[1]
	if !isValidHex(pubkey, 64) {
		return fmt.Errorf("invalid pubkey hex: %s", pubkey)
	}

	// Validate report type (element 2) - required for NIP-56
	if len(tag) >= 3 {
		reportType := tag[2]
		if !isValidReportType(reportType) {
			return fmt.Errorf("invalid report type: %s", reportType)
		}

		// Context validation: impersonation only makes sense for profiles
		// (impersonation reports are for profiles using p tags)
		_ = reportType // This validation is complete
	}

	return nil
}

// validateETag validates an e (event) tag with report type
func validateETag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("e tag must have at least 2 elements")
	}

	// Validate event ID (element 1)
	eventID := tag[1]
	if !isValidHex(eventID, 64) {
		return fmt.Errorf("invalid event ID hex: %s", eventID)
	}

	// Validate report type (element 2) if present
	if len(tag) >= 3 {
		reportType := tag[2]
		if reportType != "" && !isValidReportType(reportType) {
			return fmt.Errorf("invalid report type: %s", reportType)
		}

		// Context validation: some report types don't make sense for events
		if reportType == "impersonation" {
			return fmt.Errorf("impersonation report type should be used with p tags, not e tags")
		}
	}

	return nil
}

// validateXTag validates an x (blob hash) tag with report type
func validateXTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("x tag must have at least 2 elements")
	}

	// Validate blob hash (element 1) - can be various hash formats
	blobHash := tag[1]
	if len(blobHash) == 0 {
		return fmt.Errorf("blob hash cannot be empty")
	}

	// Basic validation: should be reasonable hash length (SHA256, etc.)
	if len(blobHash) < 32 || len(blobHash) > 128 {
		return fmt.Errorf("blob hash length invalid: expected 32-128 characters, got %d", len(blobHash))
	}

	// Validate report type (element 2) if present
	if len(tag) >= 3 {
		reportType := tag[2]
		if reportType != "" && !isValidReportType(reportType) {
			return fmt.Errorf("invalid report type: %s", reportType)
		}

		// Context validation
		if reportType == "impersonation" {
			return fmt.Errorf("impersonation report type should be used with p tags, not x tags")
		}
	}

	return nil
}

// validateServerTag validates a server tag (for blob location)
func validateServerTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("server tag must have at least 2 elements")
	}

	serverURL := tag[1]
	if !isValidURLForReporting(serverURL) {
		return fmt.Errorf("invalid server URL: %s", serverURL)
	}

	return nil
}

// validateLabelNamespaceTag validates an L tag (label namespace from NIP-32)
func validateLabelNamespaceTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("l tag must have at least 2 elements")
	}

	namespace := tag[1]
	if len(namespace) == 0 {
		return fmt.Errorf("label namespace cannot be empty")
	}

	// Basic namespace validation
	if len(namespace) > 100 {
		return fmt.Errorf("label namespace too long: maximum 100 characters")
	}

	return nil
}

// validateLabelTag validates an l tag (label from NIP-32)
func validateLabelTag(tag nostr.Tag) error {
	if len(tag) < 2 {
		return fmt.Errorf("l tag must have at least 2 elements")
	}

	label := tag[1]
	if len(label) == 0 {
		return fmt.Errorf("label cannot be empty")
	}

	// Basic label validation
	if len(label) > 100 {
		return fmt.Errorf("label too long: maximum 100 characters")
	}

	// Validate namespace (element 2) if present
	if len(tag) >= 3 {
		namespace := tag[2]
		if len(namespace) > 100 {
			return fmt.Errorf("label namespace too long: maximum 100 characters")
		}
	}

	return nil
}

// validateXTagConstraints validates constraints specific to x tags
func validateXTagConstraints(xTags []nostr.Tag, eTags []nostr.Tag) error {
	// NIP-56: when x tag is present, client MUST include an e tag
	if len(xTags) > 0 && len(eTags) == 0 {
		return fmt.Errorf("when x tag is present, an e tag must also be included")
	}

	return nil
}

// isValidReportType checks if the report type is one of the allowed values
func isValidReportType(reportType string) bool {
	validTypes := map[string]bool{
		"nudity":        true,
		"malware":       true,
		"profanity":     true,
		"illegal":       true,
		"spam":          true,
		"impersonation": true,
		"other":         true,
	}

	return validTypes[reportType]
}

// isValidHex checks if a string is valid hexadecimal of specified length
func isValidHex(s string, expectedLength int) bool {
	if len(s) != expectedLength {
		return false
	}

	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

// isValidURLForReporting performs basic URL validation for reporting
func isValidURLForReporting(urlStr string) bool {
	if len(urlStr) == 0 {
		return false
	}

	// Must start with http:// or https://
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}

	// Basic length check
	if len(urlStr) > 2000 {
		return false
	}

	// Should contain at least a domain part
	if len(urlStr) < 10 { // "http://a.b" is minimum
		return false
	}

	return true
}

// GetValidReportTypes returns the list of valid report types for reference
func GetValidReportTypes() []string {
	return []string{
		"nudity",
		"malware", 
		"profanity",
		"illegal",
		"spam",
		"impersonation",
		"other",
	}
}

// GetReportTypeDescription returns a human-readable description of a report type
func GetReportTypeDescription(reportType string) string {
	descriptions := map[string]string{
		"nudity":        "Depictions of nudity, porn, etc.",
		"malware":       "Virus, trojan horse, worm, robot, spyware, adware, back door, ransomware, rootkit, kidnapper, etc.",
		"profanity":     "Profanity, hateful speech, etc.",
		"illegal":       "Something which may be illegal in some jurisdiction",
		"spam":          "Spam",
		"impersonation": "Someone pretending to be someone else",
		"other":         "For reports that don't fit in the above categories",
	}

	if desc, exists := descriptions[reportType]; exists {
		return desc
	}
	return "Unknown report type"
}

// validateNIP32LabelRelationships validates that l tags have corresponding L tags
func validateNIP32LabelRelationships(lTags []nostr.Tag, lNamespaces []string) error {
	// If no l tags, no validation needed
	if len(lTags) == 0 {
		return nil
	}

	// If there are l tags but no L tags, that's invalid
	if len(lNamespaces) == 0 {
		return fmt.Errorf("l tags require corresponding L tags to define namespaces")
	}

	// Check each l tag has a valid namespace
	for _, lTag := range lTags {
		if len(lTag) < 3 {
			continue // Already validated in validateLabelTag
		}

		namespace := lTag[2] // Third element is the namespace
		hasMatchingL := false

		for _, lNamespace := range lNamespaces {
			if lNamespace == namespace {
				hasMatchingL = true
				break
			}
		}

		if !hasMatchingL {
			return fmt.Errorf("l tag references namespace '%s' but no matching L tag found", namespace)
		}
	}

	return nil
}