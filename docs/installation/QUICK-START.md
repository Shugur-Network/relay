# Quick Start (Recommended)

This is the fastest way to get a Shugur Relay running. Choose between standalone (single-node) or distributed (multi-node) installation based on your needs.

## Option 1: Standalone Installation (Single Node)

Perfect for **development, testing, or small-scale production** environments. Sets up a complete relay on a single server.

### Steps

1. **Run the Installer**

    Execute the installation script with `curl`. It handles Docker installation, database setup, and service deployment automatically.

    ```bash
    curl -fsSL https://github.com/Shugur-Network/relay/raw/main/scripts/install.standalone.sh | sudo bash
    ```

2. **Follow the Prompts**

    The script will guide you through basic configuration options like relay name and settings.

The installation is complete when you see the relay running and accessible via web interface.

## Option 2: Distributed Installation (Multi-Node)

Recommended for **production-ready, high-availability** setups. Deploys a distributed relay cluster across multiple servers.

### How It Works

You run a single script from your local machine, which then connects to your servers via SSH to perform the installation.

### Steps

1. **Run the Installer**

    Execute the installation script with `curl`. It needs root privileges to use `sshpass` for password-based SSH authentication to your servers.

    ```bash
    curl -fsSL https://github.com/Shugur-Network/relay/raw/main/scripts/install.distributed.sh | sudo bash
    ```

2. **Follow the Prompts**

    The script will guide you through the following configuration:
    - **SSH Credentials**: For connecting to your remote servers.
    - **Node Information**: The number of servers in your distributed setup and their IP addresses/hostnames.

The script will then take over, and you can monitor the progress as it sets up your distributed relay cluster.

## Next Steps

For detailed production configuration and operational guidance, see:

- [Configuration Guide](../CONFIGURATION.md)
- [Architecture Overview](../ARCHITECTURE.md)
- [Concepts](../CONCEPTS.md)
- [API Reference](../API.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Performance Guide](../PERFORMANCE.md)
