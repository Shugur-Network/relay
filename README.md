<div align="center">
  <a href="https://shugur.com">
    <img src="https://github.com/Shugur-Network/relay/raw/main/banner.png" alt="Shugur Relay Banner" width="100%">
  </a>
  <p align="center">
    High-performance, reliable, and scalable Nostr relay.
  </p>
  
  <!-- Status Badges -->
  <p align="center">
    <a href="https://github.com/Shugur-Network/relay/actions/workflows/ci.yml">
      <img src="https://github.com/Shugur-Network/relay/actions/workflows/ci.yml/badge.svg" alt="CI Status">
    </a>
    <a href="https://github.com/Shugur-Network/relay/releases">
      <img src="https://img.shields.io/github/v/release/Shugur-Network/relay?include_prereleases" alt="Release">
    </a>
    <a href="https://github.com/Shugur-Network/relay/blob/main/LICENSE">
      <img src="https://img.shields.io/github/license/Shugur-Network/relay" alt="License">
    </a>
    <a href="https://goreportcard.com/report/github.com/Shugur-Network/relay">
      <img src="https://goreportcard.com/badge/github.com/Shugur-Network/relay" alt="Go Report Card">
    </a>
  </p>
  
  <!-- Technology Badges -->
  <p align="center">
    <a href="https://golang.org/">
      <img src="https://img.shields.io/badge/Go-1.24.4+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go Version">
    </a>
    <a href="https://www.cockroachlabs.com/">
      <img src="https://img.shields.io/badge/CockroachDB-v24.1.5+-6933FF?style=flat&logo=cockroachlabs&logoColor=white" alt="CockroachDB">
    </a>
    <a href="https://nostr.com/">
      <img src="https://img.shields.io/badge/Nostr-Protocol-8B5CF6?style=flat&logo=lightning&logoColor=white" alt="Nostr Protocol">
    </a>
    <a href="https://github.com/Shugur-Network/relay/pkgs/container/relay">
      <img src="https://img.shields.io/badge/Docker-Available-2496ED?style=flat&logo=docker&logoColor=white" alt="Docker">
    </a>
  </p>
  
  <!-- Quality & Activity Badges -->
  <p align="center">
    <a href="https://github.com/Shugur-Network/relay/commits/main">
      <img src="https://img.shields.io/github/commit-activity/m/Shugur-Network/relay" alt="Commit Activity">
    </a>
    <a href="https://github.com/Shugur-Network/relay">
      <img src="https://img.shields.io/github/repo-size/Shugur-Network/relay" alt="Repository Size">
    </a>
    <a href="https://github.com/Shugur-Network/relay">
      <img src="https://img.shields.io/github/languages/top/Shugur-Network/relay" alt="Top Language">
    </a>
    <a href="https://github.com/Shugur-Network/relay/commits/main">
      <img src="https://img.shields.io/github/last-commit/Shugur-Network/relay" alt="Last Commit">
    </a>
  </p>
</div>

  <!-- Community & Stats Badges -->
  <p align="center">
    <a href="https://github.com/Shugur-Network/relay/issues">
      <img src="https://img.shields.io/github/issues/Shugur-Network/relay" alt="Issues">
    </a>
    <a href="https://github.com/Shugur-Network/relay/pulls">
      <img src="https://img.shields.io/github/issues-pr/Shugur-Network/relay" alt="Pull Requests">
    </a>
    <a href="https://github.com/Shugur-Network/relay/stargazers">
      <img src="https://img.shields.io/github/stars/Shugur-Network/relay?style=social" alt="GitHub Stars">
    </a>
    <a href="https://github.com/Shugur-Network/relay/network/members">
      <img src="https://img.shields.io/github/forks/Shugur-Network/relay?style=social" alt="GitHub Forks">
    </a>
  </p>

---

Shugur Relay is a production-ready Nostr relay built in Go with CockroachDB for distributed storage. It's designed for operators who need reliability, observability, and horizontal scale.

## What is Nostr?

Nostr (Notes and Other Stuff Transmitted by Relays) is a simple, open protocol that enables a truly censorship-resistant and global social network. Unlike traditional social media platforms, Nostr doesn't rely on a central server. Instead, it uses a network of relays (like Shugur Relay) to store and transmit messages, giving users complete control over their data and communications.

Key benefits of Nostr:

- **Censorship Resistance**: No single point of control or failure
- **Data Ownership**: Users control their own data and identity
- **Interoperability**: Works across different clients and applications
- **Simplicity**: Lightweight protocol that's easy to implement and understand

