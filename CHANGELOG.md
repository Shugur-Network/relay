# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.3] - 2025-09-30

### Features

- **nips**: Add comprehensive support for NIP-60 (Cashu Wallets), NIP-61 (Nutzaps), and NIP-72 (Moderated Communities)
- **nips**: Add support for NIP-51 (Lists), NIP-52 (Calendar Events), NIP-53 (Live Activities), NIP-54 (Wiki)
- **nips**: Add support for NIP-56 (Reporting), NIP-57 (Lightning Zaps), and NIP-58 (Badges)
- **nips**: Fix NIP-09 (Event Deletion) test configuration and validation
- Production improvements with comprehensive error handling and optimization (#68) ([e79d4b9](https://github.com/Shugur-Network/Relay/commit/e79d4b9c39e305df8db9b91f3fe4669da997b7e0))
- Comprehensive lean release system with full CI/CD pipeline ([cf458ee](https://github.com/Shugur-Network/Relay/commit/cf458ee11bb8b4471b08b7fa02c7198c0ad36f5c))
- Add MaxConnections field to LimitationData and update related templates ([e710be9](https://github.com/Shugur-Network/relay/commit/e710be96c3359c4a9a3c59a366d5b487d2a8fe29))
- Enhance event dispatcher for real-time broadcasting and improve changefeed capabilities ([f408d87](https://github.com/Shugur-Network/relay/commit/f408d87002786f5b99c4596abe9f505b01c6065d))
- Enhance installation script to support interactive, direct, and piped modes for domain input ([b674c6d](https://github.com/Shugur-Network/relay/commit/b674c6dc5d9447d31b65deb9b6f1e2d8f210c518))
- Enhance metrics tracking and add real-time metrics API ([37baf98](https://github.com/Shugur-Network/relay/commit/37baf98cb2f97cceb6bbd22cbd8601c32feab564))
- Enhance NIP-28 validation and testing for public chat events ([106b0d9](https://github.com/Shugur-Network/relay/commit/106b0d99188960309e94241e6c8b8beb761bffad))
- Enhance NIP-65 validation and testing for relay list metadata events ([9b868b2](https://github.com/Shugur-Network/relay/commit/9b868b25f30be50a4e6bb233a79d8c0c86d53208))
- Implement configurable content length limits for relay metadata and WebSocket connections ([94abbeb](https://github.com/Shugur-Network/relay/commit/94abbeb0887021d58e322afa16fb389af99e7689))
- Implement cross-node event synchronization using polling instead of changefeed ([0dc0380](https://github.com/Shugur-Network/relay/commit/0dc0380e7353c7c2cb6ea9456cfa51dde679d69f))
- Implement NIP-45 COUNT command and associated tests ([ec5c79d](https://github.com/Shugur-Network/relay/commit/ec5c79df24ec0db7bbe0d16c7a80e32e2bc3d421))
- Integrate CockroachDB changefeed for real-time event synchronization across distributed relays ([aa005cd](https://github.com/Shugur-Network/relay/commit/aa005cddb7082d9802f194ee77eeef16542d4008))
- Optimize logging across relay components with proper levels and NIP validation visibility ([a79fa87](https://github.com/Shugur-Network/relay/commit/a79fa8782d58a607a9b559c6eaf7552e4741811b))
- Preserve CA certificates during cleanup for future node additions ([f32c7ed](https://github.com/Shugur-Network/relay/commit/f32c7ed4bc56a9a058277bb2c55d0a497304d247))
- Skip storage of ephemeral events and enhance broadcasting logic ([408ce68](https://github.com/Shugur-Network/relay/commit/408ce686dc3cf7e97147a130d995033699c97666))
- Update relay list event validation to use specific kind validation ([9fa3259](https://github.com/Shugur-Network/relay/commit/9fa3259b6bcc879239271ed360083714646257d0))
- Simplify release workflow with direct version input ([e76fe3b](https://github.com/Shugur-Network/Relay/commit/e76fe3b639874768d6d2f830ce1403a0a673bb1e))
- Simplify release workflow with direct version input (#69) ([ff1316a](https://github.com/Shugur-Network/Relay/commit/ff1316a767697b6c46edf305b9ea08f44a0baad8))

### Bug Fixes

- **ci**: Infrastructure (#72) ([b25bdb0](https://github.com/Shugur-Network/Relay/commit/b25bdb0))
- Include both compressed and uncompressed binaries in releases ([90a772a](https://github.com/Shugur-Network/Relay/commit/90a772a83652f725cf87f526cd102bfa223954c1))
- **ci**: Optimize ci infrastructure (#70) ([dfe700f](https://github.com/Shugur-Network/Relay/commit/dfe700f3959eec7ea8c74e51133587be3c6b03af))
- **ci**: Release modifications (#71) ([c62002b](https://github.com/Shugur-Network/Relay/commit/c62002be1333854463c2951d2539cd34d463e4a2))
- Adjust certificate ownership for relay and cockroach certs to ensure proper access ([aded889](https://github.com/Shugur-Network/relay/commit/aded889763c7b7303b41c7b6e1450966e5cd4a47))
- Correct delegation logging to use struct field instead of slice indexing ([5d7d3cd](https://github.com/Shugur-Network/relay/commit/5d7d3cda6b1836252f48301b474dc7c12a2d5eec))
- Enhance cleanup process and add port availability checks in installation script ([2460bbb](https://github.com/Shugur-Network/relay/commit/2460bbb543c219e56fb7d0ba33ba097ba3a10504))
- Extract real client IPs from proxy headers (v1.3.2.1) ([0cd278c](https://github.com/Shugur-Network/relay/commit/0cd278c8486bfff6cfff55d912ec51fdceb1f6ea))
- Revert pgx/v5 from 5.7.6 to 5.7.4 to resolve query timeout issues ([#37](https://github.com/Shugur-Network/relay/issues/37)) ([c8011f9](https://github.com/Shugur-Network/relay/commit/c8011f9aa5b940ff9a3a95a525cec75b0cb236e1))
- Update repository references from 'Relay' to 'relay' in various files ([3477d96](https://github.com/Shugur-Network/relay/commit/3477d962cb33b62d5fdb247c40bf7c838ba2390c))
- Update versioning prefix in CI configuration for consistency ([31a957d](https://github.com/Shugur-Network/relay/commit/31a957d72d1f9c91ff0c0596bc34fc0028129ccf))

### Documentation

- Enhance README with comprehensive improvements (#55) ([de38fac](https://github.com/Shugur-Network/Relay/commit/de38fac))
- Enhance contributor guidelines and optimize release workflow (#49) ([db281b6](https://github.com/Shugur-Network/Relay/commit/db281b6619968c2c6c01d446f8f4abd362c6bad0))
- Enhance README (#56) ([c2cdfd1](https://github.com/Shugur-Network/Relay/commit/c2cdfd16233b80953600f4c0d55b5bac0917bde0))
- Update README and test documentation with comprehensive NIP support (35+ NIPs)
- Add changelog entries for new NIP implementations

### Configuration

- Update development setup and port configurations (#52) ([ff68727](https://github.com/Shugur-Network/Relay/commit/ff68727))

### Continuous Integration

- Optimize workflows for development efficiency (#54) ([dcafa23](https://github.com/Shugur-Network/Relay/commit/dcafa23d90c66ad406d6afd2b69741a8bb8ed575))
- Fix multi-env-testing workflow syntax errors (#57) ([d357ca9](https://github.com/Shugur-Network/Relay/commit/d357ca9688d1d8c3501fe7c3a6d04441f6b17394))
- Fix dependabot fetch-metadata errors and release workflow automation (#66) ([d292be6](https://github.com/Shugur-Network/relay/commit/d292be6e94201ea30d19ff3bbf6fd6465bd579ee))
- Fix dependabot fetch-metadata errors and release workflow automation (#67) ([3870e76](https://github.com/Shugur-Network/relay/commit/3870e7662da3c68e362a8b69d49a3a6430dfe6b2))

### Dependencies

- **deps**: Bump the actions-minor-patch group across 1 directory with 10 updates (#64) ([6198672](https://github.com/Shugur-Network/Relay/commit/619867280fc105a5341d131530ef9554fef2041c))
- **deps**: Bump actions/setup-go from 5.0.2 to 6.0.0 (#59) ([5048a00](https://github.com/Shugur-Network/Relay/commit/5048a0037ffb34841cec5805f3f299f8ce4a4531))

### Miscellaneous Tasks

- Update actions/checkout and actions/setup-go versions in CI workflows (#63) ([ecccedd](https://github.com/Shugur-Network/relay/commit/ecccedd3bd29bf975985e79f409c994450d316d9))
- Remove deprecated GitHub workflows for enhanced release notes and PR status comments ([4ffb59f](https://github.com/Shugur-Network/Relay/commit/4ffb59f50f9d1deb054bddc31e4b36efd6accd56))
- Remove release-please configuration files ([00827f5](https://github.com/Shugur-Network/Relay/commit/00827f5912d575a0d75ac599b66b2e9c3186f309))

### Performance Improvements

- Optimize cockroachdb schema with zstd compression and enhanced indexes (#61) ([6542d13](https://github.com/Shugur-Network/relay/commit/6542d13766f0c444120dedc573d8302253f44db8))
- Set fixed preallocation for query results to 500 (matches typical filter cap) ([#20](https://github.com/Shugur-Network/relay/issues/20)) ([a6fb50e](https://github.com/Shugur-Network/relay/commit/a6fb50e8e7be6b689ab2b99317aa77d5b2059f06))

## [1.3.3-rc.3] - 2025-09-25

### Features

- Production improvements with comprehensive error handling and optimization
- Comprehensive lean release system with full CI/CD pipeline
- Simplify release workflow with direct version input

### Bug Fixes

- **ci**: Infrastructure improvements
- Include both compressed and uncompressed binaries in releases
- **ci**: Optimize ci infrastructure
- **ci**: Release modifications

### Configuration

- Update development setup and port configurations

### Continuous Integration

- Optimize workflows for development efficiency
- Fix multi-env-testing workflow syntax errors
- Fix dependabot fetch-metadata errors and release workflow automation

### Dependencies

- **deps**: Bump the actions-minor-patch group across 1 directory with 10 updates
- **deps**: Bump actions/setup-go from 5.0.2 to 6.0.0

### Documentation

- Enhance README with comprehensive improvements
- Enhance contributor guidelines and optimize release workflow

### Miscellaneous Tasks

- Update actions/checkout and actions/setup-go versions in CI workflows
- Remove deprecated GitHub workflows for enhanced release notes and PR status comments
- Remove release-please configuration files

### Performance Improvements

- Optimize cockroachdb schema with zstd compression and enhanced indexes

## [1.3.3-rc.2] - 2025-09-15

### Features

- Add MaxConnections field to LimitationData and update related templates
- Enhance event dispatcher for real-time broadcasting and improve changefeed capabilities
- Enhance installation script to support interactive, direct, and piped modes for domain input
- Enhance metrics tracking and add real-time metrics API
- Enhance NIP-28 validation and testing for public chat events
- Enhance NIP-65 validation and testing for relay list metadata events
- Implement configurable content length limits for relay metadata and WebSocket connections
- Implement cross-node event synchronization using polling instead of changefeed
- Implement NIP-45 COUNT command and associated tests
- Integrate CockroachDB changefeed for real-time event synchronization across distributed relays
- Optimize logging across relay components with proper levels and NIP validation visibility
- Preserve CA certificates during cleanup for future node additions
- Skip storage of ephemeral events and enhance broadcasting logic
- Update relay list event validation to use specific kind validation

### Bug Fixes

- Adjust certificate ownership for relay and cockroach certs to ensure proper access
- Correct delegation logging to use struct field instead of slice indexing
- Enhance cleanup process and add port availability checks in installation script
- Extract real client IPs from proxy headers (v1.3.2.1)
- Update repository references from 'Relay' to 'relay' in various files
- Update versioning prefix in CI configuration for consistency

### Performance Improvements

- Set fixed preallocation for query results to 500 (matches typical filter cap)

## [1.3.3-rc.1] - 2025-09-15

### Bug Fixes

- Revert pgx/v5 from 5.7.6 to 5.7.4 to resolve query timeout issues

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
  - Relay identity display in NIP-11 relay information document

- **Advanced Configuration System**:
  - Enhanced configuration validation with comprehensive error handling
  - Improved default configuration structure and organization
  - Added configuration hot-reloading capabilities
  - Enhanced environment variable support with proper precedence

- **Database Performance Improvements**:
  - Optimized query performance with enhanced indexing strategies
  - Improved connection pooling and resource management
  - Enhanced event storage and retrieval mechanisms
  - Database migration system for schema updates

### Changed

- **Relay Architecture Enhancements**:
  - Improved WebSocket connection handling with better resource management
  - Enhanced event processing pipeline for better throughput
  - Optimized memory usage and garbage collection performance
  - Improved error handling and logging throughout the application

- **Security Improvements**:
  - Enhanced input validation and sanitization
  - Improved rate limiting and DoS protection
  - Better authentication and authorization mechanisms
  - Enhanced cryptographic operations and key management

### Fixed

- **Connection Stability**: Resolved WebSocket connection stability issues
- **Memory Leaks**: Fixed memory leaks in event processing and connection handling
- **Race Conditions**: Eliminated race conditions in concurrent operations
- **Event Validation**: Enhanced event validation logic for better compliance
- **Database Queries**: Optimized database queries for better performance

## [1.2.1] - 2025-08-15

### Fixed

- **Database Connection**: Fixed database connection pooling issues
- **WebSocket Handling**: Improved WebSocket message processing
- **Memory Management**: Resolved memory usage optimization

### Security

- Enhanced input validation for all endpoints
- Improved rate limiting mechanisms

## [1.2.0] - 2025-08-01

### Added

- **NIP-65 Support**: Relay List Metadata implementation
- **Enhanced Metrics**: Comprehensive relay metrics and monitoring
- **Configuration Improvements**: Better configuration management and validation

### Changed

- **Performance Optimizations**: Improved query performance and connection handling
- **Logging Enhancements**: Better structured logging with correlation IDs

### Fixed

- **Event Processing**: Resolved event processing bottlenecks
- **WebSocket Stability**: Improved WebSocket connection reliability

## [1.1.0] - 2025-07-15

### Added

- **Core NIP Support**: Implemented NIPs 1, 2, 3, 4, 9, 11, 15, 16, 17, 20, 22, 23, 25, 28, 33, 40, 44
- **Database Integration**: CockroachDB support with connection pooling
- **WebSocket Server**: Full WebSocket relay implementation
- **Configuration System**: YAML-based configuration with validation

### Changed

- **Architecture**: Modular architecture with clear separation of concerns
- **Build System**: Enhanced build process with version management

### Fixed

- **Initial Release**: Baseline implementation with core functionality

## [1.0.0] - 2025-07-01

### Added

- **Initial Release**: Basic Nostr relay implementation
- **Core Features**: Event storage, retrieval, and WebSocket communication
- **Documentation**: Initial project documentation and setup guides

---

For the complete version history and detailed release notes, see the [GitHub Releases](https://github.com/Shugur-Network/relay/releases) page.