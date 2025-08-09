package nips

import (
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// SearchOptions represents the search configuration for NIP-50
type SearchOptions struct {
	CaseSensitive bool // Whether search should be case sensitive
	MaxTerms      int  // Maximum number of search terms
	MinTermLength int  // Minimum length of each search term
	MaxTermLength int  // Maximum length of each search term
}

// DefaultSearchOptions returns the default search options
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		CaseSensitive: false,
		MaxTerms:      10,
		MinTermLength: 2,
		MaxTermLength: 100,
	}
}

// ValidateSearchFilter validates a filter's search parameters according to NIP-50
func ValidateSearchFilter(filter nostr.Filter, opts SearchOptions) error {
	if filter.Search == "" {
		return nil
	}

	// Split search terms
	terms := strings.Fields(filter.Search)
	if len(terms) > opts.MaxTerms {
		return fmt.Errorf("too many search terms (max %d)", opts.MaxTerms)
	}

	// Validate each term
	for _, term := range terms {
		if len(term) < opts.MinTermLength {
			return fmt.Errorf("search term too short (min %d chars): %s", opts.MinTermLength, term)
		}
		if len(term) > opts.MaxTermLength {
			return fmt.Errorf("search term too long (max %d chars): %s", opts.MaxTermLength, term)
		}
	}

	return nil
}

// BuildSearchQuery builds a CockroachDB-compatible search query
func BuildSearchQuery(search string, opts SearchOptions) (string, []string, error) {
	terms := strings.Fields(search)
	if len(terms) == 0 {
		return "", nil, fmt.Errorf("empty search string")
	}

	// Validate terms
	if err := ValidateSearchFilter(nostr.Filter{Search: search}, opts); err != nil {
		return "", nil, err
	}

	// Build query parts and collect terms
	queryParts := make([]string, len(terms))
	escapedTerms := make([]string, len(terms))

	for i, term := range terms {
		// Escape special characters for LIKE
		escapedTerm := strings.ReplaceAll(term, "%", "\\%")
		escapedTerm = strings.ReplaceAll(escapedTerm, "_", "\\_")
		escapedTerms[i] = escapedTerm
		queryParts[i] = fmt.Sprintf("content ILIKE $%d", i+1)
	}

	return "(" + strings.Join(queryParts, " AND ") + ")", escapedTerms, nil
}

// IsSearchableKind checks if an event kind should be included in search results
func IsSearchableKind(kind int) bool {
	// By default, only text notes (kind 1) are searchable
	// This can be extended based on relay configuration
	switch kind {
	case 1: // Text notes
		return true
	default:
		return false
	}
}
