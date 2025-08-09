# Comprehensive Makefile for Shugur Relay
# Production-ready build, test, and deployment automation

# Project information
PROJECT_NAME := shugur-relay
BINARY_NAME := relay
PACKAGE := github.com/Shugur-Network/Relay

# Version information
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build paths
BUILD_DIR := bin
DIST_DIR := dist
CMD_DIR := cmd
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)

# Go configuration
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags and linker flags
BUILD_FLAGS := -v
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
RACE_FLAGS := -race
CGO_ENABLED := 0

# Docker configuration
DOCKER_IMAGE := ghcr.io/shugur-network/relay
DOCKER_TAG ?= $(VERSION)

# Test configuration
TEST_FLAGS := -v -race
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Default target
.PHONY: all
all: clean lint test build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/ $(DIST_DIR)/ $(COVERAGE_FILE) $(COVERAGE_HTML)
	@$(GOCLEAN) -cache -testcache -modcache

# Create build directories
.PHONY: directories
directories:
	@mkdir -p $(BUILD_DIR) $(DIST_DIR)

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) verify

# Development build
.PHONY: dev
dev: directories
	@echo "Building development version..."
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH) ./$(CMD_DIR)
	@echo "Development build completed: $(BINARY_PATH)"

# Production build
.PHONY: build
build: directories
	@echo "Building production version $(VERSION)..."
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_PATH) ./$(CMD_DIR)
	@echo "Production build completed: $(BINARY_PATH)"

# Build with race detection
.PHONY: build-race
build-race: directories
	@echo "Building with race detection..."
	$(GOBUILD) $(BUILD_FLAGS) $(RACE_FLAGS) -o $(BINARY_PATH) ./$(CMD_DIR)
	@echo "Race detection build completed: $(BINARY_PATH)"

# Cross-compilation targets  
.PHONY: build-linux
build-linux: directories
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)

.PHONY: build-linux-arm64
build-linux-arm64: directories
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)

.PHONY: build-darwin
build-darwin: directories
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)

.PHONY: build-darwin-arm64
build-darwin-arm64: directories
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)

.PHONY: build-windows
build-windows: directories
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

.PHONY: build-all
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows
	@echo "All platform builds completed"
	@ls -la $(DIST_DIR)/

# Create release archives
.PHONY: release
release: build-all
	@echo "Creating release archives..."
	@cd $(DIST_DIR) && \
	for binary in *; do \
		if [[ $$binary == *.exe ]]; then \
			zip "$${binary%.exe}.zip" "$$binary"; \
		else \
			tar -czf "$$binary.tar.gz" "$$binary"; \
		fi; \
	done
	@echo "Release archives created in $(DIST_DIR)/"

# Development tools installation
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	@$(GOGET) -u golang.org/x/tools/cmd/goimports
	@$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOGET) -u github.com/air-verse/air@latest
	@$(GOGET) -u github.com/securecodewarrior/github-action-gosec/cmd/gosec@latest

# Code formatting
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@goimports -w .

# Code linting
.PHONY: lint
lint:
	@echo "Running linters..."
	@golangci-lint run --timeout=5m

# Security scanning
.PHONY: security
security:
	@echo "Running security scan..."
	@gosec ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@$(GOTEST) $(TEST_FLAGS) ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) $(TEST_FLAGS) -coverprofile=$(COVERAGE_FILE) ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run integration tests
.PHONY: test-integration
test-integration: build
	@echo "Running integration tests..."
	@./bin/relay --version
	@for test_file in tests/test_nip*.sh; do \
		if [ -f "$$test_file" ]; then \
			echo "Running $$test_file"; \
			bash "$$test_file" || exit 1; \
		fi; \
	done

# Module management
.PHONY: tidy
tidy:
	@echo "Tidying modules..."
	@$(GOMOD) tidy

.PHONY: update
update:
	@echo "Updating dependencies..."
	@$(GOMOD) tidy
	@$(GOGET) -u ./...

# Vendor dependencies
.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	@$(GOMOD) vendor

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		--target production \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		-f docker/Dockerfile .

.PHONY: docker-build-dev
docker-build-dev:
	@echo "Building development Docker image..."
	@docker build \
		--target development \
		-t $(DOCKER_IMAGE):dev \
		-f docker/Dockerfile .

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest

# Development environment
.PHONY: dev-up
dev-up:
	@echo "Starting development environment..."
	@docker-compose -f docker/compose/docker-compose.dev.yml up -d

.PHONY: dev-down
dev-down:
	@echo "Stopping development environment..."
	@docker-compose -f docker/compose/docker-compose.dev.yml down

