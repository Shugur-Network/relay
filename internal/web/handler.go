package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/constants"
	"github.com/Shugur-Network/relay/internal/identity"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/Shugur-Network/relay/internal/storage"
	"go.uber.org/zap"
)

// DashboardData represents the data passed to the dashboard template
type DashboardData struct {
	Name          string                       `json:"name"`
	Description   string                       `json:"description"`
	Software      string                       `json:"software"`
	Version       string                       `json:"version"`
	Contact       string                       `json:"contact"`
	Icon          string                       `json:"icon"`
	Host          string                       `json:"host"`
	Pubkey        string                       `json:"pubkey"`
	RelayID       string                       `json:"relay_id"`
	SupportedNIPs []interface{}                `json:"supported_nips"`
	Limitation    *LimitationData              `json:"limitation"`
	Stats         *StatsData                   `json:"stats"`
	Uptime        string                       `json:"uptime"`
	Cluster       *storage.CockroachClusterInfo `json:"cluster"`
}

// LimitationData represents relay limitations
type LimitationData struct {
	MaxMessageLength int  `json:"max_message_length"`
	MaxSubscriptions int  `json:"max_subscriptions"`
	MaxFilters       int  `json:"max_filters"`
	MaxEventTags     int  `json:"max_event_tags"`
	AuthRequired     bool `json:"auth_required"`
	PaymentRequired  bool `json:"payment_required"`
}

// StatsData represents relay statistics
type StatsData struct {
	ActiveConnections int64 `json:"active_connections"`
	MessagesProcessed int64 `json:"messages_processed"`
	EventsStored      int64 `json:"events_stored"`
}

// Handler provides HTTP handlers for the web dashboard
type Handler struct {
	config    *config.Config
	logger    *zap.Logger
	startTime time.Time
	db        interface {
		GetTotalEventCount(ctx context.Context) (int64, error)
		GetCockroachClusterInfo(ctx context.Context) (*storage.CockroachClusterInfo, error)
		GetClusterHealth(ctx context.Context) (map[string]interface{}, error)
	} // Database interface
}

// NewHandler creates a new web handler
func NewHandler(cfg *config.Config, logger *zap.Logger, node interface{}) *Handler {
	h := &Handler{
		config:    cfg,
		logger:    logger,
		startTime: time.Now(),
	}

	// Set database interface if node provides it
	if nodeWithDB, ok := node.(interface {
		DB() *storage.DB
	}); ok {
		h.db = nodeWithDB.DB()
	}

	return h
}

// HandleDashboard serves the main dashboard page
func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// Load template
	tmplPath := filepath.Join("web", "templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		h.logger.Error("Failed to parse dashboard template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare dashboard data
	data := h.getDashboardData(r.Host)

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		h.logger.Error("Failed to execute dashboard template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleStatic serves static files
func (h *Handler) HandleStatic(w http.ResponseWriter, r *http.Request) {
	// Remove "/static/" prefix from URL path
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	fullPath := filepath.Join("web", "static", filePath)

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// HandleStatsAPI serves the stats API endpoint
func (h *Handler) HandleStatsAPI(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current stats
	stats := h.getStatsData()
	uptime := h.formatUptime(time.Since(h.startTime))

	// Create response structure
	response := struct {
		Stats  *StatsData `json:"stats"`
		Uptime string     `json:"uptime"`
	}{
		Stats:  stats,
		Uptime: uptime,
	}

	// Encode and send response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode stats response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// getDashboardData prepares data for the dashboard template
func (h *Handler) getDashboardData(host string) *DashboardData {
	metadata := constants.DefaultRelayMetadata(h.config)

	// Get relay identity for the relay ID
	relayIdentity, err := identity.GetOrCreateRelayIdentity()
	relayID := "unknown"
	if err == nil {
		relayID = relayIdentity.RelayID
	}

	// Clean host (remove port if present)
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Get cluster information
	clusterInfo := h.getClusterData()

	return &DashboardData{
		Name:          metadata.Name,
		Description:   metadata.Description,
		Software:      metadata.Software,
		Version:       metadata.Version,
		Contact:       metadata.Contact,
		Icon:          metadata.Icon,
		Host:          host,
		Pubkey:        metadata.PubKey,
		RelayID:       relayID,
		SupportedNIPs: metadata.SupportedNIPs,
		Limitation: &LimitationData{
			MaxMessageLength: metadata.Limitation.MaxMessageLength,
			MaxSubscriptions: metadata.Limitation.MaxSubscriptions,
			MaxFilters:       metadata.Limitation.MaxFilters,
			MaxEventTags:     metadata.Limitation.MaxEventTags,
			AuthRequired:     metadata.Limitation.AuthRequired,
			PaymentRequired:  metadata.Limitation.PaymentRequired,
		},
		Stats:   h.getStatsData(),
		Uptime:  h.formatUptime(time.Since(h.startTime)),
		Cluster: clusterInfo,
	}
}

// getStatsData retrieves current statistics
func (h *Handler) getStatsData() *StatsData {
	var eventsStored int64

	// Get events stored from database if available
	if h.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		count, err := h.db.GetTotalEventCount(ctx)
		if err != nil {
			h.logger.Warn("Failed to get total event count", zap.Error(err))
			eventsStored = 0
		} else {
			eventsStored = count
		}

		// Update the metrics gauge with current count
		metrics.EventsStored.Set(float64(eventsStored))
	}

	// Get other metrics - using our tracking functions
	stats := &StatsData{
		ActiveConnections: metrics.GetActiveConnectionsCount(),
		MessagesProcessed: metrics.GetMessagesProcessedCount(),
		EventsStored:      eventsStored,
	}

	return stats
}

// getClusterData retrieves CockroachDB cluster information
func (h *Handler) getClusterData() *storage.CockroachClusterInfo {
	if h.db == nil {
		return &storage.CockroachClusterInfo{
			IsCluster: false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clusterInfo, err := h.db.GetCockroachClusterInfo(ctx)
	if err != nil {
		h.logger.Warn("Failed to get cluster information", zap.Error(err))
		return &storage.CockroachClusterInfo{
			IsCluster: false,
		}
	}

	return clusterInfo
}

// HandleClusterAPI serves the cluster API endpoint
func (h *Handler) HandleClusterAPI(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.db == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if requesting health or full cluster info
	requestType := r.URL.Query().Get("type")
	
	if requestType == "health" {
		health, err := h.db.GetClusterHealth(ctx)
		if err != nil {
			h.logger.Error("Failed to get cluster health", zap.Error(err))
			http.Error(w, "Failed to get cluster health", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(health); err != nil {
			h.logger.Error("Failed to encode cluster health response", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		clusterInfo, err := h.db.GetCockroachClusterInfo(ctx)
		if err != nil {
			h.logger.Error("Failed to get cluster information", zap.Error(err))
			http.Error(w, "Failed to get cluster information", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(clusterInfo); err != nil {
			h.logger.Error("Failed to encode cluster info response", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// formatUptime formats duration as a human-readable string
func (h *Handler) formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}
