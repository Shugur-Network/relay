package relay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/domain"
	"github.com/Shugur-Network/relay/internal/errors"
	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/gorilla/websocket"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var (
	clientBanList = make(map[string]time.Time)
	banListMutex  sync.Mutex
	// Track rate-limit violations by IP
	clientExceededCount = make(map[string]int)
)

// extractRealClientIP extracts the real client IP from request headers when behind a proxy
func extractRealClientIP(r *http.Request) string {
	var extractedIP string
	var source string

	// Try X-Real-IP first (set by Caddy)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		extractedIP = strings.TrimSpace(realIP)
		return extractedIP
	}

	// Try X-Forwarded-For (contains comma-separated list of IPs)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP in the chain (the original client)
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			extractedIP = strings.TrimSpace(parts[0])
			source = "X-Forwarded-For"
			logger.Debug("Client IP extracted from X-Forwarded-For header",
				zap.String("forwarded_ip", extractedIP),
				zap.String("source", source),
				zap.String("full_header", forwardedFor),
				zap.String("raw_remote_addr", r.RemoteAddr))
			return extractedIP
		}
	}

	// Fallback to RemoteAddr (direct connection)
	extractedIP = normalizeIP(r.RemoteAddr)
	source = "RemoteAddr"
	logger.Debug("No proxy headers found, using RemoteAddr",
		zap.String("client_ip", extractedIP),
		zap.String("source", source),
		zap.String("x_real_ip", r.Header.Get("X-Real-IP")),
		zap.String("x_forwarded_for", r.Header.Get("X-Forwarded-For")))

	return extractedIP
}

// normalizeIP converts a network address to a normalized IP string
func normalizeIP(addr string) string {
	// Extract the IP portion (remove port)
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If splitting fails, assume addr is already an IP
		host = addr
	}

	// Normalize IPv4-mapped IPv6 addresses
	ip := net.ParseIP(host)
	if ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
		return ip.String()
	}

	return host
}

// generateClientID generates a unique client ID for event dispatcher
func generateClientID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// cleanExpiredBans periodically removes expired bans from the ban list
func cleanExpiredBans() {
	for {
		time.Sleep(10 * time.Minute)

		banListMutex.Lock()
		now := time.Now()
		var unbanCount int
		for ip, expiry := range clientBanList {
			if now.After(expiry) {
				logger.Debug("Removing expired ban",
					zap.String("client_ip", ip),
					zap.Time("ban_expired", expiry))
				delete(clientBanList, ip)
				unbanCount++
			}
		}
		banListMutex.Unlock()

		if unbanCount > 0 || len(clientBanList) > 0 {
			logger.Debug("Ban list cleanup completed",
				zap.Int("unbanned_count", unbanCount),
				zap.Int("remaining_bans", len(clientBanList)))

			// Log current active bans for debugging
			if len(clientBanList) > 0 {
				for ip, expiry := range clientBanList {
					logger.Debug("Active ban",
						zap.String("client_ip", ip),
						zap.Time("expires", expiry),
						zap.Duration("remaining", time.Until(expiry)))
				}
			}
		}
	}
}

