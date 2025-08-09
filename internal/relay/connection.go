package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shugur-Network/Relay/internal/config"
	"github.com/Shugur-Network/Relay/internal/domain"
	"github.com/Shugur-Network/Relay/internal/logger"
	"github.com/Shugur-Network/Relay/internal/metrics"
	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var (
	clientBanList = make(map[string]time.Time)
	banListMutex  sync.Mutex
	// Track rate-limit violations by IP
	clientExceededCount = make(map[string]int)
)

// normalizeIP converts a network address to a normalized IP string
func normalizeIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return addr
	}
	return ip.String()
}

// cleanExpiredBans periodically removes expired bans from the ban list
func cleanExpiredBans() {
	for {
		time.Sleep(10 * time.Minute)

		banListMutex.Lock()
		now := time.Now()
		for ip, expiry := range clientBanList {
			if now.After(expiry) {
				delete(clientBanList, ip)
			}
		}
		banListMutex.Unlock()
	}
}

// handleWebSocketConnection handles the upgrade of an HTTP connection to WebSocket
func handleWebSocketConnection(ctx context.Context, w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader, node domain.NodeInterface, relayConfig config.RelayConfig) {
	clientIP := normalizeIP(r.RemoteAddr)

	// Check if client is banned
	banListMutex.Lock()
	banExpiry, banned := clientBanList[clientIP]
	banListMutex.Unlock()

	if banned && time.Now().Before(banExpiry) {
		logger.Warn("Banned client attempted to reconnect",
			zap.String("client", clientIP),
		)
		http.Error(w, "You are temporarily banned due to excessive messages.", http.StatusForbidden)
		return
	}

	// Reset exceeded count on new allowed connection
	banListMutex.Lock()
	delete(clientExceededCount, clientIP)
	banListMutex.Unlock()

	// Check global connection limit using metrics counter
	if metrics.GetActiveConnectionsCount() >= int64(relayConfig.ThrottlingConfig.MaxConnections) {
		metrics.ErrorsCount.WithLabelValues("max_connections").Inc()
		logger.Warn("Max connections reached, rejecting new WebSocket connection")
		http.Error(w, "Max connections reached", http.StatusServiceUnavailable)
		return
	}
	// Ensure we decrement on error
	connectionSuccess := false
	defer func() {
		if !connectionSuccess {
			metrics.DecrementActiveConnections()
		}
	}()

	// Upgrade the connection
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}

	// Enable compression
	wsConn.EnableWriteCompression(true)
	_ = wsConn.SetCompressionLevel(2) // nolint:errcheck // compression level is non-critical

	// Update metrics
	metrics.IncrementActiveConnections()
	connectionSuccess = true

	// Create new connection and register it
	conn := NewWsConnection(ctx, wsConn, node, relayConfig)
	node.RegisterConn(conn)

	// Handle messages in a goroutine
	go conn.HandleMessages(ctx, relayConfig)
}

// WsConnection represents a single WebSocket client connection
type WsConnection struct {
	ws           *websocket.Conn
	node         domain.NodeInterface
	lastActivity time.Time
	idleTimeout  time.Duration
	maxLifetime  time.Duration // Maximum lifetime of a connection
	startTime    time.Time     // When the connection was established

	pingTicker *time.Ticker

	subMu         sync.RWMutex
	subscriptions map[string][]nostr.Filter

	writeMu            sync.Mutex
	closeMu            sync.Once
	limiter            *rate.Limiter
	isClosed           atomic.Bool
	metricsDecremented atomic.Bool // Flag to prevent double-decrementing metrics
	closeReason        string

	exceededLimitCount int
	backpressureChan   chan struct{} // Channel for backpressure handling
}

// Ensure WsConnection implements domain.WebSocketConnection
var _ domain.WebSocketConnection = (*WsConnection)(nil)