Learn more in our [Nostr Concepts](https://github.com/Shugur-Network/docs/blob/main/CONCEPTS.md) documentation.

## 📋 Nostr Protocol Support

### Supported NIPs (Nostr Improvement Proposals)

Shugur Relay implements the following NIPs for maximum compatibility with Nostr clients:

#### Core Protocol

- **[NIP-01](https://github.com/nostr-protocol/nips/blob/master/01.md)**: Basic protocol flow description
- **[NIP-02](https://github.com/nostr-protocol/nips/blob/master/02.md)**: Contact List and Petnames
- **[NIP-03](https://github.com/nostr-protocol/nips/blob/master/03.md)**: OpenTimestamps Attestations for Events
- **[NIP-04](https://github.com/nostr-protocol/nips/blob/master/04.md)**: Encrypted Direct Message
- **[NIP-09](https://github.com/nostr-protocol/nips/blob/master/09.md)**: Event Deletion
- **[NIP-11](https://github.com/nostr-protocol/nips/blob/master/11.md)**: Relay Information Document

#### Enhanced Features

- **[NIP-15](https://github.com/nostr-protocol/nips/blob/master/15.md)**: End of Stored Events Notice
- **[NIP-16](https://github.com/nostr-protocol/nips/blob/master/16.md)**: Event Treatment
- **[NIP-17](https://github.com/nostr-protocol/nips/blob/master/17.md)**: Private Direct Messages
- **[NIP-20](https://github.com/nostr-protocol/nips/blob/master/20.md)**: Command Results
- **[NIP-22](https://github.com/nostr-protocol/nips/blob/master/22.md)**: Event `created_at` Limits
- **[NIP-23](https://github.com/nostr-protocol/nips/blob/master/23.md)**: Long-form Content
- **[NIP-24](https://github.com/nostr-protocol/nips/blob/master/24.md)**: Extra metadata fields and tags
- **[NIP-25](https://github.com/nostr-protocol/nips/blob/master/25.md)**: Reactions

#### Advanced Features

- **[NIP-26](https://github.com/nostr-protocol/nips/blob/master/26.md)**: Delegated Event Signing
- **[NIP-28](https://github.com/nostr-protocol/nips/blob/master/28.md)**: Public Chat
- **[NIP-33](https://github.com/nostr-protocol/nips/blob/master/33.md)**: Addressable Events
- **[NIP-40](https://github.com/nostr-protocol/nips/blob/master/40.md)**: Expiration Timestamp
- **[NIP-44](https://github.com/nostr-protocol/nips/blob/master/44.md)**: Encrypted Payloads (Versioned)
- **[NIP-45](https://github.com/nostr-protocol/nips/blob/master/45.md)**: Counting Events
- **[NIP-50](https://github.com/nostr-protocol/nips/blob/master/50.md)**: Search Capability
- **[NIP-59](https://github.com/nostr-protocol/nips/blob/master/59.md)**: Gift Wrap
- **[NIP-65](https://github.com/nostr-protocol/nips/blob/master/65.md)**: Relay List Metadata
- **[NIP-78](https://github.com/nostr-protocol/nips/blob/master/78.md)**: Application-specific data

### Protocol Features

- **WebSocket Connection**: Real-time bidirectional communication
- **Event Validation**: Cryptographic signature verification
- **Subscription Management**: Efficient filtering and real-time updates
- **Rate Limiting**: Protection against spam and abuse
- **Event Storage**: Persistent storage with CockroachDB
- **Search Support**: Full-text search capabilities (NIP-50)
- **Relay Information**: Discoverable relay metadata (NIP-11)

## 🚀 Features

- **Production-Ready**: Built for reliability and performance with enterprise-grade features.
- **Horizontally Scalable**: Stateless architecture allows easy scaling across multiple nodes.
- **Distributed Database**: Uses CockroachDB for high availability and global distribution.
- **Advanced Throttling**: Sophisticated rate limiting and abuse prevention mechanisms.
- **NIP Compliance**: Implements essential Nostr Improvement Proposals (NIPs).
- **Observability**: Built-in metrics, logging, and monitoring capabilities.
- **Easy Deployment**: One-command installation with automated scripts.
- **Configurable**: Extensive configuration options for fine-tuning behavior.

## ⚡ Quick Start

### Distributed Installation (Recommended)

Get a distributed Shugur Relay cluster running with one command:

```bash
curl -fsSL https://github.com/Shugur-Network/relay/raw/main/scripts/install.distributed.sh | sudo bash
```

### Standalone Installation

For a single-node setup:

```bash
curl -fsSL https://github.com/Shugur-Network/relay/raw/main/scripts/install.standalone.sh | sudo bash
```

For manual setup or other installation methods, see our [Installation Guide](https://docs.shugur.com/installation/).

## 🏗️ Build from Source

```bash
# Clone and build
git clone https://github.com/Shugur-Network/Relay.git
cd Relay

# Build the binary
go build -o bin/relay ./cmd

# Run the relay
./bin/relay
```

## 🐳 Docker Quick Start

### Development Environment

```bash
# Clone repository
git clone https://github.com/Shugur-Network/Relay.git
cd Relay

# Start development database
docker-compose -f docker/compose/docker-compose.development.yml up -d

# Run relay
go run ./cmd --config config/development.yaml
```

**Development Ports:**
- **WebSocket**: `ws://localhost:8081`
- **Metrics**: `http://localhost:8182/metrics`
- **Database Admin**: `http://localhost:9091`

### Production Environment

```bash
# Using official Docker image
docker run -p 8080:8080 ghcr.io/shugur-network/relay:latest

# Or using Docker Compose
docker-compose -f docker/compose/docker-compose.standalone.yml up -d
```

**Production Ports:**
- **WebSocket**: `ws://localhost:8080`
- **Metrics**: `http://localhost:8180/metrics`
- **Database Admin**: `http://localhost:9090`

### Multi-Environment Setup

Run development, testing, and production environments simultaneously:

```bash
# Start all environments
docker-compose -f docker/compose/docker-compose.development.yml up -d
docker-compose -f docker/compose/docker-compose.test.yml up -d
docker-compose -f docker/compose/docker-compose.standalone.yml up -d

# Run relay instances
go run ./cmd --config config/development.yaml &  # Port 8081
go run ./cmd --config config/test.yaml &         # Port 8082
go run ./cmd --config config/production.yaml &   # Port 8080
```

For detailed port mapping, see [config/PORT_MAPPING.md](config/PORT_MAPPING.md).

## 📚 Documentation

Comprehensive documentation is available in our [documentation](https://docs.shugur.com) and [documentation repository](https://github.com/Shugur-Network/docs):

## 🤝 Contributing

We welcome contributions from the community! Please read our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before getting started.

## 🔒 Security

Security is a top priority. If you discover a security vulnerability, please follow our [Security Policy](SECURITY.md) for responsible disclosure.

## License

Shugur Relay is open-source software licensed under the [MIT License](LICENSE).