// handleWebSocketConnection handles the upgrade of an HTTP connection to WebSocket
func handleWebSocketConnection(ctx context.Context, w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader, node domain.NodeInterface, relayConfig config.RelayConfig) {
	clientIP := extractRealClientIP(r)

	logger.Debug("New WebSocket connection attempt",
		zap.String("client_ip", clientIP),
		zap.String("user_agent", r.Header.Get("User-Agent")),
		zap.String("origin", r.Header.Get("Origin")))

	// Check if client is banned
	banListMutex.Lock()
	banExpiry, banned := clientBanList[clientIP]
	banListMutex.Unlock()

	if banned && time.Now().Before(banExpiry) {
		// Use new error handling system
		banErr := errors.ClientBannedError("excessive messages", time.Until(banExpiry).String()).
			WithSeverity(errors.SeverityMedium)
		errors.HandleHTTPError(w, r, banErr)
		return
	}

	// Reset exceeded count on new allowed connection
	banListMutex.Lock()
	delete(clientExceededCount, clientIP)
	banListMutex.Unlock()

	// Check global connection limit using metrics counter
	if metrics.GetActiveConnectionsCount() >= int64(relayConfig.ThrottlingConfig.MaxConnections) {
		// Use new error handling system
		limitErr := errors.ConnectionLimitError(
			int(metrics.GetActiveConnectionsCount()), 
			relayConfig.ThrottlingConfig.MaxConnections).
			WithSeverity(errors.SeverityMedium)
		errors.HandleHTTPError(w, r, limitErr)
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
		// Use new error handling system
		upgradeErr := errors.WebSocketError("connection upgrade", err).
			WithSeverity(errors.SeverityMedium)
		errors.HandleWebSocketError(wsConn, "upgrade", upgradeErr)
		return
	}

	// Enable compression
	wsConn.EnableWriteCompression(true)
	_ = wsConn.SetCompressionLevel(2) // nolint:errcheck // compression level is non-critical

	// Update metrics
	metrics.IncrementActiveConnections()
	connectionSuccess = true

	// Create new connection and register it
	conn := NewWsConnection(ctx, wsConn, node, relayConfig, clientIP)
	node.RegisterConn(conn)

	logger.Debug("WebSocket connection established successfully",
		zap.String("client_ip", clientIP),
		zap.Int64("active_connections", metrics.GetActiveConnectionsCount()))

	// Handle messages in a goroutine
	go conn.HandleMessages(ctx, relayConfig)
}

// WsConnection represents a single WebSocket client connection
type WsConnection struct {
	ws           *websocket.Conn
	node         domain.NodeInterface
	realClientIP string // Real client IP (extracted from proxy headers)
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

	// Event dispatcher integration
	clientID    string
	eventChan   chan *nostr.Event
	eventCtx    context.Context
	eventCancel context.CancelFunc
}

// Ensure WsConnection implements domain.WebSocketConnection
var _ domain.WebSocketConnection = (*WsConnection)(nil)

// NewWsConnection initializes a new WebSocket connection
func NewWsConnection(
	ctx context.Context,
	ws *websocket.Conn,
	node domain.NodeInterface,
	cfg config.RelayConfig,
	realClientIP string,
) *WsConnection {
	// Basic rate limiter
	limiter := rate.NewLimiter(
		rate.Limit(cfg.ThrottlingConfig.RateLimit.MaxEventsPerSecond),
		cfg.ThrottlingConfig.RateLimit.BurstSize,
	)

	// Create context for event handling
	eventCtx, eventCancel := context.WithCancel(ctx)

	conn := &WsConnection{
		ws:               ws,
		node:             node,
		realClientIP:     realClientIP,
		idleTimeout:      cfg.IdleTimeout,
		maxLifetime:      24 * time.Hour, // Maximum connection lifetime
		startTime:        time.Now(),
		lastActivity:     time.Now(),
		subscriptions:    make(map[string][]nostr.Filter),
		pingTicker:       time.NewTicker(15 * time.Second),
		limiter:          limiter,
		backpressureChan: make(chan struct{}, 100), // Buffer for backpressure
		// Event dispatcher integration
		clientID:    generateClientID(),
		eventCtx:    eventCtx,
		eventCancel: eventCancel,
	}

	// Register with event dispatcher for real-time notifications
	if eventDispatcher := node.GetEventDispatcher(); eventDispatcher != nil {
		conn.eventChan = eventDispatcher.AddClient(conn.clientID)
		// Start processing events from dispatcher
		go conn.processDispatcherEvents()
	}

	// WebSocket compression
	ws.EnableWriteCompression(true)
	_ = ws.SetCompressionLevel(2) // nolint:errcheck // compression level is non-critical

	// Deadlines + read limit
	_ = ws.SetReadDeadline(time.Now().Add(60 * time.Second)) // nolint:errcheck // deadline is non-critical

	// Set WebSocket read limit based on configured content length with buffer for JSON overhead
	readLimitBytes := int64(cfg.ThrottlingConfig.MaxContentLen * 2) // 2x buffer for JSON overhead
	if readLimitBytes < 1024*1024 {                                 // Minimum 1MB
		readLimitBytes = 1024 * 1024
	}
	if readLimitBytes > 32*1024*1024 { // Maximum 32MB
		readLimitBytes = 32 * 1024 * 1024
	}
	ws.SetReadLimit(readLimitBytes)

	// Ping handler - must echo back the same data
	ws.SetPingHandler(func(appData string) error {
		conn.lastActivity = time.Now()
		conn.writeMu.Lock()
		defer conn.writeMu.Unlock()
		// Echo back the same ping data in the pong response
		_ = conn.ws.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
		return nil
	})

	// Start monitoring
	go conn.monitorConnection(ctx)

	return conn
}