// NewWsConnection initializes a new WebSocket connection
func NewWsConnection(
	ctx context.Context,
	ws *websocket.Conn,
	node domain.NodeInterface,
	cfg config.RelayConfig,
) *WsConnection {
	// Basic rate limiter
	limiter := rate.NewLimiter(
		rate.Limit(cfg.ThrottlingConfig.RateLimit.MaxEventsPerSecond),
		cfg.ThrottlingConfig.RateLimit.BurstSize,
	)

	conn := &WsConnection{
		ws:               ws,
		node:             node,
		idleTimeout:      cfg.IdleTimeout,
		maxLifetime:      24 * time.Hour, // Maximum connection lifetime
		startTime:        time.Now(),
		lastActivity:     time.Now(),
		subscriptions:    make(map[string][]nostr.Filter),
		pingTicker:       time.NewTicker(30 * time.Second),
		limiter:          limiter,
		backpressureChan: make(chan struct{}, 100), // Buffer for backpressure
	}

	// WebSocket compression
	ws.EnableWriteCompression(true)
	_ = ws.SetCompressionLevel(2) // nolint:errcheck // compression level is non-critical

	// Deadlines + read limit
	_ = ws.SetReadDeadline(time.Now().Add(120 * time.Second)) // nolint:errcheck // deadline is non-critical
	ws.SetReadLimit(8 * 1024 * 1024)                          // 8MB

	// Ping handler
	ws.SetPingHandler(func(appData string) error {
		conn.lastActivity = time.Now()
		conn.writeMu.Lock()
		defer conn.writeMu.Unlock()
		_ = conn.ws.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
		return nil
	})

	// Start monitoring
	go conn.monitorConnection(ctx)

	return conn
}

// RemoteAddr returns the client's remote address
func (c *WsConnection) RemoteAddr() string {
	return c.ws.RemoteAddr().String()
}

// SendMessage handles backpressure and rate limiting
func (c *WsConnection) SendMessage(msg []byte) {
	c.sendMessageInternal(msg, true)
}

// SendMessageNoRateLimit sends a message without rate limiting (for subscription responses)
func (c *WsConnection) SendMessageNoRateLimit(msg []byte) {
	c.sendMessageInternal(msg, false)
}

// sendMessageInternal handles the actual message sending with optional rate limiting
func (c *WsConnection) sendMessageInternal(msg []byte, applyRateLimit bool) {
	if c.isClosed.Load() {
		return
	}

	// Check backpressure
	select {
	case c.backpressureChan <- struct{}{}:
		defer func() { <-c.backpressureChan }()
	default:
		// Backpressure is too high, close connection
		c.Close()
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.isClosed.Load() {
		return
	}

	// Apply rate limiting only if requested
	if applyRateLimit && !c.limiter.Allow() {
		c.exceededLimitCount++
		if c.exceededLimitCount > 5 {
			c.Close()
			return
		}
		return
	}

	// Reset exceeded count on successful send
	c.exceededLimitCount = 0

	// Set write deadline
	_ = c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second)) // nolint:errcheck // deadline is non-critical
	if err := c.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		logger.Error("Failed to write message", zap.Error(err))
		c.Close()
	}

	// Update metrics
	metrics.MessagesSent.Inc()
	metrics.MessageSizeBytesSent.Observe(float64(len(msg)))
}

// sendMessage marshals a top-level array like ["NOTICE", "xyz"] or ["CLOSED", subID, reason].
func (c *WsConnection) sendMessage(msgType string, args ...interface{}) {
	data := append([]interface{}{msgType}, args...)
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Warn("Failed to marshal message", zap.Error(err))
		return
	}

	// Bypass rate limiting for EVENT and COUNT responses (subscription data)
	if msgType == "EVENT" || msgType == "COUNT" {
		c.SendMessageNoRateLimit(raw)
	} else {
		c.SendMessage(raw)
	}
}

// sendNotice is a convenience for sending ["NOTICE", <message>].
func (c *WsConnection) sendNotice(message string) {
	c.sendMessage("NOTICE", message)
}

// sendClosed is a convenience for sending ["CLOSED", <subID>, <reason>].
func (c *WsConnection) sendClosed(subID, reason string) {
	c.sendMessage("CLOSED", subID, reason)
}

// sendOK sends an OK response for an event with status and message
func (c *WsConnection) sendOK(eventID string, accepted bool, message string) {
	msg := []interface{}{"OK", eventID, accepted, message}
	data, _ := json.Marshal(msg)
	c.SendMessage(data)
}

// sendEOSE sends an EOSE (End of Stored Events) message
func (c *WsConnection) sendEOSE(subID string) {
	c.sendMessage("EOSE", subID)
}

