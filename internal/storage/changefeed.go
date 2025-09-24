package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// ChangefeedEvent represents a changefeed event from CockroachDB
type ChangefeedEvent struct {
	Table string        `json:"table"`
	Key   []interface{} `json:"key"`
	Value *EventRowData `json:"value"`
	After *EventRowData `json:"after"`
}

// EventRowData represents the event data structure from the database
type EventRowData struct {
	ID        string          `json:"id"`
	PubKey    string          `json:"pubkey"`
	CreatedAt int64           `json:"created_at"`
	Kind      int             `json:"kind"`
	Tags      json.RawMessage `json:"tags"`
	Content   string          `json:"content"`
	Sig       string          `json:"sig"`
}

// ToNostrEvent converts EventRowData to a nostr.Event
func (e *EventRowData) ToNostrEvent() (*nostr.Event, error) {
	evt := &nostr.Event{
		ID:        e.ID,
		PubKey:    e.PubKey,
		CreatedAt: nostr.Timestamp(e.CreatedAt),
		Kind:      e.Kind,
		Content:   e.Content,
		Sig:       e.Sig,
	}

	// Parse tags if they exist
	if len(e.Tags) > 0 {
		if err := json.Unmarshal(e.Tags, &evt.Tags); err != nil {
			logger.Warn("Failed to unmarshal tags in changefeed event", zap.Error(err))
			evt.Tags = []nostr.Tag{}
		}
	}

	return evt, nil
}

// EventDispatcher manages real-time event distribution across relay instances
type EventDispatcher struct {
	db              *DB
	clients         map[string]chan *nostr.Event
	clientsMu       sync.RWMutex
	eventBuffer     chan *nostr.Event
	ctx             context.Context
	cancel          context.CancelFunc
	changefeedQuery string
}

// NewEventDispatcher creates a new event dispatcher for real-time events
func NewEventDispatcher(db *DB) *EventDispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventDispatcher{
		db:              db,
		clients:         make(map[string]chan *nostr.Event),
		eventBuffer:     make(chan *nostr.Event, 1000),
		ctx:             ctx,
		cancel:          cancel,
		changefeedQuery: "", // Using polling instead of sinkless changefeed
	}
}

// Start begins listening to the changefeed and processing events
func (ed *EventDispatcher) Start() error {
	if !ed.db.isConnected() {
		return logger.NewError("database is not connected")
	}

	logger.Info("Starting event dispatcher...")

	// Check if we're running in cluster mode
	isCluster, err := ed.db.isClusterMode(ed.ctx)
	if err != nil {
		logger.Warn("Failed to detect cluster mode, defaulting to standalone", zap.Error(err))
		isCluster = false
	}

	if !isCluster {
		logger.Info("Standalone mode detected - skipping cross-node synchronization")
		// Only start local event processing
		go ed.processEvents()
		logger.Info("✅ Event dispatcher started in standalone mode")
		return nil
	}

	logger.Info("Cluster mode detected - enabling cross-node synchronization")

	// Verify changefeed capability before starting
	if err := ed.verifyChangefeedSupport(); err != nil {
		logger.Warn("Changefeed not supported, running without cross-node sync", zap.Error(err))
		// Still start local processing
		go ed.processEvents()
		logger.Info("✅ Event dispatcher started without cross-node sync")
		return nil
	}

	go ed.processEvents()
	go ed.listenToChangefeed()

	logger.Info("✅ Event dispatcher started with cross-node synchronization")
	return nil
}

// Stop stops the event dispatcher
func (ed *EventDispatcher) Stop() {
	logger.Info("Stopping event dispatcher...")
	ed.cancel()

	// Close all client channels
	ed.clientsMu.Lock()
	for clientID, clientChan := range ed.clients {
		close(clientChan)
		delete(ed.clients, clientID)
	}
	ed.clientsMu.Unlock()

	close(ed.eventBuffer)
	logger.Info("✅ Event dispatcher stopped")
}

// AddClient registers a new client for event notifications
func (ed *EventDispatcher) AddClient(clientID string) chan *nostr.Event {
	ed.clientsMu.Lock()
	defer ed.clientsMu.Unlock()

	clientChan := make(chan *nostr.Event, 100)
	ed.clients[clientID] = clientChan

	logger.Debug("Added event dispatcher client", zap.String("client_id", clientID))
	return clientChan
}

// RemoveClient unregisters a client from event notifications
func (ed *EventDispatcher) RemoveClient(clientID string) {
	ed.clientsMu.Lock()
	defer ed.clientsMu.Unlock()

	if clientChan, exists := ed.clients[clientID]; exists {
		close(clientChan)
		delete(ed.clients, clientID)
		logger.Debug("Removed event dispatcher client", zap.String("client_id", clientID))
	}
}

// GetClientCount returns the number of active clients
func (ed *EventDispatcher) GetClientCount() int {
	ed.clientsMu.RLock()
	defer ed.clientsMu.RUnlock()
	return len(ed.clients)
}

