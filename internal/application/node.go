package application

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/constants"
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

	db              *storage.DB
	config          *config.Config
	WorkerPool      *workers.WorkerPool
	EventProcessor  *storage.EventProcessor
	EventDispatcher *storage.EventDispatcher
	Validator       domain.EventValidator
	EventValidator  *relay.EventValidator

	wsConns   map[domain.WebSocketConnection]bool
	wsConnsMu sync.RWMutex

	blacklistPubKeys map[string]struct{}
	whitelistPubKeys map[string]struct{}

	rateLimiter *limiter.RateLimiter
	startTime   time.Time
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
	node, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build node: %w", err)
	}
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

// Shutdown gracefully shuts down the node with configurable timeout.
func (n *Node) Shutdown() {
	logger.Info("Initiating graceful shutdown...")
	shutdownTimeout := 30 * time.Second // Hardcoded 30-second timeout

	// Create a timeout context for shutdown operations
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	var shutdownErrors []error

	// Step 1: Stop accepting new connections and close existing WebSocket connections gracefully
	n.shutdownWebSocketConnections(shutdownCtx)

	// Step 2: Stop the event dispatcher
	if n.EventDispatcher != nil {
		logger.Debug("Stopping event dispatcher...")
		n.EventDispatcher.Stop()
		logger.Debug("✅ Event dispatcher stopped")
	}

	// Step 3: Shut down the EventProcessor
	if n.EventProcessor != nil {
		logger.Debug("Shutting down event processor...")
		n.EventProcessor.Shutdown()
		logger.Debug("✅ Event processor stopped")
	}

	// Step 4: Wait for all WorkerPool tasks to finish with timeout
	logger.Debug("Waiting for worker pool to finish...")
	done := make(chan struct{})
	go func() {
		defer close(done)
		n.WorkerPool.Wait()
	}()

	select {
	case <-done:
		logger.Debug("✅ Worker pool finished")
	case <-shutdownCtx.Done():
		shutdownErrors = append(shutdownErrors, fmt.Errorf("worker pool shutdown timed out after %v", shutdownTimeout))
		logger.Warn("Worker pool shutdown timed out", zap.Duration("timeout", shutdownTimeout))
	}

	// Step 5: Cancel the node context
	if n.cancel != nil {
		logger.Debug("Canceling node context...")
		n.cancel()
		logger.Debug("✅ Node context canceled")
	}

	// Step 6: Close DB with retry mechanism and timeout
	if n.db != nil {
		logger.Debug("Closing database connection...")
		if err := n.shutdownDatabase(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, err)
		} else {
			logger.Debug("✅ Database connection closed")
		}
	}

	// Report final shutdown status
	if len(shutdownErrors) > 0 {
		logger.Warn("Node shutdown completed with errors",
			zap.Int("error_count", len(shutdownErrors)),
			zap.Errors("errors", shutdownErrors),
			zap.Duration("shutdown_timeout", shutdownTimeout))
	} else {
		logger.Info("✅ Node shutdown completed successfully",
			zap.Duration("shutdown_timeout", shutdownTimeout))
	}
}

// shutdownWebSocketConnections gracefully closes all active WebSocket connections.
func (n *Node) shutdownWebSocketConnections(ctx context.Context) {
	n.wsConnsMu.Lock()
	connectionCount := len(n.wsConns)
	connections := make([]domain.WebSocketConnection, 0, connectionCount)
	for conn := range n.wsConns {
		connections = append(connections, conn)
	}
	n.wsConnsMu.Unlock()

	if connectionCount == 0 {
		logger.Debug("✅ No WebSocket connections to close")
		return
	}

	logger.Info("Closing WebSocket connections gracefully",
		zap.Int("connection_count", connectionCount))

	// Close connections gracefully with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		
		// Close all connections - the connection.Close() method handles graceful closure
		for _, conn := range connections {
			conn.Close()
		}

		// Clear the connections map
		n.wsConnsMu.Lock()
		n.wsConns = make(map[domain.WebSocketConnection]bool)
		n.wsConnsMu.Unlock()
	}()

	select {
	case <-done:
		logger.Debug("✅ WebSocket connections closed")
	case <-ctx.Done():
		logger.Warn("WebSocket connection shutdown timed out")
		// Force clear the map in case of timeout
		n.wsConnsMu.Lock()
		n.wsConns = make(map[domain.WebSocketConnection]bool)
		n.wsConnsMu.Unlock()
	}
}

// shutdownDatabase closes the database connection with timeout and retry logic.
func (n *Node) shutdownDatabase(ctx context.Context) error {
	var lastErr error

	for i := 0; i < constants.MaxDBRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("database shutdown timed out after %d attempts: %w", i, ctx.Err())
		default:
		}

		if err := n.db.CloseDB(); err != nil {
			lastErr = err
			logger.Warn("Failed to close database, retrying...",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", constants.MaxDBRetries),
				zap.Error(err))
			
			// Wait with context timeout awareness
			select {
			case <-time.After(constants.DBRetryDelay * time.Second):
				continue
			case <-ctx.Done():
				return fmt.Errorf("database shutdown timed out during retry delay: %w", ctx.Err())
			}
		}
		return nil // Success
	}

	return fmt.Errorf("database shutdown failed after %d retries: %w", constants.MaxDBRetries, lastErr)
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

// GetConnectionCount returns the current number of active connections (for health checks)
func (n *Node) GetConnectionCount() int {
	n.wsConnsMu.RLock()
	defer n.wsConnsMu.RUnlock()
	return len(n.wsConns)
}

// GetStartTime returns when the node was started (for health checks)
func (n *Node) GetStartTime() time.Time {
	return n.startTime
}
