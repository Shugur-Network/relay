package relay

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/constants"
	"github.com/Shugur-Network/relay/internal/domain"
	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/Shugur-Network/relay/internal/relay/nips"
	"github.com/Shugur-Network/relay/internal/web"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Server holds references to the relay configuration and node logic.
type Server struct {
	cfg        config.RelayConfig
	fullCfg    *config.Config
	node       domain.NodeInterface
	webHandler *web.Handler
}

// NewServer constructs a new Server with the given RelayConfig and NodeInterface.
func NewServer(relayCfg config.RelayConfig, node domain.NodeInterface, fullCfg *config.Config) *Server {
	webHandler := web.NewHandler(fullCfg, logger.New("web"), node)

	return &Server{
		cfg:        relayCfg,
		fullCfg:    fullCfg,
		node:       node,
		webHandler: webHandler,
	}
}

// ListenAndServe starts your WebSocket relay server and serves NIP-11 on normal HTTP requests.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:    1024 * 1024,
		WriteBufferSize:   1024 * 1024,
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: true,
		HandshakeTimeout:  10 * time.Second,
	}

	// Start background task to clean expired bans
	go cleanExpiredBans()

	// Root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Track request metrics
		metrics.HTTPRequests.Inc()
		start := time.Now()
		defer func() {
			metrics.HTTPRequestDuration.Observe(time.Since(start).Seconds())
		}()

		if isWebSocketRequest(r) {
			// Handle as relay WebSocket connection
			handleWebSocketConnection(ctx, w, r, upgrader, s.node, s.cfg)
		} else {
			// Handle HTTP requests with input validation
			switch {
			case r.URL.Path == "/" && r.Header.Get("Accept") != "application/nostr+json":
				// Serve dashboard for browser requests with validation
				web.SecureValidatedHandlerFunc(s.webHandler.HandleDashboard)(w, r)
			case r.Header.Get("Accept") == "application/nostr+json":
				// Apply security headers for API endpoints
				apiHeaders := web.APISecurityHeaders()
				apiHeaders.Apply(w)
				// Serve NIP-11 metadata for Nostr clients
				metadata := constants.DefaultRelayMetadata(s.fullCfg)
				nips.ServeRelayMetadata(w, metadata)
			case strings.HasPrefix(r.URL.Path, "/static/"):
				// Serve static files with validation
				web.SecureValidatedHandlerFunc(s.webHandler.HandleStatic)(w, r)
			case r.URL.Path == "/api/info":
				// Apply security headers for API endpoints
				apiHeaders := web.APISecurityHeaders()
				apiHeaders.Apply(w)
				// Serve relay info API with validation
				web.ValidatedHandlerFunc(web.APIInputValidation(), func(w http.ResponseWriter, r *http.Request) {
					metadata := constants.DefaultRelayMetadata(s.fullCfg)
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					nips.ServeRelayMetadata(w, metadata)
				})(w, r)
			case r.URL.Path == "/api/stats":
				// Serve relay statistics API with validation
				web.SecureValidatedAPIHandlerFunc(s.webHandler.HandleStatsAPI)(w, r)
			case r.URL.Path == "/api/metrics":
				// Serve real-time metrics API with validation
				web.SecureValidatedAPIHandlerFunc(s.webHandler.HandleMetricsAPI)(w, r)
			case r.URL.Path == "/api/cluster":
				// Serve cluster information API with validation
				web.SecureValidatedAPIHandlerFunc(s.webHandler.HandleClusterAPI)(w, r)
			default:
				// Log invalid requests for security monitoring
				logger.Warn("Invalid request path",
					zap.String("path", r.URL.Path),
					zap.String("client_ip", r.RemoteAddr),
					zap.String("user_agent", r.Header.Get("User-Agent")))
				http.NotFound(w, r)
			}
		}
	})

	httpSrv := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown when context is canceled
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down WebSocket server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	logger.Info("Relay WebSocket server listening", zap.String("address", addr))
	return httpSrv.ListenAndServe()
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(r *http.Request) bool {
	return strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}