// verifyChangefeedSupport checks if the database supports changefeeds
func (ed *EventDispatcher) verifyChangefeedSupport() error {
	rows, err := ed.db.Pool.Query(ed.ctx, "SHOW CLUSTER SETTING cluster.organization")
	if err != nil {
		return fmt.Errorf("failed to check changefeed support: %w", err)
	}
	defer rows.Close()

	// If we can query cluster settings, changefeeds should be supported
	logger.Debug("Changefeed support verified")
	return nil
}

// listenToChangefeed starts the cross-node synchronization listener
func (ed *EventDispatcher) listenToChangefeed() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Cross-node sync listener crashed", zap.Any("error", r))
		}
	}()

	for {
		select {
		case <-ed.ctx.Done():
			logger.Info("Cross-node sync listener stopped")
			return
		default:
			if err := ed.runChangefeed(); err != nil {
				logger.Error("Cross-node sync error, retrying in 10 seconds", zap.Error(err))
				select {
				case <-ed.ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}
		}
	}
}

// runChangefeed implements cross-node event synchronization using polling
func (ed *EventDispatcher) runChangefeed() error {
	logger.Info("Starting cross-node event polling for distributed synchronization...")

	// Track the latest timestamp we've seen to avoid duplicates
	var lastSeen = time.Now().Unix()

	// Create a ticker for polling new events
	ticker := time.NewTicker(2 * time.Second) // Poll every 2 seconds
	defer ticker.Stop()

	logger.Info("✅ Cross-node polling started, checking for events every 2s...")

	for {
		select {
		case <-ed.ctx.Done():
			logger.Info("Cross-node polling stopping due to context cancellation")
			return nil
		case <-ticker.C:
			// Query for events created after our last seen timestamp
			currentTime := time.Now().Unix()

			query := `
				SELECT id, pubkey, kind, created_at, content, tags, sig 
				FROM events 
				WHERE created_at > $1 AND created_at <= $2
				ORDER BY created_at ASC`

			rows, err := ed.db.Pool.Query(ed.ctx, query, lastSeen, currentTime)
			if err != nil {
				logger.Error("Failed to query for new events", zap.Error(err))
				continue
			}

			newEventsCount := 0
			for rows.Next() {
				var eventData EventRowData
				err := rows.Scan(
					&eventData.ID,
					&eventData.PubKey,
					&eventData.Kind,
					&eventData.CreatedAt,
					&eventData.Content,
					&eventData.Tags,
					&eventData.Sig,
				)
				if err != nil {
					logger.Error("Failed to scan event row", zap.Error(err))
					continue
				}

				// Convert to Nostr event
				event, err := eventData.ToNostrEvent()
				if err != nil {
					logger.Warn("Failed to convert event to Nostr event", zap.Error(err))
					continue
				}

				logger.Debug("Found new cross-node event",
					zap.String("event_id", event.ID),
					zap.String("pubkey", event.PubKey),
					zap.Int("kind", event.Kind),
					zap.Int64("created_at", event.CreatedAt.Time().Unix()))

				// Send to event buffer for processing
				select {
				case ed.eventBuffer <- event:
					newEventsCount++
				default:
					logger.Warn("Event buffer full, dropping cross-node event", zap.String("event_id", event.ID))
				}
			}
			rows.Close()

			if newEventsCount > 0 {
				logger.Info("Synchronized cross-node events",
					zap.Int("count", newEventsCount),
					zap.Int64("time_range", currentTime-lastSeen))
			}

			// Update our last seen timestamp
			lastSeen = currentTime
		}
	}
}

// processEvents processes events from the buffer and broadcasts them to clients
func (ed *EventDispatcher) processEvents() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var batch []*nostr.Event

	for {
		select {
		case <-ed.ctx.Done():
			return
		case event := <-ed.eventBuffer:
			batch = append(batch, event)
		case <-ticker.C:
			if len(batch) > 0 {
				ed.broadcastEvents(batch)
				batch = batch[:0] // Clear batch
			}
		}
	}
}

// broadcastEvents sends events to all registered clients
func (ed *EventDispatcher) broadcastEvents(events []*nostr.Event) {
	ed.clientsMu.RLock()
	clientCount := len(ed.clients)
	ed.clientsMu.RUnlock()

	if len(events) > 0 {
		logger.Info("Broadcasting events to clients",
			zap.Int("event_count", len(events)),
			zap.Int("client_count", clientCount))
	}

	ed.clientsMu.RLock()
	defer ed.clientsMu.RUnlock()

	for clientID, clientChan := range ed.clients {
		for _, event := range events {
			select {
			case clientChan <- event:
				logger.Debug("Event sent to client successfully",
					zap.String("client_id", clientID),
					zap.String("event_id", event.ID))
			default:
				// Client buffer is full, drop the event
				logger.Warn("Dropped event for client - buffer full",
					zap.String("client_id", clientID),
					zap.String("event_id", event.ID))
			}
		}
	}
}
