package application

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/constants"
	"github.com/Shugur-Network/relay/internal/domain"
	"github.com/Shugur-Network/relay/internal/limiter"
	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/Shugur-Network/relay/internal/relay"
	"github.com/Shugur-Network/relay/internal/storage"
	"github.com/Shugur-Network/relay/internal/workers"
	"go.uber.org/zap"
)

// NodeBuilder is used to incrementally construct a Node instance.
type NodeBuilder struct {
	ctx     context.Context
	cancel  context.CancelFunc
	config  *config.Config
	privKey ed25519.PrivateKey

	database        *storage.DB
	eventDispatcher *storage.EventDispatcher
	workerPool      *workers.WorkerPool
	validator       domain.EventValidator
	eventVal        *relay.EventValidator
	eventProc       *storage.EventProcessor
	rateLimiter     *limiter.RateLimiter

	blacklist map[string]struct{}
	whitelist map[string]struct{}
}

// NewNodeBuilder creates a new NodeBuilder with its own cancelable context.
func NewNodeBuilder(ctx context.Context, cfg *config.Config, privKey ed25519.PrivateKey) *NodeBuilder {
	c, cancel := context.WithCancel(ctx)
	return &NodeBuilder{
		ctx:     c,
		cancel:  cancel,
		config:  cfg,
		privKey: privKey,
	}
}

// BuildDB initializes the database connection with support for both standalone and distributed modes.
// func (b *NodeBuilder) BuildDB() error {
// 	// Check if certs directory exists to determine if we should use secure connection
// 	certsExist := false
// 	if _, err := os.Stat("./certs"); err == nil {
// 		if _, err := os.Stat("./certs/ca.crt"); err == nil {
// 			certsExist = true
// 		}
// 	}

// 	var defaultDbURI, targetDbURI string

// 	// Determine connection type based on certificate availability
// 	if certsExist {
// 		// Distributed mode with certificates - use secure connection
// 		logger.Info("Building distributed database connection (secure mode)",
// 			zap.String("server", b.config.Database.Server),
// 			zap.Int("port", b.config.Database.Port),
// 			zap.Bool("certs_found", true))

// 		// Connect to default database first to create shugur database if needed
// 		defaultDbURI = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=require&sslrootcert=%s&sslcert=%s&sslkey=%s",
// 			"root",
// 			b.config.Database.Server,
// 			b.config.Database.Port,
// 			"defaultdb",
// 			"./certs/ca.crt",
// 			"./certs/client.root.crt",
// 			"./certs/client.root.key",
// 		)

// 		targetDbURI = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=require&sslrootcert=%s&sslcert=%s&sslkey=%s",
// 			"root",
// 			b.config.Database.Server,
// 			b.config.Database.Port,
// 			constants.DatabaseName,
// 			"./certs/ca.crt",
// 			"./certs/client.root.crt",
// 			"./certs/client.root.key",
// 		)
// 	} else {
// 		// Standalone/Development mode without certificates - use insecure connection
// 		logger.Info("Building database connection (insecure mode)",
// 			zap.String("server", b.config.Database.Server),
// 			zap.Int("port", b.config.Database.Port),
// 			zap.Bool("certs_found", false))

// 		// Connect to default database first to create shugur database if needed
// 		defaultDbURI = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
// 			"root",
// 			b.config.Database.Server,
// 			b.config.Database.Port,
// 			"defaultdb",
// 		)

// 		targetDbURI = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
// 			"root",
// 			b.config.Database.Server,
// 			b.config.Database.Port,
// 			constants.DatabaseName,
// 		)
// 	}

// 	// First, connect to default database to ensure shugur database exists
// 	logger.Info("Connecting to default database to check/create shugur database...")
// 	defaultConn, err := storage.InitDB(b.ctx, defaultDbURI)
// 	if err != nil {
// 		b.cancel()
// 		return fmt.Errorf("failed to connect to default database: %w", err)
// 	}

// 	// Create shugur database if it doesn't exist
// 	if err := defaultConn.CreateDatabaseIfNotExists(b.ctx, constants.DatabaseName); err != nil {
// 		if closeErr := defaultConn.CloseDB(); closeErr != nil {
// 			logger.Warn("Failed to close default database connection", zap.Error(closeErr))
// 		}
// 		b.cancel()
// 		return fmt.Errorf("failed to create %s database: %w", constants.DatabaseName, err)
// 	}

// 	// Close connection to default database
// 	if err := defaultConn.CloseDB(); err != nil {
// 		logger.Warn("Failed to close default database connection", zap.Error(err))
// 	}

