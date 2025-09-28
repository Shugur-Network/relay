# Shugur Relay - AI Coding Agent Instructions

## Architecture Overview

**Shugur Relay** is a production-grade Nostr relay built in Go with CockroachDB. The architecture follows a layered design:

- **`cmd/`** - CLI entry points using Cobra (`main.go`, `root.go`, `version.go`)
- **`internal/application/`** - Application orchestration (`Node` struct coordinates all components)
- **`internal/domain/`** - Core interfaces and business logic contracts
- **`internal/relay/`** - Nostr protocol handling and WebSocket management
- **`internal/storage/`** - CockroachDB operations with connection pooling
- **`internal/config/`** - Configuration management using Viper with validation

## Core Components & Patterns

### 1. Node Architecture (`internal/application/node.go`)
The `Node` struct is the central orchestrator that ties together:
```go
type Node struct {
    db              *storage.DB
    WorkerPool      *workers.WorkerPool
    EventProcessor  *storage.EventProcessor
    EventDispatcher *storage.EventDispatcher
    Validator       domain.EventValidator
    // ...
}
```

### 2. NIP Validation System (`internal/relay/nips/`)
Each NIP has its own file (e.g., `nip01.go`, `nip58.go`). **Follow this pattern for new NIPs:**
- Create `nip##.go` with validation functions named `ValidateXxx(event *nostr.Event) error`
- Add validation case to `internal/relay/plugin_validator.go` in `validateWithDedicatedNIPs()`
- Add kind to `AllowedKinds` map and required tags to `RequiredTags` map
- Update `internal/constants/relay_metadata.go` to advertise NIP support

### 3. Configuration Pattern (`internal/config/`)
Uses embedded YAML defaults (`defaults.yaml`) with struct tags for validation:
```go
type Config struct {
    General     GeneralConfig     `mapstructure:"general" validate:"required"`
    Relay       RelayConfig       `mapstructure:"relay" validate:"required"`
    // Custom validators registered in init()
}
```

### 4. Database Operations (`internal/storage/`)
- Connection pooling optimized for WebSocket load in `createPoolBasedOnLoad()`
- Event processing uses worker pools for concurrent database operations
- Bloom filters for duplicate detection and query optimization

## Development Workflows

### Building & Running
```bash
# Quick development build
go build -o ./bin/relay ./cmd

# Production build with version info (preferred)
make build

# Run with development config
./bin/relay start --config config/development.yaml

# Use VS Code tasks (Ctrl+Shift+P → "Tasks: Run Task")
# - "Build Relay" (default build task)
# - "Run Relay" (builds then runs)
```

### Testing NIPs
```bash
# Run specific NIP test with standardized format
./tests/nips/test_nip01.sh

# All NIP tests follow this UX pattern:
# - Colored output with ✓/✗ indicators  
# - Test counter and summary
# - Exit code 0 for success, 1 for failures
# - Uses `nak` CLI tool for event publishing
```

### Environment Setup
```bash
# Development database (CockroachDB)
docker-compose -f docker/compose/docker-compose.development.yml up -d

# Production environment
make prod-up

# Multiple environments simultaneously (different ports)
# - Development: ws://localhost:8081  
# - Test: ws://localhost:8082
# - Production: ws://localhost:8080
```

## Project-Specific Conventions

### 1. Error Handling
- Use structured logging with `zap`: `logger.Debug("NIP-58: Validating badge", zap.String("event_id", event.ID))`
- Return wrapped errors with context: `fmt.Errorf("invalid badge definition tags: %w", err)`
- Validation functions return detailed error messages for debugging

### 2. Event Validation Pipeline
Events flow through: **Basic validation** → **NIP-specific validation** → **Plugin validator** → **Database storage**

Located in `internal/relay/plugin_validator.go:validateWithDedicatedNIPs()`:
```go
switch event.Kind {
case 30009: // Badge Definition
    return nips.ValidateBadgeDefinition(event)
case 8:     // Badge Award  
    return nips.ValidateBadgeAward(event)
}
```

### 3. Configuration Loading
Configuration precedence: **CLI flags** > **Config file** > **Environment variables** > **Embedded defaults**
- Use `config.Load(cfgFile, logger)` in command initialization
- Override with CLI flags in `PersistentPreRunE`

### 4. Metrics & Observability
- Prometheus metrics exposed at `/metrics` endpoint
- Use `internal/metrics` package for custom metrics
- Structured logging with correlation IDs for tracing

## Integration Points

### 1. CockroachDB Integration
- Connection string format: `postgresql://user:pass@host:port/database?sslmode=require`
- Pool size scales with `maxWSConnections` setting
- Schema management via `internal/storage/schema.sql`

### 2. Nostr Protocol (`github.com/nbd-wtf/go-nostr`)
- Event validation: `event.CheckSignature()` for BIP-340 verification
- Filter matching: Use `nostr.Filter` for event queries
- WebSocket message types: `["EVENT", ...]`, `["REQ", ...]`, `["CLOSE", ...]`

### 3. Rate Limiting (`internal/limiter/`)
- Per-connection limits configurable in `relay.throttling_config`
- Uses token bucket algorithm with time-based refill

## Critical Development Notes

### 1. NIP Test Development
When creating NIP tests, use the standardized pattern from `tests/nips/test_nip58.sh`:
```bash
# Standardized test function signatures
print_result() { local test_name=$1; local success=$2; local nip=$3; }
# Test naming: "Test N: Description (NIP-XX)"
# Color output: GREEN/RED/BLUE/YELLOW with NC reset
```

### 2. Database Schema Changes
- Update `internal/storage/schema.sql` for schema changes
- Use database migrations for production deployments
- Test with `make db-reset` for clean state

### 3. Docker Deployment
- Multi-stage build in `docker/Dockerfile`
- Environment-specific compose files in `docker/compose/`
- Health checks and graceful shutdown built-in

### 4. Version Management
Version info set via build-time ldflags:
```bash
-X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD)
```

Use `make bump-patch/minor/major` for version updates.

## Quick Reference Commands

```bash
# Essential development commands
make build          # Production build with version info
make dev           # Development build  
make test          # Run all tests
make lint          # Code quality checks
make qa            # Full quality assurance pipeline

# Environment management
make dev-up        # Start development environment
make db-reset      # Reset development database
./bin/relay start --config config/development.yaml

# Testing
bash tests/nips/test_nip##.sh  # Test specific NIP
make test-integration          # Full integration test suite
```

This codebase prioritizes production reliability, comprehensive testing, and maintainable architecture. Follow the established patterns for consistency and leverage the extensive tooling for efficient development.