// HandleMessages processes incoming messages from the client
func (c *WsConnection) HandleMessages(ctx context.Context, cfg config.RelayConfig) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered from panic in HandleMessages",
				zap.Any("panic", r),
				zap.String("client", c.ws.RemoteAddr().String()),
			)
		}
		// Always ensure connection is properly closed and unregistered
		c.closeReason = "message handler terminated"
		c.Close()
		c.node.UnregisterConn(c)
	}()

	clientIP := normalizeIP(c.ws.RemoteAddr().String())

	// Check if client is banned
	banListMutex.Lock()
	banExpiry, banned := clientBanList[clientIP]
	banListMutex.Unlock()

	if banned && time.Now().Before(banExpiry) {
		logger.Warn("Banned client attempted to send messages", zap.String("client", clientIP))
		c.closeReason = "client banned"
		c.sendNotice("You are temporarily banned due to excessive messages.")
		c.Close()
		return
	}

	c.ws.SetReadLimit(16 * 1024 * 1024) // Increase limit for large events

	lastPong := time.Now()
	c.ws.SetPongHandler(func(string) error {
		c.lastActivity = time.Now()
		lastPong = time.Now()
		return nil
	})

	connCtx, cancel := context.WithTimeout(ctx, 24*time.Hour)
	defer cancel()

	for {
		select {
		case <-connCtx.Done():
			c.closeReason = "connection context canceled"
			return
		default:
			// Keep going
		}

		_ = c.ws.SetReadDeadline(time.Now().Add(120 * time.Second)) // nolint:errcheck // deadline is non-critical
		if time.Since(lastPong) > 180*time.Second {
			logger.Debug("No pong in 180s, closing connection",
				zap.String("client", c.ws.RemoteAddr().String()))
			c.closeReason = "no pong response"
			return
		}

		// Read message
		_, rawMsg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.closeReason = "client closed connection"
				logger.Debug("Client closed connection normally",
					zap.String("client", c.ws.RemoteAddr().String()))
			} else {
				c.closeReason = "read error"
				logger.Debug("WS read error, disconnecting client",
					zap.Error(err),
					zap.String("client", c.ws.RemoteAddr().String()))
			}
			return
		}

		// Update metrics
		metrics.IncrementMessagesProcessed() // This handles both counter and local tracking
		messageSize := float64(len(rawMsg))
		metrics.MessageSizeBytes.Observe(messageSize)

		_ = c.ws.SetReadDeadline(time.Time{}) // nolint:errcheck // deadline reset is non-critical
		c.lastActivity = time.Now()

		var arr []interface{}
		if err := json.Unmarshal(rawMsg, &arr); err != nil {
			c.sendNotice("invalid: malformed JSON from client")
			continue
		}
		if len(arr) == 0 {
			c.sendNotice("invalid: empty command array")
			continue
		}

		cmdType, ok := arr[0].(string)
		if !ok {
			c.sendNotice("invalid: command must be a string")
			continue
		}

		if cmdType == "EVENT" {
			if !c.limiter.Allow() {
				// Track repeated violations
				banListMutex.Lock()
				clientExceededCount[clientIP]++
				count := clientExceededCount[clientIP]
				banListMutex.Unlock()

				logger.Debug("Client exceeded message rate limit",
					zap.String("client", clientIP),
					zap.Int("count", count))

				c.sendNotice("Rate limit exceeded: too many messages")

				if count >= cfg.ThrottlingConfig.BanThreshold {
					logger.Debug("Banning client due to repeated rate limit violations",
						zap.String("client", clientIP),
					)
					banListMutex.Lock()
					clientBanList[clientIP] = time.Now().Add(time.Duration(cfg.ThrottlingConfig.BanDuration) * time.Second)
					delete(clientExceededCount, clientIP)
					banListMutex.Unlock()

					c.sendNotice("You have been temporarily banned.")
					c.Close()
					return
				}
				continue
			}
			// Reset exceeded count on successful message
			c.exceededLimitCount = 0
		}

		// Update command metrics
		metrics.CommandsReceived.WithLabelValues(cmdType).Inc()

		// Process the command
		start := time.Now()
		switch cmdType {
		case "EVENT":
			c.handleEvent(ctx, arr)
		case "REQ":
			c.handleRequest(ctx, arr)
		case "COUNT":
			c.handleCountRequest(ctx, arr)
		case "CLOSE":
			c.handleClose(arr)
		default:
			c.sendNotice("invalid: unknown command '" + cmdType + "'")
		}
		metrics.CommandProcessingDuration.WithLabelValues(cmdType).Observe(time.Since(start).Seconds())
	}
}

