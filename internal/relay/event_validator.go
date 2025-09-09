package relay

import (
	"sync"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/storage"
)

// EventValidator provides validation services for Nostr events
type EventValidator struct {
	validator   *PluginValidator
	rateLimiter *RateLimiter
	db          *storage.DB
}

// RateLimiter tracks event creation rates by pubkey
type RateLimiter struct {
	limitPerMin    int // Max events per minute for all kinds
	limitPerSec    int // Max events per second for all kinds
	countsByPubkey map[string]EventCounter
	mutex          sync.RWMutex
	windowSize     time.Duration
	cleanupTicker  *time.Ticker
	circuitBreaker map[string]time.Time // Tracks when to allow retry after circuit break
}

// EventCounter tracks event counts within a time window
type EventCounter struct {
	lastSeen time.Time // Track last activity
}

// NewEventValidator creates a new event validator instance
func NewEventValidator(cfg *config.Config, db *storage.DB) *EventValidator {
	// Create rate limiter with general limits
	limiter := &RateLimiter{
		limitPerMin:    cfg.Relay.ThrottlingConfig.RateLimit.MaxEventsPerSecond * 60,
		limitPerSec:    cfg.Relay.ThrottlingConfig.RateLimit.MaxEventsPerSecond,
		countsByPubkey: make(map[string]EventCounter),
		windowSize:     time.Minute,
		circuitBreaker: make(map[string]time.Time),
		cleanupTicker:  time.NewTicker(5 * time.Minute),
	}

	// Start cleanup goroutine
	go limiter.cleanupInactiveCounters()

	validator := &EventValidator{
		validator:   NewPluginValidator(cfg, db),
		db:          db,
		rateLimiter: limiter,
	}

	return validator
}

// cleanupInactiveCounters removes counters for inactive pubkeys
func (rl *RateLimiter) cleanupInactiveCounters() {
	for range rl.cleanupTicker.C {
		rl.mutex.Lock()
		now := time.Now()
		for pubkey, counter := range rl.countsByPubkey {
			if now.Sub(counter.lastSeen) > 30*time.Minute {
				delete(rl.countsByPubkey, pubkey)
			}
		}
		// Clean up expired circuit breakers
		for pubkey, expiry := range rl.circuitBreaker {
			if now.After(expiry) {
				delete(rl.circuitBreaker, pubkey)
			}
		}
		rl.mutex.Unlock()
	}
}
