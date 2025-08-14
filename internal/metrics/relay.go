package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Global counters for dashboard display (since prometheus metrics can't be read directly)
var (
	messagesProcessedCount int64
	activeConnectionsCount int64
	messagesSentCount      int64
	activeSubscrCount      int64
	lastEventTimestamp     int64
	lastConnTimestamp      int64
	responseTimeSum        int64
	responseTimeCount      int64
	errorCount             int64
)

// GetMessagesProcessedCount returns the current count of processed messages since start
func GetMessagesProcessedCount() int64 {
	return atomic.LoadInt64(&messagesProcessedCount)
}

// IncrementMessagesProcessed increments both the prometheus counter and our local counter
func IncrementMessagesProcessed() {
	MessagesReceived.Inc()
	atomic.AddInt64(&messagesProcessedCount, 1)
	atomic.StoreInt64(&lastEventTimestamp, time.Now().Unix())
}

// GetActiveConnectionsCount returns the current number of active WebSocket connections
func GetActiveConnectionsCount() int64 {
	return atomic.LoadInt64(&activeConnectionsCount)
}

// IncrementActiveConnections increments both the prometheus gauge and our local counter
func IncrementActiveConnections() {
	ActiveConnections.Inc()
	atomic.AddInt64(&activeConnectionsCount, 1)
	atomic.StoreInt64(&lastConnTimestamp, time.Now().Unix())
}

// DecrementActiveConnections decrements both the prometheus gauge and our local counter
func DecrementActiveConnections() {
	ActiveConnections.Dec()
	atomic.AddInt64(&activeConnectionsCount, -1)
}

// GetMessagesSentCount returns the current count of sent messages
func GetMessagesSentCount() int64 {
	return atomic.LoadInt64(&messagesSentCount)
}

// IncrementMessagesSent increments the sent messages counter
func IncrementMessagesSent() {
	MessagesSent.Inc()
	atomic.AddInt64(&messagesSentCount, 1)
}

// GetActiveSubscriptionsCount returns the current number of active subscriptions
func GetActiveSubscriptionsCount() int64 {
	return atomic.LoadInt64(&activeSubscrCount)
}

// IncrementActiveSubscriptions increments the active subscriptions counter
func IncrementActiveSubscriptions() {
	ActiveSubscriptions.Inc()
	atomic.AddInt64(&activeSubscrCount, 1)
}

// DecrementActiveSubscriptions decrements the active subscriptions counter
func DecrementActiveSubscriptions() {
	ActiveSubscriptions.Dec()
	atomic.AddInt64(&activeSubscrCount, -1)
}

// AddResponseTime adds a response time measurement
func AddResponseTime(responseTimeMs float64) {
	atomic.AddInt64(&responseTimeSum, int64(responseTimeMs))
	atomic.AddInt64(&responseTimeCount, 1)
}

// GetAverageResponseTime returns the average response time in milliseconds
func GetAverageResponseTime() float64 {
	sum := atomic.LoadInt64(&responseTimeSum)
	count := atomic.LoadInt64(&responseTimeCount)
	if count == 0 {
		return 0
	}
	return float64(sum) / float64(count)
}

// IncrementErrorCount increments the error counter
func IncrementErrorCount() {
	atomic.AddInt64(&errorCount, 1)
}

// GetErrorCount returns the current error count
func GetErrorCount() int64 {
	return atomic.LoadInt64(&errorCount)
}

// GetEventsPerSecond calculates events per second over the last minute
func GetEventsPerSecond() float64 {
	lastEvent := atomic.LoadInt64(&lastEventTimestamp)
	if lastEvent == 0 {
		return 0
	}
	
	now := time.Now().Unix()
	timeDiff := now - lastEvent
	if timeDiff == 0 {
		return 0
	}
	
	// Simple approximation - in production you'd want a sliding window
	return float64(atomic.LoadInt64(&messagesProcessedCount)) / float64(timeDiff)
}

// GetConnectionsPerSecond calculates connections per second
func GetConnectionsPerSecond() float64 {
	lastConn := atomic.LoadInt64(&lastConnTimestamp)
	if lastConn == 0 {
		return 0
	}
	
	now := time.Now().Unix()
	timeDiff := now - lastConn
	if timeDiff == 0 {
		return 0
	}
	
	return float64(atomic.LoadInt64(&activeConnectionsCount)) / float64(timeDiff)
}

// GetErrorRate calculates the error rate as a percentage
func GetErrorRate() float64 {
	errors := atomic.LoadInt64(&errorCount)
	messages := atomic.LoadInt64(&messagesProcessedCount)
	if messages == 0 {
		return 0
	}
	return (float64(errors) / float64(messages)) * 100
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
