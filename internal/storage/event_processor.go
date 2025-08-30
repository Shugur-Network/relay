package storage

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/Shugur-Network/relay/internal/relay/nips"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// EventProcessor manages event processing with a worker pool
type EventProcessor struct {
	eventChan   chan nostr.Event
	db          *DB
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(ctx context.Context, db *DB, bufferSize int) *EventProcessor {
	ctx, cancel := context.WithCancel(ctx)

	// Use CPU count to determine worker count
	workerCount := runtime.NumCPU() * 2

	ep := &EventProcessor{
		eventChan:   make(chan nostr.Event, bufferSize),
		db:          db,
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		go ep.processEvents(ctx)
	}

	return ep
}

// QueueDeletion is called by the validator AFTER it has verified
// that the deleter has the right to try.  The function will:
//  1. delete all owned referenced events (same pubkey)
//  2. store the deletion event itself
//
// It reuses the same retry / back‑pressure mechanism.
func (ep *EventProcessor) QueueDeletion(evt nostr.Event) bool {
	select {
	case ep.eventChan <- evt:
		return true
	default:
		logger.Warn("Deletion queue full, dropping event",
			zap.String("event_id", evt.ID),
			zap.String("pubkey", evt.PubKey),
			zap.Int("kind", evt.Kind))
		return false
	}
}

// QueueEvent adds an event to processing queue with non-blocking behavior
func (ep *EventProcessor) QueueEvent(evt nostr.Event) bool {
	// Check bloom filter first to avoid processing duplicates
	if ep.db.Bloom.Test([]byte(evt.ID)) {
		return true // Already processed, consider it "queued"
	}

	// Try to add to queue non-blocking
	select {
	case ep.eventChan <- evt:
		return true
	default:
		// Queue full - this is backpressure
		logger.Warn("Event processing queue full, dropping event",
			zap.String("event_id", evt.ID),
			zap.String("pubkey", evt.PubKey),
			zap.Int("kind", evt.Kind))
		return false
	}
}

// processEvents handles database insertion with retries
func (ep *EventProcessor) processEvents(ctx context.Context) {
	for {
		select {
		case <-ep.ctx.Done():
			return
		case evt, ok := <-ep.eventChan:
			if !ok {
				// Channel closed
				return
			}

			// Process with retries and backoff
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				if attempt > 0 {
					// Exponential backoff
					backoff := time.Duration(1<<attempt) * 50 * time.Millisecond
					time.Sleep(backoff)
				}

				ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
				switch {
				case nips.IsEphemeral(evt.Kind):
					// Ephemeral events (NIP-16) should not be stored
					logger.Debug("Skipping storage of ephemeral event",
						zap.String("event_id", evt.ID),
						zap.Int("kind", evt.Kind))
					err = nil // No error, just don't store
				case nips.IsDeletionEvent(evt):
					err = ep.db.persistDeletion(ctx, evt)
				case nips.IsReplaceable(evt.Kind):
					err = ep.db.InsertReplaceableEvent(ctx, evt)
				case nips.IsAddressable(evt):
					err = ep.db.InsertAddressableEvent(ctx, evt)
				default:
					err = ep.db.InsertEvent(ctx, evt)
				}
				cancel()

				if err == nil || strings.Contains(err.Error(), "duplicate key") {
					// For ephemeral events, skip bloom filter and metrics but still broadcast
					if nips.IsEphemeral(evt.Kind) {
						// Broadcast ephemeral event immediately to local clients for real-time streaming
						if ep.db.eventDispatcher != nil {
							logger.Debug("Broadcasting ephemeral event to local clients",
								zap.String("event_id", evt.ID),
								zap.String("pubkey", evt.PubKey),
								zap.Int("kind", evt.Kind))

							// Send event to local event dispatcher for immediate broadcasting
							select {
							case ep.db.eventDispatcher.eventBuffer <- &evt:
								logger.Debug("Ephemeral event added to local broadcast buffer", zap.String("event_id", evt.ID))
							default:
								logger.Warn("Local broadcast buffer full, ephemeral event may not stream immediately", zap.String("event_id", evt.ID))
							}
						}
					} else {
						// Only add to bloom filter after successful insertion for non-ephemeral events
						ep.db.Bloom.AddString(evt.ID)

						// Increment the stored events metric only for new events
						if err == nil {
							metrics.EventsStored.Inc()

							// Broadcast event immediately to local clients for real-time streaming
							// This ensures same-node clients get events instantly without waiting for changefeed
							if ep.db.eventDispatcher != nil {
								logger.Debug("Broadcasting event to local clients",
									zap.String("event_id", evt.ID),
									zap.String("pubkey", evt.PubKey),
									zap.Int("kind", evt.Kind))

								// Send event to local event dispatcher for immediate broadcasting
								select {
								case ep.db.eventDispatcher.eventBuffer <- &evt:
									logger.Debug("Event added to local broadcast buffer", zap.String("event_id", evt.ID))
								default:
									logger.Warn("Local broadcast buffer full, event may not stream immediately", zap.String("event_id", evt.ID))
								}
							}
						}
					}

					err = nil
					break
				}
			}

			if err != nil {
				logger.Error("Failed to insert event after retries",
					zap.String("event_id", evt.ID),
					zap.String("pubkey", evt.PubKey),
					zap.Int("kind", evt.Kind),
					zap.Error(err))
			} else {
				logger.Debug("Event successfully processed",
					zap.String("event_id", evt.ID),
					zap.String("pubkey", evt.PubKey),
					zap.Int("kind", evt.Kind))
			}
		}
	}
}

// Shutdown gracefully stops processing
func (ep *EventProcessor) Shutdown() {
	ep.cancel()
	// Don't close the channel as it might be in use
}
