package domain

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

// EventHandler defines the operations needed to process Nostr events.
type EventHandler interface {
	// Handle an incoming event from a client
	BroadcastEvent(ctx context.Context, evt *nostr.Event) error

	// Query stored events based on a filter
	QueryEvents(filter nostr.Filter) ([]nostr.Event, error)

	// Get the event count for a filter
	GetEventCount(ctx context.Context, filter nostr.Filter) (int64, error)
}

// ValidationResult represents the outcome of event validation
type ValidationResult struct {
	Valid  bool
	Reason string
	Error  error
}

// EventValidator defines the interface for validating Nostr events
type EventValidator interface {
	// Validate a Nostr event
	ValidateEvent(ctx context.Context, event nostr.Event) (bool, string)

	// Validate a Nostr filter
	ValidateFilter(filter nostr.Filter) error

	// Validate and process an event
	ValidateAndProcessEvent(ctx context.Context, event nostr.Event) (bool, string, error)
}
