# Quick Start (Recommended)

This is the fastest way to get a **distributed, production-ready** Shugur Relay running. The automated script handles everything from Docker installation to certificate generation and service deployment across multiple nodes.

## How It Works

You run a single script from your local machine, which then connects to your servers via SSH to perform the installation.

## Steps

1. **Run the Installer**

    Execute the installation script with `curl`. It needs root privileges to use `sshpass` for password-based SSH authentication to your servers.

    ```bash
    curl -fsSL https://raw.githubusercontent.com/Shugur-Network/Relay/main/scripts/install.distributed.sh | sudo bash
    ```

2. **Follow the Prompts**

    Execute the script with `sudo`. It needs root privileges to use `sshpass` for password-based SSH authentication to your servers.

    ```bash
    sudo ./install.distributed.sh
    ```

3. **Follow the Prompts**

    The script will guide you through the following configuration:
    - **SSH Credentials**: For connecting to your remote servers.
    - **Node Information**: The number of servers in your distributed setup and their IP addresses/hostnames.

The script will then take over, and you can monitor the progress as it sets up your distributed relay.

## Next Steps

For detailed production configuration and operational guidance, see:

- [Configuration Guide](../CONFIGURATION.md)
- [Architecture Overview](../ARCHITECTURE.md)
- [Concepts](../CONCEPTS.md)
- [API Reference](../API.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Performance Guide](../PERFORMANCE.md)
