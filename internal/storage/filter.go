package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// CompiledFilter represents a pre-compiled filter for efficient matching
type CompiledFilter struct {
	IDs     map[string]bool
	Authors map[string]bool
	Kinds   map[int]bool
	Since   *time.Time
	Until   *time.Time
	Tags    map[string]map[string]bool
	Limit   int
	Search  string
}

// CompileFilter pre-compiles a nostr filter for efficient matching
func CompileFilter(f nostr.Filter) *CompiledFilter {
	cf := &CompiledFilter{
		IDs:     make(map[string]bool),
		Authors: make(map[string]bool),
		Kinds:   make(map[int]bool),
		Tags:    make(map[string]map[string]bool),
		Limit:   f.Limit,
		Search:  f.Search,
	}

	// Set default limit of 500 if no limit specified
	if cf.Limit <= 0 {
		cf.Limit = 500
	}

	// Pre-compile IDs
	for _, id := range f.IDs {
		cf.IDs[id] = true
	}

	// Pre-compile Authors
	for _, author := range f.Authors {
		cf.Authors[author] = true
	}

	// Pre-compile Kinds
	for _, kind := range f.Kinds {
		cf.Kinds[kind] = true
	}

	// Set time bounds
	if f.Since != nil {
		t := f.Since.Time()
		cf.Since = &t
	}
	if f.Until != nil {
		t := f.Until.Time()
		cf.Until = &t
	}

	// Pre-compile Tags
	for tagName, tagValues := range f.Tags {
		cf.Tags[tagName] = make(map[string]bool)
		for _, value := range tagValues {
			cf.Tags[tagName][value] = true
		}
	}

	return cf
}

// GetBestIndex determines the most efficient index to use for the filter
func (cf *CompiledFilter) GetBestIndex() string {
	// If we have IDs, use the primary key index
	if len(cf.IDs) > 0 {
		return "id"
	}

	// If we have both authors and kinds, use the composite index
	if len(cf.Authors) > 0 && len(cf.Kinds) > 0 {
		return "pubkey_kind_created"
	}

	// If we only have kinds, use the kind index
	if len(cf.Kinds) > 0 {
		return "kind_created"
	}

	// Default to created_at index
	return "created_at"
}

// BuildQuery constructs the SQL query using the most efficient index
func (cf *CompiledFilter) BuildQuery() (string, []interface{}, error) {
	query := strings.Builder{}
	args := make([]interface{}, 0, 10)
	argIndex := 1

	// Start with base SELECT
	query.WriteString(`SELECT id, pubkey, kind, created_at, content, tags, sig FROM events`)

	// Add WHERE clause based on best index
	switch cf.GetBestIndex() {
	case "id":
		// Use primary key index
		placeholders := make([]string, len(cf.IDs))
		i := 0
		for id := range cf.IDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, id)
			argIndex++
			i++
		}
		query.WriteString(fmt.Sprintf(" WHERE id = ANY(ARRAY[%s]::text[])", strings.Join(placeholders, ",")))

	case "pubkey_kind_created":
		// Use composite index for authors and kinds
		authorPlaceholders := make([]string, len(cf.Authors))
		i := 0
		for author := range cf.Authors {
			authorPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, author)
			argIndex++
			i++
		}
		kindPlaceholders := make([]string, len(cf.Kinds))
		i = 0
		for kind := range cf.Kinds {
			kindPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, kind)
			argIndex++
			i++
		}
		query.WriteString(fmt.Sprintf(" WHERE pubkey = ANY(ARRAY[%s]::text[]) AND kind = ANY(ARRAY[%s]::integer[])",
			strings.Join(authorPlaceholders, ","), strings.Join(kindPlaceholders, ",")))

	case "kind_created":
		// Use kind index
		kindPlaceholders := make([]string, len(cf.Kinds))
		i := 0
		for kind := range cf.Kinds {
			kindPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, kind)
			argIndex++
			i++
		}
		query.WriteString(fmt.Sprintf(" WHERE kind = ANY(ARRAY[%s]::integer[])", strings.Join(kindPlaceholders, ",")))

	default:
		// Use created_at index
		query.WriteString(" WHERE true")
	}

	// Add time filters
	if cf.Since != nil {
		query.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIndex))
		args = append(args, cf.Since.Unix())
		argIndex++
	}
	if cf.Until != nil {
		query.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIndex))
		args = append(args, cf.Until.Unix())
		argIndex++
	}

	// Add search filter if present
	if cf.Search != "" {
		query.WriteString(fmt.Sprintf(" AND content ILIKE $%d", argIndex))
		args = append(args, "%"+cf.Search+"%")
		argIndex++
	}

	// Add tag filters
	for tagName, tagValues := range cf.Tags {
		if len(tagValues) > 0 {
			query.WriteString(fmt.Sprintf(" AND tags @> $%d", argIndex))
			tagArray := make([][]string, len(tagValues))
			i := 0
			for value := range tagValues {
				tagArray[i] = []string{tagName, value}
				i++
			}
			args = append(args, tagArray)
			argIndex++
		}
	}

	// // Add ordering and limit - use DESC order to get newest events first
	// query.WriteString(" ORDER BY created_at DESC LIMIT $")
	// Add ordering and limit
	// Use ASC order for since-only filters to get oldest events since the timestamp
	// Use DESC order for all other cases to get newest events first
	if cf.Since != nil && cf.Until == nil {
		query.WriteString(" ORDER BY created_at ASC LIMIT $")
	} else {
		query.WriteString(" ORDER BY created_at DESC LIMIT $")
	}
	query.WriteString(fmt.Sprintf("%d", argIndex))
	args = append(args, cf.Limit)

	return query.String(), args, nil
}