// 	// Now connect to the shugur database
// 	logger.Info("Connecting to shugur database...")
// 	dbConn, err := storage.InitDB(b.ctx, targetDbURI)
// 	if err != nil {
// 		b.cancel()
// 		return fmt.Errorf("failed to initialize database connection to %s: %w", constants.DatabaseName, err)
// 	}

// 	b.database = dbConn

// 	// Initialize database schema on first run
// 	if err := dbConn.InitializeSchema(b.ctx); err != nil {
// 		logger.Error("Failed to initialize database schema", zap.Error(err))
// 		return fmt.Errorf("failed to initialize database schema: %w", err)
// 	}

// 	// Verify schema exists
// 	if err := dbConn.VerifySchema(b.ctx); err != nil {
// 		logger.Error("Database schema verification failed", zap.Error(err))
// 		return fmt.Errorf("database schema verification failed: %w", err)
// 	}

// 	// Initialize EventsStored metric with current count
// 	if count, err := dbConn.GetTotalEventCount(b.ctx); err != nil {
// 		logger.Warn("Failed to get initial event count for metrics", zap.Error(err))
// 	} else {
// 		metrics.EventsStored.Set(float64(count))
// 		logger.Info("Initialized EventsStored metric", zap.Int64("count", count))
// 	}

// 	if err := b.database.RebuildBloomFilter(b.ctx); err != nil {
// 		logger.Warn("Failed to rebuild bloom filter", zap.Error(err))
// 	}

// 	return nil
// }

// small helpers
func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
func allExist(paths ...string) bool {
	for _, p := range paths {
		if !fileExists(p) {
			return false
		}
	}
	return true
}

