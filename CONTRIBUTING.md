# 🚀 Contributing to Shugur Relay

Thank you for your interest in contributing to Shugur Relay! This project thrives on community contributions, from bug reports to code enhancements. We welcome developers of all skill levels to join our growing ecosystem.

## 📋 Table of Contents

- [Ways to Contribute](#ways-to-contribute)
- [Development Workflow](#development-workflow)
- [Commit Message Convention](#commit-message-convention)
- [Release Process](#release-process)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Code Style](#code-style)
- [Getting Help](#getting-help)

## 🛠️ Ways to Contribute

### 🐛 **Bug Reports**
- Search existing issues first to avoid duplicates
- Use the bug report template with detailed reproduction steps
- Include environment details (OS, Go version, database version)
- Provide logs and error messages when possible

### ✨ **Feature Requests**
- Check if the feature aligns with project goals
- Use the feature request template
- Discuss implementation approach before coding
- Consider backward compatibility implications

### 📝 **Documentation**
- Improve existing documentation clarity
- Add examples and use cases
- Translate documentation (future consideration)
- Update API documentation for code changes

### 🔧 **Code Contributions**
- Bug fixes and security improvements
- Performance optimizations
- New Nostr NIP implementations
- Infrastructure and tooling improvements

## 🔄 Development Workflow

### Quick Start Guide

1. **Fork & Clone**
   ```bash
   git clone https://github.com/your-username/relay.git
   cd relay
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-awesome-feature
   # or
   git checkout -b fix/bug-description
   ```

3. **Make Changes**
   - Follow coding standards and add tests
   - Update documentation if needed
   - Test locally before committing

4. **Commit with Conventional Format**
   ```bash
   git commit -m "feat: add WebSocket connection pooling"
   git commit -m "fix: resolve memory leak in event processing"
   git commit -m "docs: update NIP-01 implementation guide"
   ```

5. **Push & Create PR**
   ```bash
   git push origin feature/your-awesome-feature
   ```
   Then create a Pull Request from GitHub interface.

## 📝 Commit Message Convention

We use **Conventional Commits** for consistent versioning and changelog generation:

### Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
- **`feat:`** - New features (minor version bump)
- **`fix:`** - Bug fixes (patch version bump)
- **`docs:`** - Documentation changes
- **`style:`** - Code formatting (no logic changes)
- **`refactor:`** - Code refactoring
- **`perf:`** - Performance improvements
- **`test:`** - Test additions or modifications
- **`chore:`** - Maintenance tasks
- **`ci:`** - CI/CD pipeline changes

### Breaking Changes
```bash
feat!: redesign configuration format
# or
feat: redesign configuration format

BREAKING CHANGE: Configuration file format changed from YAML to TOML
```

### Examples
```bash
feat(relay): add connection rate limiting
fix(storage): resolve race condition in event indexing
docs(nip): update NIP-11 relay information document
perf(filter): optimize subscription matching algorithm
```

## 🔄 Release Process

### Batched Release Strategy

We use a **batched release approach** to group multiple contributions:

1. **Development Phase**
   - Contributors submit PRs with conventional commits
   - PRs are reviewed and merged to `main` branch
   - **No automatic releases created**

2. **Release Preparation**
   - Maintainers trigger release-please when ready
   - System creates a **Release PR** with:
     - Version bump based on accumulated changes
     - Generated changelog
     - Updated VERSION file

3. **Release Execution**
   - Review and merge the Release PR
   - Automated release with artifacts (binaries, Docker images)
   - Tags created automatically

### Version Strategy
- **Patch** (`1.3.3` → `1.3.4`): Bug fixes only
- **Minor** (`1.3.3` → `1.4.0`): New features, backward compatible
- **Major** (`1.3.3` → `2.0.0`): Breaking changes

### Release Candidates
- Used for testing before stable releases
- Format: `v1.4.0-rc.1`, `v1.4.0-rc.2`
- Automatically progress until stable release

## 📋 Pull Request Guidelines

### Before Submitting

- [ ] **Single Responsibility**: Each PR addresses one feature/fix
- [ ] **Tests Pass**: Run full test suite locally
- [ ] **Conventional Commits**: Use proper commit message format
- [ ] **Documentation Updated**: Include relevant docs changes
- [ ] **No Breaking Changes**: Unless explicitly marked
- [ ] **Issue Reference**: Link to related issue (`Fixes #123`)

### PR Requirements

#### **Code Quality**
- [ ] Code follows project style guidelines
- [ ] All functions have appropriate comments
- [ ] Error handling is comprehensive
- [ ] No hardcoded values (use configuration)

#### **Testing Coverage**
- [ ] Unit tests for new functionality
- [ ] Integration tests for complex features
- [ ] Edge cases covered
- [ ] Performance impact considered

#### **Security Considerations**
- [ ] No credentials in code or logs
- [ ] Input validation implemented
- [ ] SQL injection prevention
- [ ] Rate limiting considerations

#### **Nostr Protocol Compliance**
- [ ] NIP specifications followed correctly
- [ ] Backward compatibility maintained
- [ ] Client compatibility tested
- [ ] Event validation implemented

### PR Size Guidelines

| Size | Lines Changed | Description | Review Time |
|------|---------------|-------------|-------------|
| **XS** | < 20 | Documentation, config | < 30 min |
| **S** | 20-100 | Bug fixes, minor features | 1-2 hours |
| **M** | 100-500 | New features, refactoring | 2-4 hours |
| **L** | 500-1000 | Major features, architecture | 1-2 days |
| **XL** | > 1000 | Split into smaller PRs | N/A |

### Review Process

1. **Automated Checks** (required)
   - CI/CD pipeline passes
   - Security scans pass
   - Performance benchmarks acceptable

2. **Code Review** (required)
   - At least 1 maintainer approval
   - Address all review comments
   - Final approval from core team

3. **Testing Phase** (for major features)
   - Deploy to staging environment
   - Integration testing with real Nostr clients
   - Performance testing under load

## 🛠️ Development Setup

### Prerequisites

| Component | Version | Purpose |
|-----------|---------|---------|
| **Go** | 1.24.4+ | Main development language |
| **CockroachDB** | v24.1.5+ | Primary database |
| **Docker** | 20.0+ | Development containers |
| **Git** | 2.0+ | Version control |

### Quick Setup (Recommended)

#### Option 1: Docker Development Environment

```bash
# Clone and setup
git clone https://github.com/Shugur-Network/relay.git
cd relay

# Start development environment
docker-compose -f docker/compose/docker-compose.local.yml up -d

# Install dependencies
go mod download

# Run relay in development mode
go run ./cmd --config config/development.yaml
```

#### Option 2: Manual Setup

```bash
# 1. Setup CockroachDB
docker run -d \
  --name cockroach-dev \
  -p 26257:26257 \
  -p 8080:8080 \
  cockroachdb/cockroach:v24.1.5 \
  start-single-node --insecure

# 2. Configure relay
cp config/development.yaml config.yaml
# Edit database connection settings if needed

# 3. Run relay
go run ./cmd

# 4. Verify installation
curl http://localhost:8080/health
```

### Development Tools

#### **Recommended VS Code Extensions**
- `golang.go` - Go language support
- `ms-vscode.vscode-json` - JSON schema validation
- `bradlc.vscode-tailwindcss` - For web interface
- `github.copilot` - AI assistance

#### **Development Commands**

```bash
# Build binary
make build

# Run tests
make test

# Run with race detection
make test-race

# Lint code
make lint

# Format code
make fmt

# Generate documentation
make docs

# Clean build artifacts
make clean
```

### Environment Configuration

#### **Development Environment Variables**

```bash
# Required
export SHUGUR_DB_HOST=localhost
export SHUGUR_DB_PORT=26257
export SHUGUR_DB_USER=root
export SHUGUR_DB_SSL_MODE=disable

# Optional
export SHUGUR_LOG_LEVEL=debug
export SHUGUR_METRICS_ENABLED=true
export SHUGUR_WEB_ENABLED=true
```

#### **Configuration Files**

| File | Purpose | Environment |
|------|---------|-------------|
| `config/development.yaml` | Local development | Development |
| `config/production.yaml` | Production template | Production |
| `config.yaml` | Your local config | Local override |

### Database Management

#### **Database Schema**

```bash
# Apply migrations
go run ./cmd migrate

# Reset database (⚠️ destructive)
go run ./cmd migrate --reset

# Check migration status
go run ./cmd migrate --status
```

#### **Database Operations**

```bash
# Connect to database
docker exec -it cockroach-dev ./cockroach sql --insecure

# Backup development data
docker exec cockroach-dev ./cockroach dump shugur --insecure > backup.sql

# Restore from backup
docker exec -i cockroach-dev ./cockroach sql --insecure < backup.sql
```

## 🧪 Testing

### Test Strategy

We maintain comprehensive test coverage across multiple layers:

#### **Unit Tests**
```bash
# Run all unit tests
go test ./...

# Run with coverage report
go test -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/relay -v

# Run with race detection
go test -race ./...
```

#### **Integration Tests**
```bash
# Start test database
docker run -d --name cockroach-test -p 26258:26257 \
  cockroachdb/cockroach:v24.1.5 start-single-node --insecure

# Run integration tests
SHUGUR_DB_PORT=26258 go test -tags=integration ./tests/integration/...
```

#### **NIP Compliance Tests**
```bash
cd tests/nips

# Test specific NIP implementation
./test_nip01.sh    # Basic protocol flow
./test_nip11.sh    # Relay information document
./test_nip15.sh    # End of stored events notice
./test_nip50.sh    # Search capability

# Test all implemented NIPs
./run_all_tests.sh
```

#### **Performance Tests**
```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Performance regression detection
go test -bench=. -count=5 ./internal/storage/... > bench.txt
```

### Test Writing Guidelines

#### **Unit Test Structure**
```go
func TestEventProcessor_ProcessEvent(t *testing.T) {
    // Arrange
    processor := NewEventProcessor(mockStorage, mockValidator)
    event := &nostr.Event{...}
    
    // Act
    result, err := processor.ProcessEvent(context.Background(), event)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expected, result.Status)
}
```

#### **Table-Driven Tests**
```go
func TestFilter_Matches(t *testing.T) {
    tests := []struct {
        name     string
        filter   Filter
        event    *nostr.Event
        expected bool
    }{
        {"exact kind match", Filter{Kinds: []int{1}}, &nostr.Event{Kind: 1}, true},
        {"kind mismatch", Filter{Kinds: []int{1}}, &nostr.Event{Kind: 2}, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.filter.Matches(tt.event))
        })
    }
}
```

#### **Test Data Management**
- Use `testdata/` directory for test fixtures
- Create helper functions for common test objects
- Use factories for complex object creation
- Clean up resources in test teardown

## 📏 Code Style & Standards

### Go Style Guidelines

#### **Formatting**
```bash
# Format code (required before commit)
go fmt ./...

# Imports organization
goimports -w .

# Lint code
golangci-lint run
```

#### **Naming Conventions**
```go
// ✅ Good
type EventProcessor struct {
    storage    Storage
    validator  Validator
    logger     *zap.Logger
}

func (ep *EventProcessor) ProcessEvent(ctx context.Context, event *nostr.Event) (*ProcessResult, error) {
    // Implementation
}

// ❌ Avoid
type ep struct {
    s Storage
    v Validator
    l *zap.Logger
}

func (e *ep) process(c context.Context, ev *nostr.Event) (*ProcessResult, error) {
    // Implementation
}
```

#### **Error Handling**
```go
// ✅ Good - Descriptive errors
if err := validator.ValidateEvent(event); err != nil {
    return nil, fmt.Errorf("event validation failed: %w", err)
}

// ✅ Good - Contextual errors
if err := storage.StoreEvent(ctx, event); err != nil {
    logger.Error("failed to store event", 
        zap.String("event_id", event.ID),
        zap.Error(err),
    )
    return nil, err
}

// ❌ Avoid - Silent failures
storage.StoreEvent(ctx, event)
```

#### **Documentation Standards**
```go
// ProcessEvent validates and stores a Nostr event.
// It returns a ProcessResult containing the operation status and any relevant metadata.
// Returns an error if validation fails or storage operation encounters an issue.
//
// The context should include relevant tracing information for monitoring.
func (ep *EventProcessor) ProcessEvent(ctx context.Context, event *nostr.Event) (*ProcessResult, error) {
    // Implementation
}
```

### Project Structure Guidelines

```
internal/
├── relay/          # Core relay functionality
├── storage/        # Database layer
├── config/         # Configuration management
├── logger/         # Logging utilities
├── metrics/        # Monitoring and metrics
└── domain/         # Business logic interfaces

cmd/                # CLI applications
tests/              # Integration and E2E tests
web/                # Web interface assets
docker/             # Container configurations
```

### Performance Considerations

- **Memory Management**: Minimize allocations in hot paths
- **Goroutine Safety**: Use proper synchronization primitives
- **Database Queries**: Optimize queries and use proper indexing
- **Caching**: Implement appropriate caching strategies
- **Monitoring**: Add metrics for critical operations

## 🤝 Community Guidelines

### Code of Conduct

We are committed to fostering an inclusive and welcoming community. By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

**Our Standards:**
- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Respect different viewpoints and experiences
- Take responsibility for mistakes and learn from them

### Communication Channels

| Channel | Purpose | Response Time |
|---------|---------|---------------|
| **GitHub Issues** | Bug reports, feature requests | 24-48 hours |
| **GitHub Discussions** | Questions, ideas, community chat | 1-3 days |
| **Pull Requests** | Code review and collaboration | 1-2 days |

### Recognition

We appreciate all contributions! Contributors will be:
- Listed in our `CONTRIBUTORS.md` file
- Mentioned in release notes for significant contributions
- Eligible for community rewards and recognition

## 🆘 Getting Help

### Before Asking for Help

1. **Search Documentation**
   - Check the [README](README.md) for basic setup
   - Review existing [issues](https://github.com/Shugur-Network/relay/issues)
   - Browse [pull requests](https://github.com/Shugur-Network/relay/pulls) for similar work

2. **Check Resources**
   - [Nostr Protocol Documentation](https://github.com/nostr-protocol/nostr)
   - [NIP Specifications](https://github.com/nostr-protocol/nips)
   - [CockroachDB Documentation](https://www.cockroachlabs.com/docs/)

### How to Ask for Help

#### **For Bug Reports**
Use the bug report template with:
- Detailed steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Relevant logs and error messages

#### **For Feature Requests**
Use the feature request template with:
- Clear description of the feature
- Use case and motivation
- Proposed implementation approach
- Consideration of alternatives

#### **For Questions**
- Be specific about what you're trying to achieve
- Include relevant code snippets
- Mention what you've already tried
- Specify your development environment

### Support Resources

- **Documentation**: Comprehensive guides and API docs
- **Examples**: Sample configurations and implementations
- **Community**: Active community of contributors and users
- **Maintainers**: Core team available for complex issues

### Response Expectations

| Issue Type | Target Response | Target Resolution |
|------------|----------------|-------------------|
| **Security Issues** | 2 hours | 24 hours |
| **Critical Bugs** | 4 hours | 2-3 days |
| **Bug Reports** | 24 hours | 1-2 weeks |
| **Feature Requests** | 48 hours | Discussion-based |
| **Questions** | 1-3 days | As needed |

---

## 🎉 Thank You!

Thank you for taking the time to contribute to Shugur Relay! Your contributions help build a more robust and feature-rich Nostr ecosystem. Every contribution, no matter how small, makes a difference.

**Happy coding!** 🚀

---

*This document is a living guide. If you have suggestions for improvements, please open an issue or submit a pull request.*
