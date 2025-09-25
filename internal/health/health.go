package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/constants"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentStatus represents the status of a specific component
type ComponentStatus struct {
	Name    string                 `json:"name"`
	Status  HealthStatus           `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthResponse represents the complete health check response
type HealthResponse struct {
	Status     HealthStatus       `json:"status"`
	Timestamp  time.Time          `json:"timestamp"`
	Version    string             `json:"version"`
	Uptime     string             `json:"uptime"`
	Components []*ComponentStatus `json:"components"`
	Summary    map[string]interface{} `json:"summary"`
}

// DatabaseInterface defines the database operations needed for health checks
type DatabaseInterface interface {
	Ping() error
	Stats() DatabaseStats
	GetClusterHealth(ctx context.Context) (map[string]interface{}, error)
}

// NodeInterface defines the node operations needed for health checks
type NodeInterface interface {
	GetConnectionCount() int
	GetStartTime() time.Time
}

// DatabaseStats represents database connection pool statistics (matches storage.DatabaseStats)
type DatabaseStats struct {
	OpenConnections    int
	InUse             int  
	Idle              int
	MaxOpenConnections int
	MaxIdleConnections int
}

// HealthChecker performs comprehensive health checks
type HealthChecker struct {
	db       DatabaseInterface
	node     NodeInterface
	cfg      *config.Config
	logger   *zap.Logger
	startTime time.Time
	version   string
	mu       sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db DatabaseInterface, node NodeInterface, cfg *config.Config, logger *zap.Logger, version string) *HealthChecker {
	return &HealthChecker{
		db:        db,
		node:      node,
		cfg:       cfg,
		logger:    logger.Named("health"),
		startTime: time.Now(),
		version:   version,
	}
}

// CheckHealth performs a comprehensive health check
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	startTime := time.Now()
	components := make([]*ComponentStatus, 0)
	
	// Check database health
	dbStatus := h.checkDatabase(ctx)
	components = append(components, dbStatus)

	// Check memory health
	memStatus := h.checkMemory()
	components = append(components, memStatus)

	// Check connections health
	connStatus := h.checkConnections()
	components = append(components, connStatus)

	// Check system resources
	systemStatus := h.checkSystemResources()
	components = append(components, systemStatus)

	// Determine overall status
	overallStatus := h.determineOverallStatus(components)

	// Calculate uptime
	uptime := time.Since(h.startTime)

	response := &HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    h.version,
		Uptime:     h.formatUptime(uptime),
		Components: components,
		Summary: map[string]interface{}{
			"total_components":     len(components),
			"healthy_components":   h.countComponentsByStatus(components, StatusHealthy),
			"degraded_components":  h.countComponentsByStatus(components, StatusDegraded),
			"unhealthy_components": h.countComponentsByStatus(components, StatusUnhealthy),
			"check_duration_ms":    time.Since(startTime).Milliseconds(),
		},
	}

	return response
}

// checkDatabase checks database connectivity and performance
func (h *HealthChecker) checkDatabase(ctx context.Context) *ComponentStatus {
	status := &ComponentStatus{
		Name:    "database",
		Details: make(map[string]interface{}),
	}

	// Check basic connectivity
	if err := h.db.Ping(); err != nil {
		status.Status = StatusUnhealthy
		status.Message = "Database connection failed"
		status.Details["error"] = err.Error()
		return status
	}

	// Get connection pool stats
	stats := h.db.Stats()
	status.Details["open_connections"] = stats.OpenConnections
	status.Details["in_use"] = stats.InUse
	status.Details["idle"] = stats.Idle
	status.Details["max_open_connections"] = stats.MaxOpenConnections
	status.Details["max_idle_connections"] = stats.MaxIdleConnections

	// Check connection pool health
	connectionUtilization := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
	status.Details["connection_utilization_percent"] = connectionUtilization

	// Get cluster health if available
	if clusterHealth, err := h.db.GetClusterHealth(ctx); err == nil {
		status.Details["cluster"] = clusterHealth
	}

	// Determine database status
	if connectionUtilization > 90 {
		status.Status = StatusDegraded
		status.Message = "High database connection utilization"
	} else if connectionUtilization > 95 {
		status.Status = StatusUnhealthy
		status.Message = "Critical database connection utilization"
	} else {
		status.Status = StatusHealthy
		status.Message = "Database is healthy"
	}

	return status
}

// checkMemory checks memory usage
func (h *HealthChecker) checkMemory() *ComponentStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := &ComponentStatus{
		Name:    "memory",
		Details: make(map[string]interface{}),
	}

	// Convert to MB for readability
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024
	heapMB := float64(m.HeapAlloc) / 1024 / 1024

	status.Details["alloc_mb"] = allocMB
	status.Details["sys_mb"] = sysMB
	status.Details["heap_mb"] = heapMB
	status.Details["num_gc"] = m.NumGC
	status.Details["gc_cpu_fraction"] = m.GCCPUFraction

	// Memory thresholds (these should be configurable)
	const (
		memoryWarningMB  = 500  // 500MB
		memoryCriticalMB = 1000 // 1GB
	)

	if allocMB > memoryCriticalMB {
		status.Status = StatusUnhealthy
		status.Message = fmt.Sprintf("High memory usage: %.1f MB", allocMB)
	} else if allocMB > memoryWarningMB {
		status.Status = StatusDegraded
		status.Message = fmt.Sprintf("Elevated memory usage: %.1f MB", allocMB)
	} else {
		status.Status = StatusHealthy
		status.Message = fmt.Sprintf("Memory usage normal: %.1f MB", allocMB)
	}

	return status
}

// checkConnections checks WebSocket connection health
func (h *HealthChecker) checkConnections() *ComponentStatus {
	status := &ComponentStatus{
		Name:    "connections",
		Details: make(map[string]interface{}),
	}

	connectionCount := h.node.GetConnectionCount()
	status.Details["active_connections"] = connectionCount
	
	// Get connection limits from config
	maxConnections := h.cfg.Relay.ThrottlingConfig.MaxConnections
	if maxConnections == 0 {
		maxConnections = 1000 // Default fallback
	}

	connectionUtilization := float64(connectionCount) / float64(maxConnections) * 100
	status.Details["max_connections"] = maxConnections
	status.Details["connection_utilization_percent"] = connectionUtilization

	// Determine connection status
	if connectionUtilization > 90 {
		status.Status = StatusDegraded
		status.Message = fmt.Sprintf("High connection utilization: %d/%d (%.1f%%)", 
			connectionCount, maxConnections, connectionUtilization)
	} else if connectionUtilization > 95 {
		status.Status = StatusUnhealthy
		status.Message = fmt.Sprintf("Critical connection utilization: %d/%d (%.1f%%)", 
			connectionCount, maxConnections, connectionUtilization)
	} else {
		status.Status = StatusHealthy
		status.Message = fmt.Sprintf("Connection count normal: %d/%d (%.1f%%)", 
			connectionCount, maxConnections, connectionUtilization)
	}

	return status
}

// checkSystemResources checks system-level resources
func (h *HealthChecker) checkSystemResources() *ComponentStatus {
	status := &ComponentStatus{
		Name:    "system",
		Details: make(map[string]interface{}),
	}

	status.Details["goroutines"] = runtime.NumGoroutine()
	status.Details["cpus"] = runtime.NumCPU()
	
	goroutineCount := runtime.NumGoroutine()
	
	// Goroutine thresholds
	const (
		goroutineWarning  = 1000
		goroutineCritical = 5000
	)

	if goroutineCount > goroutineCritical {
		status.Status = StatusUnhealthy
		status.Message = fmt.Sprintf("High goroutine count: %d", goroutineCount)
	} else if goroutineCount > goroutineWarning {
		status.Status = StatusDegraded  
		status.Message = fmt.Sprintf("Elevated goroutine count: %d", goroutineCount)
	} else {
		status.Status = StatusHealthy
		status.Message = fmt.Sprintf("System resources normal: %d goroutines", goroutineCount)
	}

	return status
}

// determineOverallStatus determines the overall health status from components
func (h *HealthChecker) determineOverallStatus(components []*ComponentStatus) HealthStatus {
	unhealthyCount := 0
	degradedCount := 0

	for _, comp := range components {
		switch comp.Status {
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	// If any component is unhealthy, overall status is unhealthy
	if unhealthyCount > 0 {
		return StatusUnhealthy
	}

	// If any component is degraded, overall status is degraded
	if degradedCount > 0 {
		return StatusDegraded
	}

	return StatusHealthy
}

// countComponentsByStatus counts components with a specific status
func (h *HealthChecker) countComponentsByStatus(components []*ComponentStatus, status HealthStatus) int {
	count := 0
	for _, comp := range components {
		if comp.Status == status {
			count++
		}
	}
	return count
}

// formatUptime formats uptime duration as a human-readable string
func (h *HealthChecker) formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// HandleHealth is the HTTP handler for health checks
func (h *HealthChecker) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), constants.HealthCheckTimeout*time.Second)
	defer cancel()

	// Check for ready parameter for readiness probes
	ready := r.URL.Query().Get("ready")
	
	healthResponse := h.CheckHealth(ctx)

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if ready == "1" {
		// For readiness probes, return 200 only if healthy
		switch healthResponse.Status {
		case StatusHealthy:
			statusCode = http.StatusOK
		case StatusDegraded:
			statusCode = http.StatusOK // Still ready, just degraded
		case StatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		}
	} else {
		// For liveness probes, return 200 unless completely unhealthy
		switch healthResponse.Status {
		case StatusHealthy, StatusDegraded:
			statusCode = http.StatusOK
		case StatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		h.logger.Error("Failed to encode health response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log health check results for monitoring
	h.logger.Debug("Health check completed",
		zap.String("status", string(healthResponse.Status)),
		zap.Int("status_code", statusCode),
		zap.String("client_ip", r.RemoteAddr),
		zap.Int64("duration_ms", healthResponse.Summary["check_duration_ms"].(int64)))
}