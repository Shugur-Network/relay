package domain

import (
	"github.com/Shugur-Network/Relay/internal/config"
	"github.com/Shugur-Network/Relay/internal/storage"
)

// NodeInterface defines the core capabilities required by the relay.
type NodeInterface interface {
	// Database access
	DB() *storage.DB

	// Configuration access
	Config() *config.Config

	// Event processing
	// BroadcastEvent(ctx context.Context, evt *nostr.Event) error
	// QueryEvents(filter nostr.Filter) ([]nostr.Event, error)

	// Connection management
	RegisterConn(conn WebSocketConnection)
	UnregisterConn(conn WebSocketConnection)
	GetActiveConnectionCount() int64

	// Validation
	GetValidator() EventValidator

	// Event processor access
	GetEventProcessor() *storage.EventProcessor
}
