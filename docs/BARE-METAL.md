# Bare Metal Installation

This guide provides detailed instructions for installing Shugur Relay directly on a server without using Docker. This method is for advanced users who require maximum performance and control over their environment.

## Prerequisites

### System Requirements

- **Operating System**: Ubuntu 22.04 LTS or newer (other Linux distributions may work)
- **CPU**: 4 cores minimum (8+ cores recommended for production)
- **RAM**: 8 GB minimum (16+ GB recommended for production)
- **Storage**: 50 GB SSD minimum (NVMe recommended for production)

### Software Dependencies

- **Go**: Version 1.21 or newer
- **CockroachDB**: Version v23.1.x or newer
- **Git**, **curl**, **wget**, **openssl**
- **systemd** (for service management)

## Installation Steps

### 1. Install Go

```bash
# Download and install Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### 2. Install CockroachDB

```bash
# Download CockroachDB
cd /tmp
wget https://binaries.cockroachdb.com/cockroach-v23.1.11.linux-amd64.tgz
tar -xzf cockroach-v23.1.11.linux-amd64.tgz

# Install CockroachDB binary
sudo cp cockroach-v23.1.11.linux-amd64/cockroach /usr/local/bin/
sudo chmod +x /usr/local/bin/cockroach

# Verify installation
cockroach version
```

### 3. Create System Users

```bash
# Create CockroachDB user
sudo useradd --system --shell /bin/bash --home /var/lib/cockroach cockroach

# Create Shugur Relay user
sudo useradd --system --shell /bin/bash --home /opt/shugur-relay shugur
```

### 4. Setup CockroachDB

```bash
# Create directories
sudo mkdir -p /var/lib/cockroach/{data,certs}
sudo chown cockroach:cockroach /var/lib/cockroach/{data,certs}

# Generate certificates (for production)
sudo -u cockroach cockroach cert create-ca --certs-dir=/var/lib/cockroach/certs --ca-key=/var/lib/cockroach/certs/ca.key
sudo -u cockroach cockroach cert create-node localhost $(hostname) --certs-dir=/var/lib/cockroach/certs --ca-key=/var/lib/cockroach/certs/ca.key
sudo -u cockroach cockroach cert create-client root --certs-dir=/var/lib/cockroach/certs --ca-key=/var/lib/cockroach/certs/ca.key
sudo -u cockroach cockroach cert create-client relay --certs-dir=/var/lib/cockroach/certs --ca-key=/var/lib/cockroach/certs/ca.key

# Create systemd service
sudo tee /etc/systemd/system/cockroachdb.service > /dev/null <<EOF
[Unit]
Description=CockroachDB
After=network.target

[Service]
Type=exec
User=cockroach
ExecStart=/usr/local/bin/cockroach start --certs-dir=/var/lib/cockroach/certs --store=/var/lib/cockroach/data --listen-addr=localhost:26257 --http-addr=localhost:8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Start and enable CockroachDB
sudo systemctl daemon-reload
sudo systemctl enable cockroachdb
sudo systemctl start cockroachdb

# Initialize the cluster
sudo -u cockroach cockroach init --certs-dir=/var/lib/cockroach/certs --host=localhost:26257
```

### 5. Setup Database

```bash
# Create database and user
sudo -u cockroach cockroach sql --certs-dir=/var/lib/cockroach/certs --host=localhost:26257 <<EOF
CREATE DATABASE IF NOT EXISTS shugur_relay;
CREATE USER IF NOT EXISTS relay;
GRANT ALL ON DATABASE shugur_relay TO relay;
EOF
```

### 6. Build and Install Shugur Relay

```bash
# Clone the repository
cd /tmp
git clone https://github.com/Shugur-Network/Relay.git
cd Relay

# Build the binary
go build -ldflags "-w -s -X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o shugur-relay ./cmd

# Install binary
sudo cp shugur-relay /usr/local/bin/
sudo chmod +x /usr/local/bin/shugur-relay

# Create application directory
sudo mkdir -p /opt/shugur-relay/{config,logs,certs}
sudo chown -R shugur:shugur /opt/shugur-relay

