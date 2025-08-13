# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
