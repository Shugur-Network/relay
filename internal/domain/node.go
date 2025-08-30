package domain

import (
	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/storage"
	nostr "github.com/nbd-wtf/go-nostr"
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

	// Event dispatcher access
	GetEventDispatcher() *storage.EventDispatcher
}

// EventDispatcherClient represents a client that receives real-time event notifications
type EventDispatcherClient interface {
	// AddEventDispatcherClient registers the client for real-time event notifications
	AddEventDispatcherClient(clientID string) chan *nostr.Event

	// RemoveEventDispatcherClient unregisters the client from real-time event notifications
	RemoveEventDispatcherClient(clientID string)
}