# Copy certificates for relay
sudo cp /var/lib/cockroach/certs/{ca.crt,client.relay.crt,client.relay.key} /opt/shugur-relay/certs/
sudo chown shugur:shugur /opt/shugur-relay/certs/*
sudo chmod 600 /opt/shugur-relay/certs/*.key
```

### 7. Configure Shugur Relay

```bash
# Create configuration file
sudo -u shugur tee /opt/shugur-relay/config/config.yaml > /dev/null <<EOF
GENERAL: {}

LOGGING:
  LEVEL: info
  FILE: /opt/shugur-relay/logs/relay.log
  FORMAT: json
  MAX_SIZE: 100
  MAX_BACKUPS: 5
  MAX_AGE: 30

METRICS:
  ENABLED: true
  PORT: 2112

DATABASE:
  SERVER: localhost
  PORT: 26257
  NAME: shugur_relay
  USER: relay
  SSL_MODE: require

RELAY:
  NAME: "shugur-relay"
  DESCRIPTION: "High-performance, reliable, scalable Nostr relay"
  CONTACT: "admin@example.com"
  ICON: "https://avatars.githubusercontent.com/u/198367099?s=400&u=2bc76d4fe6f57a1c39ef00fd784dd0bf85d79bda&v=4"
  BANNER: "https://github.com/Shugur-Network/Relay/raw/main/banner.png"
  WS_ADDR: ":8080"
  PUBLIC_URL: "ws://localhost:8080"
  EVENT_CACHE_SIZE: 10000
  SEND_BUFFER_SIZE: 8192
  WRITE_TIMEOUT: 60s
  THROTTLING:
    MAX_CONTENT_LENGTH: 8192
    MAX_CONNECTIONS: 1000
    BAN_THRESHOLD: 10
    BAN_DURATION: 300
    RATE_LIMIT:
      ENABLED: true
      MAX_EVENTS_PER_SECOND: 10
      MAX_REQUESTS_PER_SECOND: 20
      BURST_SIZE: 5
      PROGRESSIVE_BAN: true
      MAX_BAN_DURATION: 24h

RELAY_POLICY:
  BLACKLIST:
    PUBKEYS: []
  WHITELIST:
    PUBKEYS: []
EOF
```

### 8. Create Systemd Service

```bash
# Create systemd service file
sudo tee /etc/systemd/system/shugur-relay.service > /dev/null <<EOF
[Unit]
Description=Shugur Relay - Nostr Relay
After=network.target cockroachdb.service
Requires=cockroachdb.service

[Service]
Type=exec
User=shugur
WorkingDirectory=/opt/shugur-relay
ExecStart=/usr/local/bin/shugur-relay --config /opt/shugur-relay/config/config.yaml
Restart=always
RestartSec=5
Environment=HOME=/opt/shugur-relay

[Install]
WantedBy=multi-user.target
EOF

# Start and enable the service
sudo systemctl daemon-reload
sudo systemctl enable shugur-relay
sudo systemctl start shugur-relay
```

### 9. Setup Reverse Proxy (Optional)

For production deployments, set up Nginx or Caddy as a reverse proxy:

```bash
# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Configure Caddy
sudo tee /etc/caddy/Caddyfile > /dev/null <<EOF
your-domain.com {
    reverse_proxy localhost:8080
}
EOF

sudo systemctl reload caddy
```

## Verification

```bash
# Check service status
sudo systemctl status cockroachdb
sudo systemctl status shugur-relay

# View logs
sudo journalctl -u shugur-relay -f

# Test the relay
curl http://localhost:8080/
```

## Maintenance

### Updates

```bash
# Update Shugur Relay
cd /tmp
git clone https://github.com/Shugur-Network/Relay.git
cd Relay
go build -ldflags "-w -s -X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o shugur-relay ./cmd

sudo systemctl stop shugur-relay
sudo cp shugur-relay /usr/local/bin/
sudo systemctl start shugur-relay
```

### Backup

```bash
# Backup database
sudo -u cockroach cockroach dump shugur_relay --certs-dir=/var/lib/cockroach/certs --host=localhost:26257 > backup.sql

# Backup configuration
sudo cp -r /opt/shugur-relay/config /backup/location/
```
