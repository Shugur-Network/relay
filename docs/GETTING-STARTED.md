# Getting Started

This guide covers the prerequisites for installing Shugur Relay. Ensure your system meets these requirements before proceeding with the installation.

## System Requirements

These are the minimum requirements to run Shugur Relay. Production environments may require more resources based on usage.

> **💡 Tip**: For production deployments, we recommend doubling these minimum requirements for better performance and reliability.

| Resource  | Standalone          | Distributed (per node) |
| :-------- | :------------------ | :--------------------- |
| **CPU**   | 2 Cores (minimum)   | 4 Cores (minimum)      |
| **RAM**   | 4 GB (minimum)      | 8 GB (minimum)         |
| **Storage** | 20 GB SSD (minimum) | 50 GB SSD (minimum)    |

## Software Prerequisites

### For Docker Installations (Recommended)

- **Docker**: Version `20.10` or newer.
- **Docker Compose**: Version `2.0` or newer.

You can install both with a single command on most Linux systems:

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```

### For Bare Metal Installations

- **Go**: Version `1.24` or newer.
- **CockroachDB**: Version `v24.1.x`.
- **Git**, **Curl**, **Wget**, **OpenSSL**.

## Network Requirements

The following ports must be accessible on your server(s).

| Port    | Protocol | Description                               | Required For      |
| :------ | :------- | :---------------------------------------- | :---------------- |
| `80/443`  | TCP      | HTTP/HTTPS for Caddy reverse proxy        | All               |
| `8080`  | TCP      | Shugur Relay WebSocket and API            | All               |
| `8181`  | TCP      | Prometheus metrics endpoint               | All               |
| `9090`  | TCP      | CockroachDB Admin UI                      | All               |
| `26257` | TCP      | CockroachDB SQL client connections        | All               |
| `26258` | TCP      | CockroachDB inter-node communication      | Distributed Only  |
| `22`    | TCP      | SSH for installation script               | Distributed Only  |

### Firewall Configuration Example (UFW)

```bash
# For Standalone
sudo ufw allow 22,80,443,8080,8181,9090,26257/tcp

# For Distributed (run on each node)
sudo ufw allow 22,80,443,8080,8181,9090,26257,26258/tcp

sudo ufw enable
```

## Next Steps

Now that you have reviewed the prerequisites, you are ready to install Shugur Relay.

➡️ **Proceed to the [Installation Guide](./installation/INSTALLATION.md)**

## Related Documentation

- **[Installation Guide](./installation/INSTALLATION.md)**: Choose your installation method
- **[Configuration Guide](./CONFIGURATION.md)**: Configure your relay after installation
- **[Architecture Overview](./ARCHITECTURE.md)**: Understand how the system works
- **[Troubleshooting](./TROUBLESHOOTING.md)**: Common issues and solutions