// BuildDB initializes the database connection with support for both standalone and distributed modes.
func (b *NodeBuilder) BuildDB() error {
	const (
		caPath    = "./certs/ca.crt"
		relayCert = "./certs/client.relay.crt"
		relayKey  = "./certs/client.relay.key"
		rootCert  = "./certs/client.root.crt"
		rootKey   = "./certs/client.root.key"
		defaultDB = "defaultdb"
	)

	host := b.config.Database.Server
	port := b.config.Database.Port
	dbName := constants.DatabaseName

	hasCA := fileExists(caPath)
	hasRelay := allExist(relayCert, relayKey)
	hasRoot := allExist(rootCert, rootKey)
	secure := hasCA && (hasRelay || hasRoot)

	var defaultDbURI, targetDbURI, targetUser string

	if secure {
		// Use root user for all cases to avoid permission issues with distributed cluster metadata access
		targetUser = "root"

		if !hasRoot {
			logger.Error("Root client certs required but not found")
			return fmt.Errorf("root client certificates not found at %s or %s", rootCert, rootKey)
		}

		logger.Info("Building distributed database connection (secure mode, verify-full)",
			zap.String("server", host),
			zap.Int("port", port),
			zap.Bool("relay_client_certs", hasRelay),
			zap.Bool("root_client_certs", hasRoot))

		// Only attempt DB creation if root client certs are available
		if hasRoot {
			defaultDbURI = fmt.Sprintf(
				"postgres://%s@%s:%d/%s?sslmode=verify-full&sslrootcert=%s&sslcert=%s&sslkey=%s",
				"root", host, port, defaultDB, caPath, rootCert, rootKey,
			)
		} else {
			logger.Info("Root client certs not present; skipping default DB provisioning step (expecting external provisioning).")
		}

		targetDbURI = fmt.Sprintf(
			"postgres://%s@%s:%d/%s?sslmode=verify-full&sslrootcert=%s&sslcert=%s&sslkey=%s",
			targetUser, host, port, dbName, caPath, rootCert, rootKey,
		)

	} else {
		// Insecure/dev mode
		targetUser = "root"
		logger.Info("Building database connection (insecure mode)",
			zap.String("server", host),
			zap.Int("port", port),
			zap.Bool("certs_found", false))

		defaultDbURI = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable", "root", host, port, defaultDB)
		targetDbURI = fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable", "root", host, port, dbName)
	}

	// Optionally connect to default DB to create the target DB (only when defaultDbURI is set).
	if defaultDbURI != "" {
		logger.Info("Connecting to default database to check/create target database...", zap.String("as_user", "root"))
		defaultConn, err := storage.InitDB(b.ctx, defaultDbURI)
		if err != nil {
			// Don’t hard fail; installer may have already provisioned the DB
			logger.Warn("Root connection to default database failed; skipping create step (assuming provisioned).", zap.Error(err))
		} else {
			if err := defaultConn.CreateDatabaseIfNotExists(b.ctx, dbName); err != nil {
				logger.Warn("CreateDatabaseIfNotExists failed; continuing (database may already exist or insufficient privileges).", zap.Error(err))
			}
			if err := defaultConn.CloseDB(); err != nil {
				logger.Warn("Failed to close default database connection", zap.Error(err))
			}
		}
	}

	// Connect to the target database
	logger.Info("Connecting to target database...",
		zap.String("db", dbName),
		zap.String("as_user", targetUser))
	dbConn, err := storage.InitDB(b.ctx, targetDbURI)
	if err != nil {
		b.cancel()
		return fmt.Errorf("failed to initialize database connection to %s: %w", dbName, err)
	}
	b.database = dbConn

	// Initialize database schema on first run
	if err := dbConn.InitializeSchema(b.ctx); err != nil {
		logger.Error("Failed to initialize database schema", zap.Error(err))
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Verify schema exists
	if err := dbConn.VerifySchema(b.ctx); err != nil {
		logger.Error("Database schema verification failed", zap.Error(err))
		return fmt.Errorf("database schema verification failed: %w", err)
	}

	// Initialize EventsStored metric with current count
	if count, err := dbConn.GetTotalEventCount(b.ctx); err != nil {
		logger.Warn("Failed to get initial event count for metrics", zap.Error(err))
	} else {
		metrics.EventsStored.Set(float64(count))
		logger.Info("Initialized EventsStored metric", zap.Int64("count", count))
	}

	if err := b.database.RebuildBloomFilter(b.ctx); err != nil {
		logger.Warn("Failed to rebuild bloom filter", zap.Error(err))
	}

	// Initialize event dispatcher for real-time notifications
	b.eventDispatcher = storage.NewEventDispatcher(b.database)

	// Set the event dispatcher reference in the database for immediate local broadcasting
	b.database.SetEventDispatcher(b.eventDispatcher)

	logger.Info("✅ Event dispatcher initialized")

	return nil
}

// BuildWorkers initializes the worker pool(s).
func (b *NodeBuilder) BuildWorkers() {
	numCPU := runtime.NumCPU()
	b.workerPool = workers.NewWorkerPool(numCPU*2, numCPU*300)
}

// BuildValidators configures the validation logic.
func (b *NodeBuilder) BuildValidators() {
	b.validator = relay.NewPluginValidator(b.config, b.database)
	b.eventVal = relay.NewEventValidator(b.config, b.database)
}

// BuildProcessor sets up the event processor.
func (b *NodeBuilder) BuildProcessor() {
	// 100000 is the buffer size from your original code
	b.eventProc = storage.NewEventProcessor(b.ctx, b.database, 100000)
}

// BuildRateLimiter sets up the rate limiter.
func (b *NodeBuilder) BuildRateLimiter() {
	b.rateLimiter = limiter.NewRateLimiter(b.config)
}

// BuildLists loads blacklists/whitelists from config.
func (b *NodeBuilder) BuildLists() {
	blacklist := make(map[string]struct{})
	whitelist := make(map[string]struct{})

	for _, pk := range b.config.RelayPolicy.Blacklist.PubKeys {
		blacklist[strings.ToLower(pk)] = struct{}{}
	}
	for _, pk := range b.config.RelayPolicy.Whitelist.PubKeys {
		whitelist[strings.ToLower(pk)] = struct{}{}
	}

	b.blacklist = blacklist
	b.whitelist = whitelist
}

// Build finalizes the node construction.
func (b *NodeBuilder) Build() (*Node, error) {
	// Validate required components
	if b.database == nil {
		return nil, fmt.Errorf("database must be built before calling Build()")
	}
	if b.eventDispatcher == nil {
		return nil, fmt.Errorf("event dispatcher must be built before calling Build()")
	}
	if b.workerPool == nil {
		return nil, fmt.Errorf("worker pool must be built before calling Build()")
	}
	if b.validator == nil {
		return nil, fmt.Errorf("validator must be built before calling Build()")
	}
	if b.eventVal == nil {
		return nil, fmt.Errorf("event validator must be built before calling Build()")
	}
	if b.eventProc == nil {
		return nil, fmt.Errorf("event processor must be built before calling Build()")
	}
	if b.rateLimiter == nil {
		return nil, fmt.Errorf("rate limiter must be built before calling Build()")
	}

	node := &Node{
		ctx:             b.ctx,
		cancel:          b.cancel,
		db:              b.database,
		EventProcessor:  b.eventProc,
		EventDispatcher: b.eventDispatcher,
		config:          b.config,
		Validator:       b.validator,
		EventValidator:  b.eventVal,
		WorkerPool:      b.workerPool,
		wsConns:         make(map[domain.WebSocketConnection]bool),
		rateLimiter:     b.rateLimiter,

		blacklistPubKeys: b.blacklist,
		whitelistPubKeys: b.whitelist,
	}

	logger.Debug("Node initialized successfully via builder")
	b.database.StartExpiredEventsCleaner(b.ctx, time.Hour)
	return node, nil
}
