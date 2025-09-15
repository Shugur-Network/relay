# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.3](https://github.com/Shugur-Network/relay/compare/v1.3.2...v1.3.3) (2025-09-15)


### Bug Fixes

* Extract real client IPs from proxy headers (v1.3.2.1) ([0cd278c](https://github.com/Shugur-Network/relay/commit/0cd278c8486bfff6cfff55d912ec51fdceb1f6ea))


### Performance Improvements

* set fixed preallocation for query results to 500 (matches typical filter cap) ([#20](https://github.com/Shugur-Network/relay/issues/20)) ([a6fb50e](https://github.com/Shugur-Network/relay/commit/a6fb50e8e7be6b689ab2b99317aa77d5b2059f06))

## [Unreleased]

### Added

- Native Prometheus metrics endpoint (`/metrics`) exposed via the HTTP server for scraping
- Repository standards and security hardening:
  - CODEOWNERS for default ownership and review routing
  - CodeQL code scanning workflow for Go
  - Trivy filesystem vulnerability scan in CI (CRITICAL/HIGH)
  - Dependabot (weekly) for Go modules and GitHub Actions

### Fixed

- **IP Address Extraction Behind Reverse Proxy** (v1.3.2.1):
  - Fixed issue where all client connections appeared as `127.0.0.1` (Caddy proxy IP) instead of real client IPs
  - Added proper extraction of real client IPs from `X-Real-IP` and `X-Forwarded-For` headers set by Caddy
  - Rate limiting and banning now works correctly per real client IP instead of globally affecting all clients
  - Enhanced logging with comprehensive IP extraction debugging and connection tracking
- **Static file path traversal**: sanitized and bounded `/static/*` paths to the `web/static` root, returning 400 on invalid paths; added `X-Content-Type-Options: nosniff` and caching headers
- **Uncontrolled allocation (GetEvents)**: preallocation at the allocation site is now a fixed, small cap (no user‑influenced size) to prevent excessive allocations

### Changed

- **Logging Optimization** (v1.3.2.1):
  - Moved verbose connection logs from `Info` to `Debug` level to reduce log volume in production
  - Connection establishment, close events, and rate limit violations now only appear in debug mode
  - Important security events (bans, blocked connections) remain at `Info` level for production monitoring
- **CI/CD**:
  - Least‑privilege default permissions and concurrency (cancel in‑progress runs per ref)
  - Skip CI for docs/images‑only changes (paths‑ignore)
  - Use setup‑go module cache; split builds (PR: linux/amd64; main: full matrix)
  - Build & push Docker images only on `main` (and releases)

## [1.3.2] - 2025-09-11

### Changed

- **Time Capsules Protocol Redesign (Breaking Change)**:
  - **BREAKING**: Replaced previous Time Capsules implementation with new NIP-XX Time Capsule specification (kind 1041) <https://github.com/Shugur-Network/NIP-XX_Time-Capsules/blob/main/NIP-XX_Time-Capsules.md>

- **New Time Capsule Implementation**:
  - Implemented kind 1041 time capsule events with drand-based timelock encryption
  - Improve integration and compliance with NIP-44 v2 for encryption and NIP-59 for gift wrapping
  - Updated validation pipeline for new event structure and payload format
  - Integrated drand randomness beacon network for decentralized timelock functionality

- **Database Schema Migration**:
  - Enhanced database indexes for efficient addressable queries and validation

- **Improve Expired Event Handling**:
  - Improved expired event cleanup and handling logic
  - Enhanced relay metadata with Time Capsules capability information

### Added

- **Enhanced Cryptographic Support**:
  - Proper payload structure validation for both public and private modes
  - Drand network integration for timelock encryption and decryption

- **New Testing Infrastructure**:
  - Created `test_nip_time_capsules.sh` - simplified interactive test script
  - Implemented complete round-trip testing (encrypt → publish → wait → decrypt)
  - Comprehensive validation of public/private timelock scenarios

- **Advanced Validation System**:
  - Enhanced event validation in `nip_time_capsules.go` with mode-specific checks
  - Proper tlock tag parsing and validation
  - Payload size limits and structure validation for both modes
  - Binary payload parsing with proper offset handling and length validation

### Removed

- **Deprecated Time Capsules Components**:
  - Removed Shamir's secret sharing implementation
  - Removed witness coordination system (kinds 1990, 1991, 1992)
  - Removed threshold-based unlocking mechanism
  - Removed share distribution endpoints
  - Removed addressable event support (kind 30095)
  - Removed external storage verification system

### Fixed

- **Addressable Event Processing**: Fixed addressable event processing to properly handle all event kinds
- **Temporary Events**: Fixed temporary event handling to ensure ephemeral events are not stored
- **Migration Issues**: Properly migrated from multi-kind to single-kind approach
- **Tlock Tag Syntax**: Corrected tlock tag format to use simple `["tlock", chain, round]` structure
- **Binary Payload Handling**: Fixed mode byte extraction and payload parsing
- **Test File Cleanup**: Removed corrupted files with binary characters in filenames
- **Validation Logic**: Enhanced error handling and validation coverage

## [1.3.0] - 2025-08-30

### Added

- **Time Capsules Feature (NIP Implementation)**:
  - Implemented complete Time Capsules protocol with event kinds 1990, 30095, 1991, 1992
  - Added threshold-based and scheduled unlock mechanisms
  - Support for Shamir's secret sharing with configurable witness thresholds
  - Comprehensive validation for time-locked events and unlock shares
  - Share distribution system for witness coordination
  - External storage support with integrity verification (URI, SHA256)
  - NIP-11 capability advertisement for Time Capsules support
  - Created extensive test suite with 47 comprehensive tests (100% pass rate)
  - Standard Nostr tag conventions (p for witnesses, e for references)

- **Enhanced Build System**:
  - Completely refactored build script with improved functionality and user experience
  - Added support for multiple build targets and configurations
  - Enhanced error handling and logging in build process
  - Improved cross-platform compatibility

- **Configurable Relay Identity**:
  - Added PUBLIC_KEY configuration field with validation
  - Support for 64-character hex public keys with automatic fallback
  - Enhanced relay metadata generation with configured identity
  - Improved relay identification and discovery capabilities

- **Relay Status and Monitoring**:
  - Added Time Capsules status endpoint for monitoring active capsules
  - Enhanced relay metadata with Time Capsules capability information
  - Improved event processing metrics and monitoring

### Changed

- **Event Kind Updates**:
  - Updated Time Capsules event kinds from 1199x to 199x range for better compatibility
  - Improved event kind validation and processing
  - Enhanced addressable event support (kind 30095)

- **Code Quality Improvements**:
  - Applied comprehensive code formatting (go fmt) across all Go files
  - Fixed staticcheck linting issues (converted if-else to switch statements)
  - Improved code structure and readability
  - Enhanced error handling and validation throughout the codebase

### Fixed

- **Repository Cleanup**:
  - Removed deprecated Time Capsules documentation and test files
  - Cleaned up repository structure and removed unused portal integration files
  - Fixed formatting inconsistencies and linting issues

## [1.2.0] - 2025-08-24

### Added

- **NIP-65 Support (Relay List Metadata)**:
  - Implemented kind 10002 relay list metadata events with comprehensive validation
  - Added proper "r" tag validation for relay URLs and markers ("read", "write")
  - Enhanced replaceable event handling to support NIP-01 range (10000-19999)
  - Created comprehensive test suite for NIP-65 functionality
  - Updated database schema to support new replaceable event ranges

- **NIP-45 Support (COUNT Command)**:
  - Implemented COUNT command for efficient event counting
  - Added dedicated NIP-45 module with proper validation and error handling
  - Created comprehensive test suite for COUNT operations
  - Added NIP-45 to relay metadata and supported NIPs list

- **Enhanced Event Processing**:
  - Improved ephemeral event handling (kinds 20000-29999) - now properly excluded from storage
  - Enhanced replaceable event validation to include kind 41 (channel metadata)
  - Better NIP-16 event treatment compliance with proper kind range handling

- **Distributed Relay Enhancements**:
  - Implemented real-time event synchronization using CockroachDB changefeeds
  - Added cross-node event broadcasting and synchronization
  - Enhanced distributed relay coordination and failover capabilities
  - Improved installation script for distributed relay setups

- **Infrastructure Improvements**:
  - Enhanced certificate management and ownership handling
  - Better FQDN prompting and validation in installation scripts
  - Improved port availability checks and cleanup processes
  - Enhanced Docker build and deployment pipeline

### Fixed

- **Event Validation**: Corrected various edge cases in NIP validation
- **Storage**: Fixed ephemeral event storage issues - ephemeral events are now properly excluded from persistent storage
- **Installation**: Improved reliability of installation scripts with better error handling

### Changed

- **Database Schema**: Updated to support extended replaceable event ranges per NIP-01
- **Event Processing**: Enhanced event treatment validation for better NIP compliance

## [1.1.0] - 2025-08-17

### Added

- **Enhanced NIP Validation System**:
  - Improved NIP validation and testing infrastructure
  - Enhanced gift wrap event validation using standardized NIP validation methods
  - Better NIP compliance checking across all supported protocols

- **Metrics and Monitoring Enhancements**:
  - Enhanced metrics tracking system with real-time capabilities
  - Added comprehensive real-time metrics API endpoints
  - Improved Prometheus metrics collection and reporting

- **Configuration Improvements**:
  - Added MaxConnections field to LimitationData for better connection management
  - Implemented configurable content length limits for relay metadata and WebSocket connections
  - Enhanced template updates for better metadata display

- **Web Interface Enhancements**:
  - Added SVG logo support for improved branding
  - Created enhanced dashboard interface with modern design
  - Updated web interface styling to match shugur.net design system

### Fixed

- **Logging and Debugging**:
  - Corrected delegation logging to use proper struct field instead of slice indexing
  - Optimized logging across relay components with appropriate log levels
  - Enhanced NIP validation visibility in logs

- **Configuration and Setup**:
  - Updated relay URLs in test scripts to use secure WebSocket connections
  - Fixed repository references from 'Relay' to 'relay' for consistency
  - Updated .gitignore and fixed import paths for better consistency

- **Connection Management**:
  - Removed maxConnections fallback logic from DefaultRelayMetadata function for cleaner implementation
  - Improved WebSocket connection handling and limits

### Changed

- **Development and Testing**:
  - Enhanced NIP validation testing framework
  - Improved test script reliability with secure connections
  - Better error handling and validation across NIP implementations

- **Documentation and Branding**:
  - Updated logo image for improved branding consistency
  - Enhanced documentation for installation, configuration, and troubleshooting
  - Improved user guidance and fixed database references

- **Code Organization**:
  - Refactored gift wrap event validation for better maintainability
  - Improved code structure and import organization
  - Enhanced error handling and validation patterns

### Added

- **New Documentation Files** (moved to [dedicated documentation repository](https://github.com/Shugur-Network/docs)):
  - `API.md` - Comprehensive API reference for WebSocket (Nostr protocol) and HTTP endpoints
  - `TROUBLESHOOTING.md` - Detailed troubleshooting guide with common issues and solutions
  - `PERFORMANCE.md` - Performance optimization and scaling recommendations
  - `CHANGELOG.md` - This changelog file to track project changes

### Fixed

- **High Priority Documentation Issues** (moved to [dedicated documentation repository](https://github.com/Shugur-Network/docs)):
  - **Empty Installation File**: Populated `installation/INSTALLATION.md` with comprehensive installation overview and decision tree
  - **Configuration Validation**: Fixed database configuration examples in `BARE-METAL.md` to match actual Go struct definitions
  - **Version Updates**: Updated Go version requirement from 1.21 to 1.24 and CockroachDB from v23.1.x to v24.1.x
  - **Removed Non-existent Parameters**: Eliminated references to `NAME`, `USER`, and `SSL_MODE` database parameters that don't exist in the codebase

### Changed

- **README.md**:
  - Updated logo to use `banner.png` with full width display
  - Enhanced NIP support matrix with categorization (Core, Enhanced, Advanced)
  - Improved quick start instructions with multiple installation options
  - Added comprehensive feature list highlighting production-readiness

- **Documentation Structure** (moved to [dedicated documentation repository](https://github.com/Shugur-Network/docs)):
  - Updated `INTRODUCTION.md` table of contents to include new documentation
  - Enhanced `installation/QUICK-START.md` with links to operational guides
  - Improved cross-references between documentation files

- **Configuration Documentation** (moved to [dedicated documentation repository](https://github.com/Shugur-Network/docs/blob/main/CONFIGURATION.md)):
  - **BREAKING**: Corrected database configuration parameters to match actual implementation
  - Removed non-existent database parameters (`NAME`, `USER`, `PASSWORD`, `SSL_MODE`, `MAX_OPEN_CONNS`, etc.)
  - Updated to only include actual parameters (`SERVER` and `PORT`)
  - Added explanation of automatic connection string management
  - Clarified metrics configuration behavior
  - Added note about environment variable naming discrepancies in install scripts

### Fixed

- **Installation Scripts**:
  - **Enhanced Cleanup**: `install.distributed.sh` now properly cleans up temporary staging directories after installation
  - **Failure Recovery**: Added exit traps to ensure cleanup occurs even if installation fails
  - **Standalone Script**: Configuration files (`Caddyfile`, `config.yaml`, `docker-compose.standalone.yml`) now remain in installation directory as intended
  - **Distributed Script**: Improved cleanup to handle root-owned directories created during certificate generation

- **API Documentation**:
  - Removed incorrect `/metrics` endpoint documentation (endpoint doesn't exist in application)
  - Corrected API endpoint listings to match actual implementation
  - Fixed metrics description to reflect internal collection vs. HTTP endpoint

- **Configuration Issues**:
  - Fixed database configuration section to match actual code implementation
  - Corrected default metrics port inconsistencies between environments
  - Updated environment variable examples to match `mapstructure` tags

- **Troubleshooting Documentation**:
  - Updated diagnostic commands to use actual API endpoints (`/api/stats` instead of non-existent `/metrics`)
  - Fixed connection testing procedures

- **Performance Documentation**:
  - Removed references to non-existent database configuration parameters
  - Updated configuration examples to use actual parameters only
  - Corrected metrics port in production examples

- **Reference Links**:
  - Fixed broken internal documentation links
  - Corrected file path references (e.g., `installation/README.md` → `installation/INSTALLATION.md`)

- **Medium Priority Documentation Issues**:
  - **Cross-References**: Added comprehensive cross-references between related documentation sections
  - **Documentation Structure**: Enhanced navigation with "Related Documentation" sections across all major docs
  - **Installation Flow**: Improved user journey from Getting Started through Installation to Configuration
  - **Developer Experience**: Added cross-references in Contributing guide for better development setup

- **Low Priority Documentation Issues**:
  - **User Experience**: Added helpful tips, best practices, and warnings throughout documentation
  - **Visual Hierarchy**: Enhanced formatting with callout boxes and consistent styling
  - **Best Practices**: Added operational tips for production deployments and development
  - **Readability**: Improved documentation flow with helpful examples and guidance

### Removed

- **Incorrect Documentation**:
  - Removed non-existent database configuration parameters from documentation
  - Removed false claims about separate metrics HTTP server
  - Removed placeholder "Coming Soon" content in favor of complete documentation

## [Previous Versions]

### Note

This changelog starts from the documentation review and improvement phase. For earlier version history, please refer to the git commit history.

---

## Documentation Quality Improvements Summary

This release represents a major documentation quality assurance initiative that included:

1. **Comprehensive Code Validation**: All documentation was verified against actual source code implementation
2. **API Accuracy**: API documentation now accurately reflects implemented endpoints
3. **Configuration Correctness**: All configuration parameters verified against Go struct definitions
4. **Complete Coverage**: Added missing documentation for troubleshooting, performance, and API reference
5. **Professional Standards**: All documentation now meets enterprise-grade documentation standards

### Technical Validation Process

- **Database Configuration**: Verified against `internal/config/database.go` and connection logic in `internal/application/node_builder.go`
- **API Endpoints**: Validated against HTTP handlers in `internal/relay/server.go` and `internal/web/handler.go`
- **Environment Variables**: Checked against `mapstructure` tags in configuration structs
- **Metrics Implementation**: Verified Prometheus metrics collection in `internal/metrics/relay.go`

The documentation ecosystem is now comprehensive, accurate, and aligned with the actual codebase implementation.
