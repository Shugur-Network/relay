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
		logger.Warn("deletion queue full – dropping", zap.String("id", evt.ID))
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
			zap.String("id", evt.ID))
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
					// Only add to bloom filter after successful insertion
					ep.db.Bloom.AddString(evt.ID)

					// Increment the stored events metric only for new events
					if err == nil {
						metrics.EventsStored.Inc()
					}

					err = nil
					break
				}
			}

			if err != nil {
				logger.Warn("Failed to insert event after retries",
					zap.String("id", evt.ID),
					zap.Error(err))
			}
		}
	}
}

// Shutdown gracefully stops processing
func (ep *EventProcessor) Shutdown() {
	ep.cancel()
	// Don't close the channel as it might be in use
}
