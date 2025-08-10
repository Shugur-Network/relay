# Architecture

Shugur Relay is designed for high availability, scalability, and operational simplicity. This is achieved through a stateless architecture and the use of a distributed SQL database.

## Core Components

```text
                               +-------------------------+
                               |   CockroachDB           |
                               | (Distributed Database)  |
                               +-----------+-------------+
                                           ^
                                           | SQL
                                           v
+----------------+      +------------------+------------------+
|    Clients     |----->|      Shugur Relay Nodes (Stateless)   |
| (Nostr Users)  |<-----| (Go Application)                    |
+----------------+      +------------------+------------------+
```

### Shugur Relay Nodes

- **Stateless Go Application**: The core relay logic is written in Go. The nodes themselves do not store any persistent state, such as event data or subscriptions. This allows them to be scaled, restarted, or replaced without data loss.
- **WebSocket Handling**: Each node manages WebSocket connections from Nostr clients, handling incoming events and outgoing subscription data.
- **API & Dashboard**: A web interface is served directly from the relay for administration.

### CockroachDB (Distributed Database)

- **Single Source of Truth**: All event data, policies, and other persistent information are stored in CockroachDB.
- **High Availability**: CockroachDB is a distributed database that replicates data across multiple nodes. If one database node fails, the system remains operational.
- **Scalability**: Both the relay nodes and the database can be scaled independently. If you have high traffic, you can add more relay nodes. If your data storage needs grow, you can add more CockroachDB nodes.

## Standalone vs. Distributed

> **💡 Architecture Tip**: Choose your deployment model based on your availability requirements and operational complexity tolerance. Start simple and scale up as needed.

### Standalone Deployment

- A single Shugur Relay node connects to a single-node CockroachDB instance.
- **Pros**: Simple to set up and manage.
- **Cons**: Not highly available. If the server or database goes down, the relay is offline.

### Distributed Deployment

- Multiple Shugur Relay nodes (often behind a load balancer) connect to a multi-node CockroachDB distributed database.
- **Pros**: Highly available and horizontally scalable. Tolerant to node failures.
- **Cons**: More complex to set up.

## Next Steps

- **[Installation Guide](./installation/INSTALLATION.md)**: Now that you understand the architecture, proceed to the installation guide to deploy the relay.
- **[Nostr Concepts](./CONCEPTS.md)**: Learn more about the Nostr protocol itself.

## Related Documentation

- **[Installation Guide](./installation/INSTALLATION.md)**: Choose your deployment method
- **[Configuration Guide](./CONFIGURATION.md)**: Configure your relay settings
- **[Performance Guide](./PERFORMANCE.md)**: Optimize for production workloads
- **[Troubleshooting Guide](./TROUBLESHOOTING.md)**: Resolve common issues
- **[API Reference](./API.md)**: WebSocket and HTTP endpoint documentation