// RemoteAddr returns the client's real remote address (extracted from proxy headers)
func (c *WsConnection) RemoteAddr() string {
	return c.realClientIP
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
		metrics.IncrementErrorCount()
		c.Close()
	}

	// Update metrics
	metrics.IncrementMessagesSent()
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
				zap.String("client", c.RemoteAddr()),
			)
		}
		// Always ensure connection is properly closed and unregistered
		c.closeReason = "message handler terminated"
		c.Close()
		c.node.UnregisterConn(c)
	}()

	clientIP := c.realClientIP

	logger.Debug("Starting message handler",
		zap.String("real_client_ip", clientIP),
		zap.String("websocket_remote_addr", c.ws.RemoteAddr().String()),
		zap.String("client_id", c.clientID))

	// Check if client is banned
	banListMutex.Lock()
	banExpiry, banned := clientBanList[clientIP]
	banListMutex.Unlock()

	if banned && time.Now().Before(banExpiry) {
		logger.Warn("Banned client attempted to send messages",
			zap.String("client_ip", clientIP),
			zap.Time("ban_expires", banExpiry))
		c.closeReason = "client banned"
		c.sendNotice("You are temporarily banned due to excessive messages.")
		c.Close()
		return
	}

	// Set WebSocket read limit based on configured content length with buffer for JSON overhead
	readLimitBytes := int64(cfg.ThrottlingConfig.MaxContentLen * 2) // 2x buffer for JSON overhead
	if readLimitBytes < 1024*1024 {                                 // Minimum 1MB
		readLimitBytes = 1024 * 1024
	}
	if readLimitBytes > 32*1024*1024 { // Maximum 32MB
		readLimitBytes = 32 * 1024 * 1024
	}
	c.ws.SetReadLimit(readLimitBytes)

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

		_ = c.ws.SetReadDeadline(time.Now().Add(60 * time.Second)) // nolint:errcheck // deadline is non-critical
		if time.Since(lastPong) > 90*time.Second {
			logger.Debug("No pong response in 90s, closing connection",
				zap.String("client", c.RemoteAddr()))
			c.closeReason = "no pong response"
			return
		}

		// Read message
		_, rawMsg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.closeReason = "client closed connection"
				logger.Debug("Client closed connection normally",
					zap.String("client", c.RemoteAddr()))
			} else {
				c.closeReason = "read error"
				logger.Debug("WS read error, disconnecting client",
					zap.Error(err),
					zap.String("client", c.RemoteAddr()))
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

				logger.Debug("Client rate limit violation",
					zap.String("client_ip", clientIP),
					zap.Int("violation_count", count),
					zap.Int("ban_threshold", cfg.ThrottlingConfig.BanThreshold),
					zap.String("real_client_ip", c.realClientIP),
					zap.String("websocket_remote_addr", c.ws.RemoteAddr().String()))

				c.sendNotice("Rate limit exceeded: too many messages")

				if count >= cfg.ThrottlingConfig.BanThreshold {
					banDuration := time.Duration(cfg.ThrottlingConfig.BanDuration) * time.Second
					logger.Warn("BANNING CLIENT due to repeated rate limit violations",
						zap.String("client_ip", clientIP),
						zap.Int("violation_count", count),
						zap.Duration("ban_duration", banDuration),
						zap.String("real_client_ip", c.realClientIP),
						zap.Time("ban_expires", time.Now().Add(banDuration)))

					banListMutex.Lock()
					clientBanList[clientIP] = time.Now().Add(banDuration)
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

// processDispatcherEvents handles real-time events from the event dispatcher
func (c *WsConnection) processDispatcherEvents() {
	if c.eventChan == nil {
		return
	}

	for {
		select {
		case <-c.eventCtx.Done():
			return
		case event := <-c.eventChan:
			if event == nil {
				return // Channel closed
			}

			// Check if connection is still active
			if c.isClosed.Load() {
				return
			}

			// Check if any subscription matches this event
			c.subMu.RLock()
			for subID, filters := range c.subscriptions {
				for _, filter := range filters {
					if c.eventMatchesFilter(event, filter) {
						// Send event to client
						c.sendMessage("EVENT", subID, event)
						logger.Debug("Sent real-time event to client",
							zap.String("sub_id", subID),
							zap.String("event_id", event.ID),
							zap.String("client", c.RemoteAddr()))
						break // Only send once per subscription
					}
				}
			}
			c.subMu.RUnlock()
		}
	}
}

// eventMatchesFilter checks if an event matches a subscription filter
func (c *WsConnection) eventMatchesFilter(event *nostr.Event, filter nostr.Filter) bool {
	// Check IDs
	if len(filter.IDs) > 0 {
		found := false
		for _, id := range filter.IDs {
			if event.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check authors
	if len(filter.Authors) > 0 {
		found := false
		for _, author := range filter.Authors {
			if event.PubKey == author {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check kinds
	if len(filter.Kinds) > 0 {
		found := false
		for _, kind := range filter.Kinds {
			if event.Kind == kind {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check since
	if filter.Since != nil && event.CreatedAt < *filter.Since {
		return false
	}

	// Check until
	if filter.Until != nil && event.CreatedAt > *filter.Until {
		return false
	}

	// Check tags
	for tagName, tagValues := range filter.Tags {
		if len(tagValues) > 0 {
			found := false
			for _, tag := range event.Tags {
				if len(tag) >= 2 && tag[0] == tagName {
					for _, value := range tagValues {
						if tag[1] == value {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// Close gracefully shuts down the WebSocket
func (c *WsConnection) Close() {
	c.closeMu.Do(func() {
		c.isClosed.Store(true)

		if c.closeReason != "" {
			logger.Debug("WebSocket connection closed",
				zap.String("reason", c.closeReason),
				zap.String("client_ip", c.RemoteAddr()),
				zap.String("real_client_ip", c.realClientIP),
				zap.Duration("connection_duration", time.Since(c.startTime)))
		}

		// Stop event dispatcher processing
		if c.eventCancel != nil {
			c.eventCancel()
		}

		// Unregister from event dispatcher
		if eventDispatcher := c.node.GetEventDispatcher(); eventDispatcher != nil && c.clientID != "" {
			eventDispatcher.RemoveClient(c.clientID)
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
				zap.String("client", c.RemoteAddr()))
		}

		// Unregister
		c.node.UnregisterConn(c)

		// Finally close
		_ = c.ws.Close()
		logger.Debug("WebSocket connection cleanup completed",
			zap.String("client", c.RemoteAddr()))
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
		case <-c.pingTicker.C:
			// Send ping to keep connection alive
			c.writeMu.Lock()
			if !c.isClosed.Load() {
				_ = c.ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
				err := c.ws.WriteControl(websocket.PingMessage, []byte("keepalive"), time.Now().Add(5*time.Second))
				_ = c.ws.SetWriteDeadline(time.Time{})
				if err != nil {
					logger.Debug("Failed to send ping, closing connection",
						zap.Error(err),
						zap.String("client", c.RemoteAddr()))
					c.writeMu.Unlock()
					c.closeReason = "ping failed"
					c.Close()
					return
				}
				logger.Debug("Sent ping to client", zap.String("client", c.RemoteAddr()))
			}
			c.writeMu.Unlock()
		case <-ticker.C:
			now := time.Now()
			c.writeMu.Lock()

			// Check idle timeout
			if now.Sub(c.lastActivity) > c.idleTimeout {
				c.writeMu.Unlock()
				c.closeReason = "idle timeout"
				c.Close()
				return
			}

			// Check max lifetime
			if now.Sub(c.startTime) > c.maxLifetime {
				c.writeMu.Unlock()
				c.closeReason = "max lifetime exceeded"
				c.Close()
				return
			}

			// Check backpressure
			if len(c.backpressureChan) > 90 { // 90% of buffer capacity
				c.writeMu.Unlock()
				c.closeReason = "backpressure overflow"
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
	metrics.IncrementActiveSubscriptions()
}

// RemoveSubscription removes a subscription
func (c *WsConnection) RemoveSubscription(subID string) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	if _, exists := c.subscriptions[subID]; exists {
		delete(c.subscriptions, subID)
		metrics.DecrementActiveSubscriptions()
	}
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
