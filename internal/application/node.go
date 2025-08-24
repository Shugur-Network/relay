package application

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/domain"
	"github.com/Shugur-Network/relay/internal/limiter"
	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/relay"
	"github.com/Shugur-Network/relay/internal/storage"
	"github.com/Shugur-Network/relay/internal/workers"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// Node ties together the various components needed to run the Shugur node.
type Node struct {
	ctx    context.Context
	cancel context.CancelFunc

	db             *storage.DB
	config         *config.Config
	WorkerPool     *workers.WorkerPool
	EventProcessor *storage.EventProcessor
	EventDispatcher *storage.EventDispatcher
	Validator      domain.EventValidator
	EventValidator *relay.EventValidator

	wsConns   map[domain.WebSocketConnection]bool
	wsConnsMu sync.RWMutex

	blacklistPubKeys map[string]struct{}
	whitelistPubKeys map[string]struct{}

	rateLimiter *limiter.RateLimiter
}

// Ensure Node implements domain.NodeInterface
var _ domain.NodeInterface = (*Node)(nil)

// New creates and configures a Node using the NodeBuilder pattern.
func New(ctx context.Context, cfg *config.Config, privKey ed25519.PrivateKey) (*Node, error) {
	// 1) Construct a NodeBuilder
	builder := NewNodeBuilder(ctx, cfg, privKey)

	// 2) Build DB first
	if err := builder.BuildDB(); err != nil {
		return nil, fmt.Errorf("failed building db: %w", err)
	}

	// 3) Build worker pool
	builder.BuildWorkers()

	// 4) Build validators
	builder.BuildValidators()

	// 5) Build event processor
	builder.BuildProcessor()

	// 6) Build rate limiter
	builder.BuildRateLimiter()

	// 7) Build black/white lists
	builder.BuildLists()

	// 8) Finally assemble the Node
	node := builder.Build()
	return node, nil
}

// Start begins the main loops for the node:
// Starts the relay server with integrated web dashboard
func (n *Node) Start(ctx context.Context) error {
	// Start the event dispatcher for real-time notifications
	if err := n.EventDispatcher.Start(); err != nil {
		logger.Error("Failed to start event dispatcher", zap.Error(err))
		return err
	}

	// Start the relay server (now includes web dashboard)
	go func() {
		addr := n.config.Relay.WSAddr
		server := relay.NewServer(n.config.Relay, n, n.config)
		if err := server.ListenAndServe(n.ctx, addr); err != nil {
			// Don't log "Server closed" as an error - it's expected during graceful shutdown
			if err.Error() != "http: Server closed" {
				logger.Error("Server error", zap.Error(err))
			} else {
				logger.Debug("Server closed gracefully", zap.Error(err))
			}
		}
	}()

	logger.Debug("Node started with integrated web dashboard and event dispatcher")
	return nil
}

// Shutdown cancels the node context and closes all resources.
func (n *Node) Shutdown() {
	logger.Debug("Shutting down node...")

	// Stop the event dispatcher
	if n.EventDispatcher != nil {
		n.EventDispatcher.Stop()
	}

	// Shut down the EventProcessor
	if n.EventProcessor != nil {
		n.EventProcessor.Shutdown()
	}

	// Wait for all WorkerPool tasks to finish
	n.WorkerPool.Wait()

	// Cancel the context
	if n.cancel != nil {
		n.cancel()
	}

	var shutdownErrors []error

	// Close DB with retry mechanism
	if n.db != nil {
		const maxRetries = 3
		const retryDelay = 1 * time.Second
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			if err := n.db.CloseDB(); err != nil {
				lastErr = err
				logger.Warn("Failed to close database, retrying...",
					zap.Int("attempt", i+1),
					zap.Int("max_attempts", maxRetries),
					zap.Error(err))
				time.Sleep(retryDelay)
				continue
			}
			lastErr = nil
			break
		}

		if lastErr != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("storage shutdown error after %d retries: %w", maxRetries, lastErr))
			logger.Error("Failed to close database after multiple attempts", zap.Error(lastErr))
		}
	}

	if len(shutdownErrors) > 0 {
		logger.Warn("Node shutdown completed with errors",
			zap.Int("error_count", len(shutdownErrors)),
			zap.Errors("errors", shutdownErrors))
	} else {
		logger.Debug("Node shut down successfully")
	}
}

// RegisterConn tracks a new WebSocket client
func (n *Node) RegisterConn(conn domain.WebSocketConnection) {
	n.wsConnsMu.Lock()
	defer n.wsConnsMu.Unlock()
	n.wsConns[conn] = true
	count := len(n.wsConns)
	logger.Debug("WebSocket client registered", zap.Int("total_connections", count))
}

// UnregisterConn removes a WebSocket client
func (n *Node) UnregisterConn(conn domain.WebSocketConnection) {
	n.wsConnsMu.Lock()
	defer n.wsConnsMu.Unlock()
	delete(n.wsConns, conn)
	count := len(n.wsConns)
	logger.Debug("WebSocket client unregistered", zap.Int("total_connections", count))
}

// GetActiveConnectionCount returns the actual number of active WebSocket connections
func (n *Node) GetActiveConnectionCount() int64 {
	n.wsConnsMu.RLock()
	defer n.wsConnsMu.RUnlock()
	return int64(len(n.wsConns))
}

// GetEventCount returns the count of events matching the given filter
func (n *Node) GetEventCount(ctx context.Context, filter nostr.Filter) (int64, error) {
	return n.db.GetEventCount(ctx, filter)
}
