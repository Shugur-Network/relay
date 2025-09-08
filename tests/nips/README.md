# Nostr Implementation Protocol (NIP) Test Suite

This directory contains comprehensive test scripts for various Nostr Implementation Protocols (NIPs). Each test validates specific functionality and ensures compliance with the respective NIP specifications.

## 🚀 Quick Start

### Prerequisites
- Local Nostr relay running on `ws://localhost:8085`
- Required tools: `nak`, `jq`, `base64`, `od`, `python3`
- Bash shell environment

### Running Tests
```bash

# Run a specific test
./tests/nips/test_nip01.sh

# Run all tests (if available)
for test in tests/nips/test_nip*.sh; do
    echo "Running $test..."
    bash "$test"
done
```

## 📋 Available Tests

### Core NIPs

| Test File | NIP | Description | Status |
|-----------|-----|-------------|---------|
| `test_nip01.sh` | NIP-01 | Basic protocol structure and event format | ✅ |
| `test_nip02.sh` | NIP-02 | Contact lists and petnames | ✅ |
| `test_nip03.sh` | NIP-03 | OpenTimestamps attestations | ✅ |
| `test_nip04.sh` | NIP-04 | Encrypted direct messages | ✅ |

### Authentication & Security

| Test File | NIP | Description | Status |
|-----------|-----|-------------|---------|
| `test_nip09.sh` | NIP-09 | Event deletion | ✅ |
| `test_nip11.sh` | NIP-11 | Relay information document | ✅ |
| `test_nip15.sh` | NIP-15 | Nostr marketplace | ✅ |
| `test_nip16.sh` | NIP-16 | Event treatment | ✅ |
| `test_nip17.sh` | NIP-17 | Reposts | ✅ |

### Advanced Features

| Test File | NIP | Description | Status |
|-----------|-----|-------------|---------|
| `test_nip20.sh` | NIP-20 | Command results | ✅ |
| `test_nip22.sh` | NIP-22 | Event `created_at` limits | ✅ |
| `test_nip23.sh` | NIP-23 | Long-form content | ✅ |
| `test_nip25.sh` | NIP-25 | Reactions | ✅ |
| `test_nip28.sh` | NIP-28 | Public chat | ✅ |

### Encryption & Privacy

| Test File | NIP | Description | Status |
|-----------|-----|-------------|---------|
| `test_nip33.sh` | NIP-33 | Parameterized replaceable events | ✅ |
| `test_nip40.sh` | NIP-40 | Expiration timestamps | ✅ |
| `test_nip44.sh` | NIP-44 | Encrypted payloads | ✅ |
| `test_nip45.sh` | NIP-45 | Counting results | ✅ |
| `test_nip50.sh` | NIP-50 | Keywords filter | ✅ |
| `test_nip59.sh` | NIP-59 | Gift wrap events | ✅ |
| `test_nip65.sh` | NIP-65 | Relay list metadata | ✅ |

### Specialized Features

| Test File | NIP | Description | Status |
|-----------|-----|-------------|---------|
| `test_nip78.sh` | NIP-78 | Application-specific data | ✅ |
| `test_nip_xx_time_capsules.sh` | NIP-XX | Time-lock encrypted messages | ✅ |

## 🔧 Test Configuration

### Environment Variables
```bash
# Relay URL (default: ws://localhost:8085)
export RELAY_URL="ws://localhost:8085"

# Test timeout (default: 30 seconds)
export TEST_TIMEOUT=30

# Verbose output
export VERBOSE=1
```

### Common Test Patterns

Most tests follow this structure:
1. **Setup**: Generate test keys and data
2. **Create**: Generate events according to NIP spec
3. **Publish**: Send events to relay
4. **Verify**: Validate event structure and content
5. **Cleanup**: Remove test data (if applicable)

## 📊 Test Results

### Success Indicators
- ✅ All test cases pass
- ✅ Events published successfully to relay
- ✅ Event structure matches NIP specification
- ✅ Content validation successful

### Common Issues
- ❌ **Relay not running**: Ensure `ws://localhost:8085` is accessible
- ❌ **Missing dependencies**: Install `nak`, `jq`, `base64`, `od`, `python3`
- ❌ **Permission denied**: Make scripts executable with `chmod +x`
- ❌ **Invalid event format**: Check NIP specification compliance

## 🎯 Specialized Tests

### NIP-XX Time Capsules (`test_nip_xx_time_capsules.sh`)
**Purpose**: Tests time-lock encrypted messages that can only be decrypted after a specific time.

**Features**:
- Public time capsules (mode 0x01)
- Private time capsules (mode 0x02) 
- Gift-wrapped private capsules (NIP-59 integration)
- Actual timelock enforcement with waiting
- Drand integration for time-lock mechanism

**Usage**:
```bash
# Run the time capsule test
./tests/nips/test_nip_xx_time_capsules.sh

# Expected output: 2 events created
# 1. Public time capsule (kind 1041)
# 2. Gift-wrapped private capsule (kind 1059)
```

### NIP-44 Encryption (`test_nip44.sh`)
**Purpose**: Tests encrypted payloads using shared secrets.

**Features**:
- Key generation and derivation
- Message encryption/decryption
- Authentication and integrity verification

### NIP-59 Gift Wrapping (`test_nip59.sh`)
**Purpose**: Tests metadata privacy through ephemeral keys.

**Features**:
- Ephemeral key generation
- Event wrapping and unwrapping
- Recipient-specific encryption

## 🔍 Debugging Tests

### Enable Verbose Output
```bash
# Run with debug information
VERBOSE=1 ./tests/nips/test_nip01.sh

# Run with bash debug mode
bash -x ./tests/nips/test_nip01.sh
```

### Check Relay Status
```bash
# Test relay connectivity
curl -s http://localhost:8085/ | jq .

# Check relay info
nak relay info ws://localhost:8085
```

### Validate Event Format
```bash
# Check event structure
echo '{"kind":1,"content":"test"}' | jq .

# Validate against NIP spec
nak event validate < event.json
```

## 📚 NIP Documentation

For detailed specifications, refer to:
- [NIP Repository](https://github.com/nostr-protocol/nips)
- [Nostr Protocol Website](https://nostr.com)
- [NIP-01: Basic Protocol](https://github.com/nostr-protocol/nips/blob/master/01.md)

## 🤝 Contributing

When adding new tests:
1. Follow the existing naming convention: `test_nip##.sh`
2. Include comprehensive error handling
3. Add clear success/failure indicators
4. Document any special requirements
5. Test against multiple relay implementations

## 📝 Notes

- All tests are designed to work with the local relay at `ws://localhost:8085`
- Tests use temporary keys and data - no permanent data is created
- Some tests may require specific relay features or configurations
- Time-based tests (like NIP-XX) may take longer to complete due to waiting periods

For questions or issues, please refer to the individual test files or the NIP specifications.