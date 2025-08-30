package nips

import (
	"context"
	"fmt"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// NIP-45: COUNT Command
// https://github.com/nostr-protocol/nips/blob/master/45.md

const (
	// CountTimeout is the maximum time allowed for a COUNT operation
	CountTimeout = 5 * time.Second
)

// CountCommand represents a parsed COUNT command
type CountCommand struct {
	SubID  string
	Filter nostr.Filter
}

// CountResponse represents the response to a COUNT command
type CountResponse struct {
	Count int64 `json:"count"`
}

// ParseCountCommand parses a COUNT command from raw message array
// Note: This function expects the filter to be parsed externally using parseFilterFromRaw
func ParseCountCommand(arr []interface{}) (*CountCommand, error) {
	// Validate array length
	if len(arr) < 3 {
		return nil, fmt.Errorf("COUNT command missing subscription ID or filter")
	}

	// Extract subscription ID
	subID, ok := arr[1].(string)
	if !ok || subID == "" {
		return nil, fmt.Errorf("COUNT command subscription ID must be a string")
	}

	// The filter parsing will be handled externally
	// This is just a placeholder structure
	return &CountCommand{
		SubID: subID,
		// Filter will be set externally after parsing
	}, nil
}

// ValidateCountFilter validates a filter for COUNT operations
func ValidateCountFilter(filter nostr.Filter) error {
	// Apply reasonable limits for COUNT operations
	if len(filter.Authors) > 100 {
		return fmt.Errorf("too many authors in filter (max 100)")
	}

	if len(filter.Kinds) > 20 {
		return fmt.Errorf("too many kinds in filter (max 20)")
	}

	// Validate kinds
	for _, kind := range filter.Kinds {
		if kind < 0 || kind > 65535 {
			return fmt.Errorf("invalid event kind: %d", kind)
		}
	}

	// Validate authors (should be valid hex)
	for _, author := range filter.Authors {
		if !nostr.IsValid32ByteHex(author) {
			return fmt.Errorf("invalid author pubkey: %s", author)
		}
	}

	return nil
}

// HandleCountRequest validates and processes a COUNT request
func HandleCountRequest(ctx context.Context, subID string, filter nostr.Filter) (*CountResponse, error) {
	// Validate the filter
	if err := ValidateCountFilter(filter); err != nil {
		logger.Warn("COUNT filter validation failed",
			zap.String("sub_id", subID),
			zap.Error(err))
		return nil, err
	}

	logger.Debug("COUNT request validated successfully",
		zap.String("sub_id", subID),
		zap.Any("filter", filter))

	// Return success - actual count will be handled by the database layer
	return &CountResponse{Count: 0}, nil
}

// FormatCountResponse formats a count response for sending to client
func FormatCountResponse(subID string, count int64) []interface{} {
	return []interface{}{
		"COUNT",
		subID,
		map[string]int64{"count": count},
	}
}