// Close gracefully shuts down the WebSocket
func (c *WsConnection) Close() {
	c.closeMu.Do(func() {
		c.isClosed.Store(true)

		if c.closeReason != "" {
			logger.Info("WebSocket connection closed",
				zap.String("reason", c.closeReason),
				zap.String("client", c.ws.RemoteAddr().String()))
		}

		// Clear any subscriptions
		c.subMu.Lock()
		oldSubs := len(c.subscriptions)
		c.subscriptions = make(map[string][]nostr.Filter)
		c.subMu.Unlock()

		// Update metrics - only decrement once
		if !c.metricsDecremented.Swap(true) {
			metrics.ActiveSubscriptions.Sub(float64(oldSubs))
			metrics.DecrementActiveConnections()
		}

		if c.pingTicker != nil {
			c.pingTicker.Stop()
		}

		// Attempt a polite close
		closeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		closeChan := make(chan struct{})
		go func() {
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, c.closeReason)
			c.writeMu.Lock()
			_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second))
			_ = c.ws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
			_ = c.ws.SetWriteDeadline(time.Time{})
			c.writeMu.Unlock()
			close(closeChan)
		}()

		select {
		case <-closeChan:
		case <-closeCtx.Done():
			logger.Debug("Close message timeout",
				zap.String("client", c.ws.RemoteAddr().String()))
		}

		// Unregister
		c.node.UnregisterConn(c)

		// Finally close
		_ = c.ws.Close()
		logger.Debug("WebSocket connection cleanup completed",
			zap.String("client", c.ws.RemoteAddr().String()))
	})
}

// monitorConnection handles connection timeouts and cleanup
func (c *WsConnection) monitorConnection(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		case <-ticker.C:
			now := time.Now()
			c.writeMu.Lock()

			// Check idle timeout
			if now.Sub(c.lastActivity) > c.idleTimeout {
				c.writeMu.Unlock()
				c.Close()
				return
			}

			// Check max lifetime
			if now.Sub(c.startTime) > c.maxLifetime {
				c.writeMu.Unlock()
				c.Close()
				return
			}

			// Check backpressure
			if len(c.backpressureChan) > 90 { // 90% of buffer capacity
				c.writeMu.Unlock()
				c.Close()
				return
			}

			c.writeMu.Unlock()
		}
	}
}

// Subscription management methods

// HasSubscription checks if a subscription exists
func (c *WsConnection) HasSubscription(subID string) bool {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	_, ok := c.subscriptions[subID]
	return ok
}

// AddSubscription adds a new subscription
func (c *WsConnection) AddSubscription(subID string, filters []nostr.Filter) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	c.subscriptions[subID] = filters
}

// RemoveSubscription removes a subscription
func (c *WsConnection) RemoveSubscription(subID string) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	delete(c.subscriptions, subID)
}

// handleEvent processes EVENT commands
func (c *WsConnection) handleEvent(ctx context.Context, arr []interface{}) {
	if len(arr) < 2 {
		c.sendNotice("Invalid event message: not enough elements")
		return
	}

	// Marshal the event data back to JSON
	eventData, err := json.Marshal(arr[1])
	if err != nil {
		c.sendNotice("Invalid event: " + err.Error())
		return
	}

	var evt nostr.Event
	if err := json.Unmarshal(eventData, &evt); err != nil {
		c.sendNotice("Invalid event: " + err.Error())
		return
	}

	// Use ValidateAndProcessEvent for comprehensive validation
	valid, msg, err := c.node.GetValidator().ValidateAndProcessEvent(ctx, evt)
	if err != nil {
		c.sendOK(evt.ID, false, "error: "+err.Error())
		return
	}
	if !valid {
		c.sendOK(evt.ID, false, msg)
		return
	}

	// Queue the event for processing
	if ok := c.node.GetEventProcessor().QueueEvent(evt); !ok {
		c.sendOK(evt.ID, false, "server busy, try again")
		return
	}

	// Update metrics for successful event
	metrics.EventsProcessed.WithLabelValues(fmt.Sprintf("%d", evt.Kind)).Inc()

	// Send successful response
	c.sendOK(evt.ID, true, "")
}

// QueryEvents reads events from storage that match a given Nostr filter.
func (c *WsConnection) QueryEvents(ctx context.Context, f nostr.Filter) ([]nostr.Event, error) {
	logger.Debug("QueryEvents called with filter", zap.Any("filter", f))

	results, err := c.node.DB().GetEvents(ctx, f)
	if err != nil {
		logger.Error("Error retrieving events from storage", zap.Error(err))
		return nil, err
	}
	return results, nil
}
