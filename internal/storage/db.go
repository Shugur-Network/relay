package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Shugur-Network/Relay/internal/logger"
	"github.com/Shugur-Network/Relay/internal/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/willf/bloom"

	"go.uber.org/zap"
)

// DBState represents the current state of the database connection
type DBState int

const (
	DBStateInitial DBState = iota
	DBStateConnecting
	DBStateConnected
	DBStateDisconnecting
	DBStateClosed
)

// DB represents the CockroachDB connection
type DB struct {
	Pool         *pgxpool.Pool
	Bloom        *bloom.BloomFilter
	state        DBState
	stateMu      sync.RWMutex
	errors       chan error
	errorCount   int32
	errorCountMu sync.RWMutex
}

// InitDB initializes the CockroachDB connection with retries
func InitDB(ctx context.Context, dbURI string) (*DB, error) {
	var pool *pgxpool.Pool
	var err error
	backoff := 2 * time.Second
	attempts := 0

	db := &DB{
		state:  DBStateConnecting,
		errors: make(chan error, 100),
	}

	for i := 0; i < 5; i++ { // Retry up to 5 times
		attempts++
		pool, err = pgxpool.New(ctx, dbURI)
		if err == nil {
			// Test the actual connection
			if err = pool.Ping(ctx); err == nil {
				db.Pool = pool
				db.Bloom = bloom.NewWithEstimates(10_000_000, 0.01) // 10M entries with 1% false positive rate
				db.state = DBStateConnected

				logger.Info("✅ DB Connected Successfully",
					zap.Int("attempts", attempts))
				metrics.DBConnections.WithLabelValues("success").Inc()
				return db, nil
			}
			// Connection pool created but ping failed, close it
			pool.Close()
		}

		logger.Warn("Failed to connect to DB, retrying...",
			zap.Error(err),
			zap.Int("attempt", attempts),
			zap.Duration("backoff", backoff))
		metrics.DBConnections.WithLabelValues("failure").Inc()
		time.Sleep(backoff)
		backoff *= 2 // Exponential backoff (2s, 4s, 8s...)
	}

	db.state = DBStateClosed
	metrics.DBErrors.WithLabelValues("connection_failed").Inc()
	return nil, fmt.Errorf("failed to connect to DB after %d attempts: %w", attempts, err)
}

// CloseDB closes the database connection
func (db *DB) CloseDB() error {
	db.stateMu.Lock()
	if db.state == DBStateDisconnecting || db.state == DBStateClosed {
		db.stateMu.Unlock()
		return nil
	}
	db.state = DBStateDisconnecting
	db.stateMu.Unlock()

	if db.Pool != nil {
		db.Pool.Close()
		db.state = DBStateClosed
		logger.Debug("Database connection closed")
		metrics.DBConnections.WithLabelValues("closed").Inc()
		return nil
	}

	return fmt.Errorf("database pool is nil")
}

// ExecuteQuery handles single-row queries (SELECT)
func (db *DB) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (pgx.Row, error) {
	if !db.isConnected() {
		return nil, fmt.Errorf("database is not connected")
	}

	logger.Debug("Executing query",
		zap.String("query", query),
		zap.Any("args", args))

	row := db.Pool.QueryRow(ctx, query, args...)
	return row, nil
}

// ExecuteBatch handles batch inserts or updates
func (db *DB) ExecuteBatch(ctx context.Context, batch *pgx.Batch) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		db.recordError(fmt.Errorf("failed to start transaction: %w", err))
		metrics.DBErrors.WithLabelValues("transaction_start_failed").Inc()
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			// Only log rollback errors if the transaction hasn't been committed
			db.recordError(fmt.Errorf("rollback failed: %w", rollbackErr))
		}
	}()

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		db.recordError(fmt.Errorf("batch execution failed: %w", err))
		metrics.DBErrors.WithLabelValues("batch_execution_failed").Inc()
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		db.recordError(fmt.Errorf("transaction commit failed: %w", err))
		metrics.DBErrors.WithLabelValues("transaction_commit_failed").Inc()
		return err
	}

	logger.Debug("Batch operation completed")
	metrics.DBOperations.WithLabelValues("batch_success").Inc()
	return nil
}

// ExecuteCommand handles INSERT, UPDATE, DELETE commands
func (db *DB) ExecuteCommand(ctx context.Context, query string, args ...interface{}) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	logger.Debug("Executing command",
		zap.String("query", query),
		zap.Any("args", args))

	_, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		db.recordError(fmt.Errorf("command execution failed: %w", err))
		logger.Error("Command execution failed",
			zap.Error(err),
			zap.String("query", query))
		metrics.DBErrors.WithLabelValues("command_execution_failed").Inc()
	}
	return err
}

// RebuildBloomFilter fetches all event IDs from CockroachDB and updates the Bloom filter.
func (db *DB) RebuildBloomFilter(ctx context.Context) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	logger.Info("Rebuilding Bloom filter from database...")

	query := `SELECT id FROM events`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		db.recordError(fmt.Errorf("failed to fetch event IDs: %w", err))
		metrics.DBErrors.WithLabelValues("bloom_filter_fetch_failed").Inc()
		return err
	}
	defer rows.Close()

	count := 0
	db.Bloom.ClearAll()

	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			db.recordError(fmt.Errorf("failed to scan event ID: %w", err))
			logger.Debug("Failed to scan event ID",
				zap.Error(err))
			continue
		}

		db.Bloom.AddString(eventID)
		count++

		if count%100000 == 0 {
			logger.Debug("Bloom filter progress",
				zap.Int("events", count))
		}
	}

	if err := rows.Err(); err != nil {
		db.recordError(fmt.Errorf("error scanning rows: %w", err))
		metrics.DBErrors.WithLabelValues("bloom_filter_scan_failed").Inc()
		return err
	}

	logger.Info("Bloom filter rebuilt successfully",
		zap.Int("total_events", count))
	metrics.DBOperations.WithLabelValues("bloom_filter_rebuild_success").Inc()
	return nil
}

// isConnected checks if the database is in a connected state
func (db *DB) isConnected() bool {
	db.stateMu.RLock()
	defer db.stateMu.RUnlock()
	return db.state == DBStateConnected
}

// recordError records an error in the database service
func (db *DB) recordError(err error) {
	db.errorCountMu.Lock()
	db.errorCount++
	count := db.errorCount
	db.errorCountMu.Unlock()

	select {
	case db.errors <- err:
	default:
		// Channel is full, log directly
		logger.Error("Database error (channel full)",
			zap.Error(err),
			zap.Int32("error_count", count))
	}
}

// Add this helper function to your DB struct
func (db *DB) executeWithRetry(ctx context.Context, f func(context.Context) error) error {
	retries := 3
	var lastErr error

	for i := 0; i < retries; i++ {
		err := f(ctx)
		if err == nil {
			return nil
		}

		// Check if error is a timeout or deadlock (retryable)
		if strings.Contains(err.Error(), "statement timeout") ||
			strings.Contains(err.Error(), "deadlock") {
			lastErr = err
			// Exponential backoff
			time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
			continue
		}

		// Not a retryable error
		return err
	}

	return fmt.Errorf("operation failed after %d retries: %w", retries, lastErr)
}
