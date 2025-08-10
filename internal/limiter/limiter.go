package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// RateLimit defines the limits for a specific rate limiting key
type RateLimit struct {
	MaxEvents    int           // Maximum number of events allowed
	WindowSize   time.Duration // Time window for the limit
	BurstSize    int           // Maximum burst size allowed
	BanThreshold int           // Number of violations before banning
	BanDuration  time.Duration // Duration of the ban
}

// Counter tracks rate limiting state for a specific key
type Counter struct {
	count       int       // Current count of events
	burstCount  int       // Current burst count
	lastReset   time.Time // Last time the counter was reset
	lastBurst   time.Time // Last time burst was reset
	banCount    int       // Number of times rate limit was exceeded
	lastBanTime time.Time // Last time a ban was issued
}

// RateLimiter manages rate limiting across different dimensions
type RateLimiter struct {
	limits map[string]RateLimit // key: "kind:pubkey" or "connection:ip"
	counts map[string]*Counter
	mutex  sync.RWMutex
	cfg    *config.Config
}

// NewRateLimiter creates a new rate limiter instance with default limits from config
func NewRateLimiter(cfg *config.Config) *RateLimiter {
	rl := &RateLimiter{
		limits: make(map[string]RateLimit),
		counts: make(map[string]*Counter),
		cfg:    cfg,
	}

	// Set default limits from config
	defaultLimit := RateLimit{
		MaxEvents:    cfg.Relay.ThrottlingConfig.RateLimit.MaxEventsPerSecond * 60, // 5000 events per minute
		WindowSize:   time.Minute,
		BurstSize:    cfg.Relay.ThrottlingConfig.RateLimit.BurstSize,
		BanThreshold: cfg.Relay.ThrottlingConfig.RateLimit.BanThreshold,
		BanDuration:  cfg.Relay.ThrottlingConfig.RateLimit.BanDuration,
	}

	// Set default limits for different types of events
	rl.SetLimit("default", defaultLimit)
	rl.SetLimit("connection", defaultLimit)
	rl.SetLimit("pubkey", defaultLimit)

	return rl
}

// SetLimit sets or updates a rate limit for a specific key
func (rl *RateLimiter) SetLimit(key string, limit RateLimit) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.limits[key] = limit
}

// Allow checks if an event should be allowed based on rate limits
func (rl *RateLimiter) Allow(key string, limit RateLimit) bool {
	// Skip rate limiting for empty keys (system events)
	if key == "" {
		return true
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	counter, exists := rl.counts[key]
	if !exists {
		counter = &Counter{
			lastReset: time.Now(),
		}
		rl.counts[key] = counter
	}

	// Reset counter if window expired
	if time.Since(counter.lastReset) > limit.WindowSize {
		counter.count = 0
		counter.lastReset = time.Now()
	}

	// Handle burst
	if time.Since(counter.lastBurst) > limit.WindowSize {
		counter.burstCount = 0
		counter.lastBurst = time.Now()
	}

	// Check if we're in burst mode
	if counter.burstCount < limit.BurstSize {
		counter.burstCount++
		counter.count++
		return true
	}

	// Normal rate limiting
	if counter.count < limit.MaxEvents {
		counter.count++
		return true
	}

	// Check for ban threshold
	if counter.banCount >= limit.BanThreshold {
		if time.Since(counter.lastBanTime) > limit.BanDuration {
			counter.banCount = 0
			counter.count = 0
			return true
		}

		// Log ban event
		logger.Warn("Rate limit exceeded, client banned",
			zap.String("key", key),
			zap.Int("ban_count", counter.banCount),
			zap.Duration("ban_duration", limit.BanDuration),
		)

		return false
	}

	counter.banCount++
	counter.lastBanTime = time.Now()

	// Log rate limit exceeded
	logger.Debug("Rate limit exceeded",
		zap.String("key", key),
		zap.Int("count", counter.count),
		zap.Int("burst_count", counter.burstCount),
	)

	return false
}

// Reset resets the counter for a specific key
func (rl *RateLimiter) Reset(key string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if counter, exists := rl.counts[key]; exists {
		counter.count = 0
		counter.burstCount = 0
		counter.banCount = 0
		counter.lastReset = time.Now()
		counter.lastBurst = time.Now()
		counter.lastBanTime = time.Now()
	}
}

// GetCounter returns the current counter state for a key
func (rl *RateLimiter) GetCounter(key string) *Counter {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	return rl.counts[key]
}

// Cleanup removes expired counters
func (rl *RateLimiter) Cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for key, counter := range rl.counts {
		// Remove counters that haven't been used in 24 hours
		if now.Sub(counter.lastReset) > 24*time.Hour {
			delete(rl.counts, key)
		}
	}
}

// String returns a string representation of the rate limiter state
func (rl *RateLimiter) String() string {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	var result string
	for key, counter := range rl.counts {
		result += fmt.Sprintf("Key: %s\n", key)
		result += fmt.Sprintf("  Count: %d\n", counter.count)
		result += fmt.Sprintf("  Burst Count: %d\n", counter.burstCount)
		result += fmt.Sprintf("  Ban Count: %d\n", counter.banCount)
		result += fmt.Sprintf("  Last Reset: %v\n", counter.lastReset)
		result += fmt.Sprintf("  Last Burst: %v\n", counter.lastBurst)
		result += fmt.Sprintf("  Last Ban: %v\n", counter.lastBanTime)
		result += "---\n"
	}
	return result
}

// GetLimit returns the rate limit for a specific key
func (rl *RateLimiter) GetLimit(key string) RateLimit {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	if limit, exists := rl.limits[key]; exists {
		return limit
	}

	// Return default limit if none exists
	return RateLimit{
		MaxEvents:    rl.cfg.Relay.ThrottlingConfig.RateLimit.MaxEventsPerSecond * 60, // 5000 events per minute
		WindowSize:   time.Minute,
		BurstSize:    rl.cfg.Relay.ThrottlingConfig.RateLimit.BurstSize,
		BanThreshold: rl.cfg.Relay.ThrottlingConfig.RateLimit.BanThreshold,
		BanDuration:  rl.cfg.Relay.ThrottlingConfig.RateLimit.BanDuration,
	}
}