.PHONY: dev-logs
dev-logs:
	@docker-compose -f docker/compose/docker-compose.dev.yml logs -f

# Production environment
.PHONY: prod-up
prod-up:
	@echo "Starting production environment..."
	@docker-compose -f docker/compose/docker-compose.production.yml up -d

.PHONY: prod-down
prod-down:
	@echo "Stopping production environment..."
	@docker-compose -f docker/compose/docker-compose.production.yml down

.PHONY: prod-logs
prod-logs:
	@docker-compose -f docker/compose/docker-compose.production.yml logs -f

# Database management
.PHONY: db-up
db-up:
	@echo "Starting database..."
	@docker-compose -f docker/compose/docker-compose.dev.yml up -d cockroach

.PHONY: db-migrate
db-migrate: build
	@echo "Running database migrations..."
	@./bin/relay migrate

.PHONY: db-reset
db-reset:
	@echo "Resetting database..."
	@docker-compose -f docker/compose/docker-compose.dev.yml down -v cockroach
	@docker-compose -f docker/compose/docker-compose.dev.yml up -d cockroach

# Run application
.PHONY: run
run: build
	@echo "Starting relay..."
	@$(BINARY_PATH)

.PHONY: run-dev
run-dev: dev
	@echo "Starting relay in development mode..."
	@$(BINARY_PATH) --config config.yaml --log-level debug

# Benchmarking
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

# Version management
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

.PHONY: bump-major
bump-major:
	@echo "$(shell echo $(VERSION) | awk -F. '{print $$1+1".0.0"}')" > VERSION
	@echo "Version bumped to $(shell cat VERSION)"

.PHONY: bump-minor
bump-minor:
	@echo "$(shell echo $(VERSION) | awk -F. '{print $$1"."$$2+1".0"}')" > VERSION
	@echo "Version bumped to $(shell cat VERSION)"

.PHONY: bump-patch
bump-patch:
	@echo "$(shell echo $(VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}')" > VERSION
	@echo "Version bumped to $(shell cat VERSION)"

# Quality assurance
.PHONY: qa
qa: clean fmt lint security test test-coverage
	@echo "Quality assurance completed successfully"

# CI pipeline
.PHONY: ci
ci: clean deps fmt lint security test build
	@echo "CI pipeline completed successfully"

# Help target
.PHONY: help
help:
	@echo "Shugur Relay - Available Make targets:"
	@echo ""
	@echo "Building:"
	@echo "  build           - Build production binary"
	@echo "  dev             - Build development binary" 
	@echo "  build-race      - Build with race detection"
	@echo "  build-all       - Build for all platforms"
	@echo "  build-linux     - Build for Linux (amd64)"
	@echo "  build-linux-arm64 - Build for Linux (arm64)"
	@echo "  build-darwin    - Build for macOS (amd64)"
	@echo "  build-darwin-arm64 - Build for macOS (arm64)"
	@echo "  build-windows   - Build for Windows (amd64)"
	@echo "  release         - Create release archives"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run unit tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  test-integration - Run integration tests"
	@echo "  bench           - Run benchmarks"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linters"
	@echo "  security        - Security scan"
	@echo "  qa              - Full quality assurance"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build    - Build production Docker image"
	@echo "  docker-build-dev - Build development Docker image"
	@echo "  docker-push     - Push Docker image"
	@echo ""
	@echo "Environment:"
	@echo "  dev-up          - Start development environment"
	@echo "  dev-down        - Stop development environment"
	@echo "  prod-up         - Start production environment"
	@echo "  prod-down       - Stop production environment"
	@echo "  db-up           - Start database only"
	@echo "  db-migrate      - Run database migrations"
	@echo "  db-reset        - Reset database"
	@echo ""
	@echo "Development:"
	@echo "  run             - Run production build"
	@echo "  run-dev         - Run development build"
	@echo "  install-tools   - Install development tools"
	@echo "  deps            - Download dependencies"
	@echo "  tidy            - Tidy modules"
	@echo "  update          - Update dependencies"
	@echo "  vendor          - Vendor dependencies"
	@echo ""
	@echo "Version:"
	@echo "  version         - Show version information"
	@echo "  bump-major      - Bump major version"
	@echo "  bump-minor      - Bump minor version"  
	@echo "  bump-patch      - Bump patch version"
	@echo ""
	@echo "Utility:"
	@echo "  clean           - Clean build artifacts"
	@echo "  ci              - Run CI pipeline"
	@echo "  help            - Show this help"
