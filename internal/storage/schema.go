package storage

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

//go:embed schema.sql
var schemaDDL string

// CreateDatabaseIfNotExists creates the specified database if it doesn't exist
func (db *DB) CreateDatabaseIfNotExists(ctx context.Context, dbName string) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	logger.Info("Checking if database exists...", zap.String("database", dbName))

	// Check if database exists
	var exists bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`,
		dbName).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		// Create database
		logger.Info("Creating database...", zap.String("database", dbName))
		_, err = db.Pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		logger.Info("✅ Database created successfully", zap.String("database", dbName))
	} else {
		logger.Info("✅ Database already exists", zap.String("database", dbName))
	}

	return nil
}

// InitializeSchema creates the necessary database and tables if they don't exist
func (db *DB) InitializeSchema(ctx context.Context) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	logger.Info("Initializing database schema...")

	// Note: The database connection should already be to the "shugur" database
	// If we're here, it means the database exists and we're connected to it

	// Execute the schema DDL to create tables
	_, err := db.Pool.Exec(ctx, schemaDDL)
	if err != nil {
		logger.Error("Failed to initialize database schema", zap.Error(err))
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Initialize changefeed for distributed event synchronization
	if err := db.InitializeChangefeed(ctx); err != nil {
		logger.Warn("Failed to initialize changefeed (this is normal for single-node setups)", zap.Error(err))
		// Don't return error as changefeed might not be available in all environments
	}

	logger.Info("✅ Database schema initialized successfully")
	return nil
}

// InitializeChangefeed verifies changefeed capability for distributed event synchronization
func (db *DB) InitializeChangefeed(ctx context.Context) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	logger.Info("Verifying changefeed capability for distributed event synchronization...")

	// Check if changefeeds are supported (CockroachDB specific)
	var hasChangefeedSupport bool
	err := db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'jobs' 
			AND table_schema = 'crdb_internal'
		)
	`).Scan(&hasChangefeedSupport)

	if err != nil || !hasChangefeedSupport {
		return fmt.Errorf("changefeed support not detected (requires CockroachDB)")
	}

	// Ensure rangefeed is enabled for changefeeds to work
	logger.Info("Enabling rangefeed setting for changefeed support...")
	_, err = db.Pool.Exec(ctx, "SET CLUSTER SETTING kv.rangefeed.enabled = true")
	if err != nil {
		logger.Warn("Failed to enable rangefeed setting", zap.Error(err))
		return fmt.Errorf("failed to enable rangefeed setting: %w", err)
	}
	logger.Info("✅ Rangefeed setting enabled successfully")

	// Test changefeed permissions by doing a dry run
	// We don't actually create a persistent changefeed here because:
	// 1. The EventDispatcher creates its own changefeed when needed
	// 2. Multiple persistent changefeeds can cause resource issues
	// 3. Internal changefeeds (used by EventDispatcher) don't need pre-creation
	
	// Test changefeed permissions by checking if the user has CHANGEFEED privileges
	// We'll try to create a temporary changefeed that we immediately cancel
	testChangefeedSQL := "CREATE CHANGEFEED FOR events WITH format='json', envelope='row', updated, initial_scan='no', resolved='10s'"

	// This will fail fast if user doesn't have changefeed permissions
	// or if changefeeds aren't properly configured
	ctx_test, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Try to create a changefeed (it will start running, so we need to close it immediately)
	rows, err := db.Pool.Query(ctx_test, testChangefeedSQL)
	if err != nil {
		logger.Warn("Changefeed test failed", 
			zap.Error(err),
			zap.String("note", "This is expected in single-node or test environments without changefeed support"))
		return fmt.Errorf("changefeed permissions test failed: %w", err)
	}
	// Close immediately to stop the changefeed test
	rows.Close()

	logger.Info("✅ Changefeed capability verified - distributed event synchronization ready")
	return nil
}

// VerifySchema checks if all required tables exist
func (db *DB) VerifySchema(ctx context.Context) error {
	if !db.isConnected() {
		return fmt.Errorf("database is not connected")
	}

	requiredTables := []string{"events"}

	for _, table := range requiredTables {
		var exists bool
		err := db.Pool.QueryRow(ctx,
			`SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("required table %s does not exist", table)
		}

		logger.Debug("✅ Table exists", zap.String("table", table))
	}

	logger.Debug("✅ Database schema verification completed")
	return nil
}
