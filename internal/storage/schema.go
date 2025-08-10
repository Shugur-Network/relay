package storage

import (
	"context"
	_ "embed"
	"fmt"

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

	logger.Info("✅ Database schema initialized successfully")
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
