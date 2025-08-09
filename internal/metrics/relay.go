package metrics

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Global counters for dashboard display (since prometheus metrics can't be read directly)
var (
	messagesProcessedCount int64
	activeConnectionsCount int64
)

// GetMessagesProcessedCount returns the current count of processed messages since start
func GetMessagesProcessedCount() int64 {
	return atomic.LoadInt64(&messagesProcessedCount)
}

// IncrementMessagesProcessed increments both the prometheus counter and our local counter
func IncrementMessagesProcessed() {
	MessagesReceived.Inc()
	atomic.AddInt64(&messagesProcessedCount, 1)
}

// GetActiveConnectionsCount returns the current number of active WebSocket connections
func GetActiveConnectionsCount() int64 {
	return atomic.LoadInt64(&activeConnectionsCount)
}

// IncrementActiveConnections increments both the prometheus gauge and our local counter
func IncrementActiveConnections() {
	ActiveConnections.Inc()
	atomic.AddInt64(&activeConnectionsCount, 1)
}

// DecrementActiveConnections decrements both the prometheus gauge and our local counter
func DecrementActiveConnections() {
	ActiveConnections.Dec()
	atomic.AddInt64(&activeConnectionsCount, -1)
}

// SyncActiveConnectionsCount synchronizes the internal counter with the actual count
// This helps prevent drift between the metrics counter and reality
func SyncActiveConnectionsCount(actualCount int64) {
	currentCount := atomic.LoadInt64(&activeConnectionsCount)
	if currentCount != actualCount {
		// Update our internal counter to match reality
		atomic.StoreInt64(&activeConnectionsCount, actualCount)

		// Update prometheus gauge as well
		ActiveConnections.Set(float64(actualCount))
	}
} // Metrics for tracking relay performance and usage
var (
	// Connection metrics
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nostr_relay_active_connections",
		Help: "The number of active WebSocket connections",
	})

	ActiveSubscriptions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nostr_relay_active_subscriptions",
		Help: "The number of active subscriptions",
	})

	// Message metrics
	MessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nostr_relay_messages_received_total",
		Help: "The total number of messages received",
	})

	MessagesSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nostr_relay_messages_sent_total",
		Help: "The total number of messages sent",
	})

	MessageSizeBytes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "nostr_relay_message_size_bytes",
		Help:    "Size of received messages in bytes",
		Buckets: prometheus.ExponentialBuckets(10, 10, 6), // 10, 100, 1000, ..., 1000000
	})

	MessageSizeBytesSent = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "nostr_relay_message_size_bytes_sent",
		Help:    "Size of sent messages in bytes",
		Buckets: prometheus.ExponentialBuckets(10, 10, 6), // 10, 100, 1000, ..., 1000000
	})

	// Command metrics
	CommandsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nostr_relay_commands_received_total",
		Help: "The total number of commands received by type",
	}, []string{"type"}) // "EVENT", "REQ", "CLOSE", etc.

	CommandProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "nostr_relay_command_processing_duration_seconds",
		Help:    "Time to process different command types",
		Buckets: prometheus.ExponentialBuckets(0.001, 10, 5), // 0.001, 0.01, 0.1, 1, 10
	}, []string{"type"})

	// Event metrics
	EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nostr_relay_events_processed_total",
		Help: "The total number of events processed by kind",
	}, []string{"kind"})

	EventsStored = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nostr_relay_events_stored",
		Help: "The total number of events currently stored in the database",
	})

	DuplicateEvents = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nostr_relay_duplicate_events_total",
		Help: "The total number of duplicate events received",
	})

	// HTTP metrics
	HTTPRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nostr_relay_http_requests_total",
		Help: "The total number of HTTP requests",
	})

	HTTPRequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "nostr_relay_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.ExponentialBuckets(0.01, 10, 5), // 0.01, 0.1, 1, 10, 100
	})

	// Error metrics
	ErrorsCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nostr_relay_errors_total",
		Help: "The total number of errors by type",
	}, []string{"type"}) // "validation", "database", "websocket", etc.

	// Database metrics
	DBConnections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nostr_relay_db_connections_total",
		Help: "Total number of database connections by status",
	}, []string{"status"}) // "success", "failure", "closed"

	DBErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nostr_relay_db_errors_total",
		Help: "Total number of database errors by type",
	}, []string{"error_type"})

	DBOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nostr_relay_db_operations_total",
		Help: "Total number of database operations by type",
	}, []string{"operation"})
)

// RegisterMetrics ensures all metrics are registered with Prometheus
func RegisterMetrics() {
	// Pre-register common command types
	commandTypes := []string{"EVENT", "REQ", "CLOSE", "COUNT"}
	for _, cmdType := range commandTypes {
		CommandsReceived.WithLabelValues(cmdType)
		CommandProcessingDuration.WithLabelValues(cmdType)
	}

	// Pre-register common event kinds
	eventKinds := []string{"0", "1", "2", "3", "4", "5", "7", "40", "41", "1059"}
	for _, kind := range eventKinds {
		EventsProcessed.WithLabelValues(kind)
	}

	// Pre-register error types
	errorTypes := []string{
		"validation", "database", "websocket", "rate_limit",
		"max_connections", "auth", "timeout",
	}
	for _, errType := range errorTypes {
		ErrorsCount.WithLabelValues(errType)
	}

	// Pre-register DB connection statuses
	dbStatuses := []string{"success", "failure", "closed"}
	for _, status := range dbStatuses {
		DBConnections.WithLabelValues(status)
	}

	// Pre-register DB error types
	dbErrorTypes := []string{
		"connection_failed", "transaction_start_failed", "batch_execution_failed",
		"transaction_commit_failed", "command_execution_failed", "bloom_filter_fetch_failed",
		"bloom_filter_scan_failed",
	}
	for _, errType := range dbErrorTypes {
		DBErrors.WithLabelValues(errType)
	}

	// Pre-register DB operations
	dbOps := []string{"batch_success", "bloom_filter_rebuild_success"}
	for _, op := range dbOps {
		DBOperations.WithLabelValues(op)
	}
}
