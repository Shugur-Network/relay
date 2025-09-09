# NIP-XX Time Capsules Python Implementation

This directory contains the **complete Python implementation and test suite** for **NIP-XX Time Capsules** - a Nostr protocol for time-locked encrypted messages using drand beacons.

## 📁 Directory Contents

### 🐍 Python Implementation

- `complete_time_capsules_demo.py` - **Complete NIP-XX compliant implementation** with full specification adherence
- `test_nip_xx_time_capsules.sh` - Test runner script with dependency validation
- `requirements-test.txt` - Python dependencies

## 🚀 Quick Start

### Prerequisites

- Local Nostr relay running on `ws://localhost:8085`
- Required tools: `nak`, `tle`, `jq`, `base64`, `od`, `python3`

### Install Dependencies

#### Go Tools

```bash
go install github.com/drand/tlock/cmd/tle@latest
go install github.com/fiatjaf/nak@latest
```

#### Python Dependencies

```bash
# Option 1: Use virtual environment (recommended)
python3 -m venv venv
source venv/bin/activate
pip3 install -r requirements-test.txt

# Option 2: Install globally (if permitted)
pip3 install websocket-client requests
```

### Run Tests

#### Python Test Runner (Recommended)

```bash
./test_nip_xx_time_capsules.sh
```

#### Python Implementation Directly

```bash
python3 complete_time_capsules_demo.py
```

## 🎯 What Gets Tested

### ✅ Complete NIP-XX Compliance

- **Event Structure**: Kind 1041, proper tags, content encoding
- **Payload Formats**: Public (0x01) and Private (0x02) modes
- **Encryption**: tlock, NIP-44 v2 alignment, HKDF-SHA256, ChaCha20
- **Authentication**: HMAC-SHA256, constant-time verification
- **Tag Parsing**: "key value" format, drand_chain/drand_round parameters

### ✅ Cross-Chain Compatibility

- **League of Entropy Mainnet** (api.drand.sh)
- **Cloudflare Drand Mirror** (drand.cloudflare.com)
- **Network Resilience**: Automatic failover and endpoint testing

### ✅ Real-World Integration

- **Relay Communication**: WebSocket posting and querying
- **Timelock Mechanics**: Actual drand round waiting and verification
- **End-to-End Flow**: Create → Post → Wait → Decrypt validation

## 🏆 Test Results

When tests pass, you'll see:

- ✅ **4/4 messages created** (2 public + 2 private)
- ✅ **4/4 perfect decryptions** (byte-for-byte accuracy)
- ✅ **Cross-chain interoperability** demonstrated
- ✅ **Full NIP-XX specification compliance** verified

## 🔧 Development

### Adding New Features

1. Update `complete_time_capsules_demo.py` for new functionality
2. Add test cases for new cryptographic scenarios
3. Document new features in the appropriate XX-*.md files
4. Ensure cross-chain compatibility is maintained

### Debugging

- Check relay logs: `./bin/relay start --config config/development.yaml`
- Verify drand connectivity: `curl https://api.drand.sh/52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971/public/latest`
- Test dependencies: `tle -h && nak -h`

## 📖 Specification Reference

See `XX.md` for the complete NIP-XX specification including:

- Event format requirements
- Cryptographic primitives
- Validation rules
- Implementation guidelines
- Security considerations

## 🤝 Contributing

When contributing to NIP-XX Time Capsules:

1. Ensure all tests pass
2. Update documentation for any protocol changes
3. Test cross-chain compatibility
4. Verify specification compliance
