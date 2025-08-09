package domain

import (
	"github.com/nbd-wtf/go-nostr"
)

// WebSocketConnection represents a client WebSocket connection.
// This abstraction is used by both the relay and application packages.
type WebSocketConnection interface {
	// Core connection methods
	SendMessage(msg []byte)
	SendEvent(subID string, evt *nostr.Event)
	Close()

	// Subscription management
	GetSubscriptions() map[string][]nostr.Filter
	AddSubscription(subID string, filters []nostr.Filter)
	RemoveSubscription(subID string)
	HasSubscription(subID string) bool

	// Remote address for logging/identification
	RemoteAddr() string
}

// ConnectionManager defines the interface for managing WebSocket connections
type ConnectionManager interface {
	RegisterConn(conn WebSocketConnection)
	UnregisterConn(conn WebSocketConnection)
}
