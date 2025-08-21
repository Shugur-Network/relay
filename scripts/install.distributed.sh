#!/bin/bash

# Shugur Relay - SECURE Distributed Cluster Installation Script
# Multi-node CockroachDB (TLS) + Per-node Relay + Caddy
# - Generates a cluster CA
# - Creates PER-NODE node certs (SAN includes localhost + node addr + domain)
# - Creates client certs for root and relay
# - Runs CockroachDB in secure mode with --certs-dir
# - Makes your current Go code (no edits) go into secure mode by mounting CA + client.root.* at /app/certs
#
# REQUIREMENTS:
# - Run locally with sudo/root (for Docker install on remote if needed)
# - ssh/sshpass/openssl/curl available locally
# - Remote nodes: Ubuntu/Debian, passwordless sudo for Docker install
# - Open ports: 22, 26257-26258, 8080, 80, 443


set -euo pipefail

# ---------- cleanup trap ----------
cleanup_on_exit() {
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    log_error "Installation failed with exit code $exit_code"
    log_info "Attempting cleanup of temporary files..."
    
    # Clean up local directories even on failure
    if [[ -d "$CERTS_DIR" ]]; then
      log_debug "Removing local certificates directory: $CERTS_DIR"
      rm -rf "$CERTS_DIR" 2>/dev/null || sudo rm -rf "$CERTS_DIR" 2>/dev/null || true
    fi
    
    if [[ -d "$CLUSTER_DIR" ]]; then
      log_debug "Removing local deployment staging directory: $CLUSTER_DIR" 
      rm -rf "$CLUSTER_DIR" 2>/dev/null || sudo rm -rf "$CLUSTER_DIR" 2>/dev/null || true
    fi
    
    # Remove any leftover certificate generation files
    rm -f ./index.txt* ./serial.txt* 2>/dev/null || true
    rm -f ./*.pem 2>/dev/null || true
    rm -f ./*.cnf 2>/dev/null || true
    rm -f ./*.csr 2>/dev/null || true
    
    log_info "Cleanup completed. Please check the error above and retry installation."
  fi
}

# Set up cleanup trap for script failures
trap cleanup_on_exit EXIT

# ---------- colors / logging ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

log_info()    { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }
log_debug()   { echo -e "${BLUE}[DEBUG]${NC} $1"; }
log_cluster() { echo -e "${PURPLE}[CLUSTER]${NC} $1"; }

# ---------- globals ----------
CLUSTER_NODES=()
CLUSTER_URLS=()
SSH_USER="ubuntu"
SSH_PASSWORD=""
CERTS_DIR="./certs"                 # local staging for certs
CLUSTER_DIR="./cluster-deploy"      # local staging for configs
REMOTE_DIR="~/shugur-relay"         # remote base dir (expands on remote shell)
MIN_NODES=3
MAX_NODES=10

# ---------- privilege check ----------
check_privileges() {
  if [[ $EUID -ne 0 && -z "${SUDO_USER:-}" ]]; then
    echo -e "${RED}[ERROR]${NC} This script requires root privileges for Docker installation on remote nodes"
    echo "Usage: sudo $0"
    exit 1
  fi
}

# ---------- local prerequisites ----------
check_prerequisites() {
  log_info "Checking prerequisites..."
  local required=("ssh" "openssl" "sshpass" "curl")
  for cmd in "${required[@]}"; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      if [[ "$cmd" == "sshpass" ]]; then
        log_error "sshpass is required. Install: sudo apt-get install -y sshpass"
      else
        log_error "Required command '$cmd' not found. Please install it."
      fi
      exit 1
    fi
  done
  log_info "✅ Prerequisites check completed"
}

# ---------- creds ----------
get_ssh_credentials() {
  log_info "SSH Authentication Setup"
  echo ""
  echo "Password will be passed to sshpass via env (not visible in process list)."
  echo ""

  read -p "SSH Username (default: ubuntu): " input_user
  SSH_USER="${input_user:-ubuntu}"

  echo -n "SSH Password for $SSH_USER: "
  read -s SSH_PASSWORD
  echo; echo

  if [[ -z "$SSH_PASSWORD" ]]; then
    log_error "Password cannot be empty"
    exit 1
  fi

  export SSHPASS="$SSH_PASSWORD"
  log_info "Credentials configured for user: $SSH_USER"
}

# ---------- cluster nodes ----------
configure_cluster_nodes() {
  set +e +u
  log_info "Configuring cluster nodes..."
  echo ""

  local total_nodes
  while true; do
    read -r -p "How many nodes? ($MIN_NODES-$MAX_NODES): " total_nodes
    if [[ "$total_nodes" =~ ^[0-9]+$ ]] && [[ $total_nodes -ge $MIN_NODES ]] && [[ $total_nodes -le $MAX_NODES ]]; then
      break
    else
      log_warn "Enter a number between $MIN_NODES and $MAX_NODES"
    fi
  done

  echo ""
  log_info "Configuring $total_nodes nodes..."
  echo "• Enter hostnames or IPs"
  echo "• Optional format: server,domain (default domain = server)"
  echo ""

  CLUSTER_NODES=()
  CLUSTER_URLS=()

  local node_count=0
  while [[ $node_count -lt $total_nodes ]]; do
    local node_input
    read -r -p "Node $((node_count + 1)) server[,domain]: " node_input

    if [[ -z "$node_input" ]]; then
      log_warn "Server cannot be empty."
      continue
    fi

    local server_name="" domain_name=""
    if [[ "$node_input" =~ ^([^,]+),(.+)$ ]]; then
      server_name="${BASH_REMATCH[1]}"
      domain_name="${BASH_REMATCH[2]}"
    else
      server_name="$node_input"
      domain_name="$node_input"
    fi

    server_name="${server_name// /}"
    domain_name="${domain_name// /}"

    if [[ -z "$server_name" ]]; then
      log_error "Parsed empty server; try again."
      continue
    fi

    CLUSTER_NODES+=("$server_name")
    CLUSTER_URLS+=("$domain_name")
    ((node_count++))
    log_info "✅ Added: $server_name (domain: $domain_name)"
    log_debug "Current count: $node_count/$total_nodes"
  done

  echo ""
  log_info "✅ Cluster configuration completed ($node_count nodes)"
  echo ""
  set -euo pipefail
}

# ---------- helpers ----------
is_ip() {
  local ip=$1
  [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]
}

# ---------- ssh helpers ----------
ssh_exec() {
  local host="$1"; shift
  local command="$1"; shift
  local description="${1:-executing command}"

  log_debug "[$host] $description"
  if ! sshpass -e ssh \
      -o StrictHostKeyChecking=no \
      -o UserKnownHostsFile=/dev/null \
      -o ConnectTimeout=10 \
      "$SSH_USER@$host" "$command"; then
    log_error "Failed on $host: $description"
    return 1
  fi
}

scp_copy() {
  local source="$1" host="$2" destination="$3"
  local description="${4:-copying files}"

  log_debug "[$host] $description"
  sshpass -e ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$SSH_USER@$host" "mkdir -p $destination" >/dev/null 2>&1 || true

  if ! sshpass -e scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r "$source" "$SSH_USER@$host:$destination"; then
    log_error "Failed to copy to $host: $description"
    return 1
  fi
}

# ---------- connectivity ----------
test_ssh_connectivity() {
  log_info "Testing SSH connectivity..."
  local failed=0
  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"
    log_debug "Testing $node_id ($node)..."
    if ssh_exec "$node" "echo 'SSH OK'"; then
      log_info "✅ $node_id ($node) SSH OK"
    else
      log_error "❌ $node_id ($node) SSH FAILED"
      ((failed++))
    fi
  done
  if [[ $failed -gt 0 ]]; then
    log_error "SSH failed for $failed node(s)"
    exit 1
  fi
  log_info "✅ All nodes accessible via SSH"
}

# ---------- docker install (remote) ----------
install_docker_on_nodes() {
  log_cluster "Checking/installing Docker on nodes (requires passwordless sudo on remote)..."
  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"

    if ssh_exec "$node" "command -v docker >/dev/null 2>&1" "checking Docker"; then
      log_debug "✅ Docker present on $node_id"
      continue
    fi

    log_info "Installing Docker on $node_id..."
    local docker_script='#!/bin/bash
if ! sudo -n true 2>/dev/null; then
  echo "ERROR: passwordless sudo required for Docker install"
  exit 1
fi
sudo apt-get update -y && \
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release && \
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null && \
sudo apt-get update -y && \
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin && \
sudo systemctl enable --now docker && \
sudo usermod -aG docker '"$SSH_USER"'
'

    ssh_exec "$node" "cat > /tmp/install_docker.sh <<'EOS'
$docker_script
EOS
chmod +x /tmp/install_docker.sh && /tmp/install_docker.sh && rm /tmp/install_docker.sh" "installing Docker"
    log_info "✅ Docker installed on $node_id"
  done
  log_cluster "✅ Docker installation completed"
}

# ---------- certs (CA, per-node, client root/relay) ----------
generate_certificates() {
  log_cluster "Generating TLS certificates for SECURE Cockroach (following official CockroachDB docs)..."
  rm -rf "$CERTS_DIR"
  mkdir -p "$CERTS_DIR/safe-dir"

  # Step 1: Create CA configuration file (per CockroachDB docs)
  cat > "$CERTS_DIR/ca.cnf" << 'EOF'
# OpenSSL CA configuration file
[ ca ]
default_ca = CA_default

[ CA_default ]
default_days = 3650
database = index.txt
serial = serial.txt
default_md = sha256
copy_extensions = copy
unique_subject = no

# Used to create the CA certificate.
[ req ]
prompt=no
distinguished_name = distinguished_name
x509_extensions = extensions

[ distinguished_name ]
organizationName = Cockroach
commonName = Cockroach CA

[ extensions ]
keyUsage = critical,digitalSignature,nonRepudiation,keyEncipherment,keyCertSign
basicConstraints = critical,CA:true,pathlen:1

# Common policy for nodes and users.
[ signing_policy ]
organizationName = supplied
commonName = optional

# Used to sign node certificates.
[ signing_node_req ]
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth,clientAuth

# Used to sign client certificates.
[ signing_client_req ]
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = clientAuth
EOF

  # Step 2: Create CA key and certificate
  log_debug "Creating CA key and certificate..."
  cd "$CERTS_DIR"
  
  # Create CA key in safe directory
  openssl genrsa -out safe-dir/ca.key 2048
  chmod 400 safe-dir/ca.key
  
  # Create CA certificate
  openssl req \
    -new \
    -x509 \
    -config ca.cnf \
    -key safe-dir/ca.key \
    -out ca.crt \
    -days 3650 \
    -batch
  
  # Reset database and index files
  rm -f index.txt serial.txt
  touch index.txt
  echo '01' > serial.txt

  # Step 3: Generate per-node certificates
  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}"
    local url="${CLUSTER_URLS[$i]}"
    local node_id="node$((i+1))"
    
    log_debug "Creating certificate for $node_id ($node)..."
    
    # Create node configuration file
    cat > "node-$node_id.cnf" << EOF
# OpenSSL node configuration file
[ req ]
prompt=no
distinguished_name = distinguished_name
req_extensions = extensions

[ distinguished_name ]
organizationName = Cockroach

[ extensions ]
subjectAltName = critical,DNS:node,DNS:localhost,IP:127.0.0.1,DNS:$node$(if [[ "$url" != "$node" ]]; then echo ",DNS:$url"; fi)$(if is_ip "$node"; then echo ",IP:$node"; fi)$(if [[ "$url" != "$node" ]] && is_ip "$url"; then echo ",IP:$url"; fi)
EOF
    
    # Create node key
    openssl genrsa -out "node-$node_id.key" 2048
    chmod 400 "node-$node_id.key"
    
    # Create node CSR
    openssl req \
      -new \
      -config "node-$node_id.cnf" \
      -key "node-$node_id.key" \
      -out "node-$node_id.csr" \
      -batch
    
    # Sign node certificate
    openssl ca \
      -config ca.cnf \
      -keyfile safe-dir/ca.key \
      -cert ca.crt \
      -policy signing_policy \
      -extensions signing_node_req \
      -out "node-$node_id.crt" \
      -outdir . \
      -in "node-$node_id.csr" \
      -batch
    
    # Verify certificate
    log_debug "Verifying certificate for $node_id..."
    openssl x509 -in "node-$node_id.crt" -text | grep "X509v3 Subject Alternative Name" -A 1 || true
  done

  # Step 4: Create client configuration file
  cat > "client.cnf" << 'EOF'
[ req ]
prompt=no
distinguished_name = distinguished_name
req_extensions = extensions

[ distinguished_name ]
organizationName = Cockroach
commonName = root

[ extensions ]
subjectAltName = DNS:root
EOF

  # Step 5: Generate client certificates (root and relay)
  for username in "root" "relay"; do
    log_debug "Creating client certificate for user: $username..."
    
    # Update client.cnf for this user
    sed -i "s/commonName = .*/commonName = $username/" client.cnf
    if [[ "$username" == "relay" ]]; then
      sed -i "s/subjectAltName = .*/subjectAltName = DNS:relay/" client.cnf
    else
      sed -i "s/subjectAltName = .*/subjectAltName = DNS:root/" client.cnf
    fi
    
    # Create client key
    openssl genrsa -out "client.$username.key" 2048
    chmod 400 "client.$username.key"
    
    # Create client CSR
    openssl req \
      -new \
      -config client.cnf \
      -key "client.$username.key" \
      -out "client.$username.csr" \
      -batch
    
    # Sign client certificate
    openssl ca \
      -config ca.cnf \
      -keyfile safe-dir/ca.key \
      -cert ca.crt \
      -policy signing_policy \
      -extensions signing_client_req \
      -out "client.$username.crt" \
      -outdir . \
      -in "client.$username.csr" \
      -batch
    
    # Verify certificate
    log_debug "Verifying client certificate for $username..."
    openssl x509 -in "client.$username.crt" -text | grep "CN=" || true
  done

  # Cleanup temporary files
  rm -f *.csr *.cnf index.txt* serial.txt*
  
  # Set proper permissions
  chmod 644 *.crt
  chmod 600 *.key
  
  cd - >/dev/null
  log_cluster "✅ Certificates generated following CockroachDB best practices"
}

# ---------- deployment structure ----------
create_deployment_structure() {
  log_debug "Creating deployment directory structure..."
  rm -rf "$CLUSTER_DIR"
  mkdir -p "$CLUSTER_DIR"
  for i in "${!CLUSTER_NODES[@]}"; do
    local node_id="node$((i+1))"
    mkdir -p "$CLUSTER_DIR/$node_id"/{config,certs/cockroach,certs/relay,logs/{relay,cockroachdb,caddy}}
  done
  log_debug "✅ Deployment structure created"
}

# ---------- per-node config & cert staging ----------
create_node_config() {
  local idx=$1
  local node="${CLUSTER_NODES[$idx]}"
  local node_url="${CLUSTER_URLS[$idx]}"
  local node_id="node$((idx+1))"

  log_debug "Creating configuration for $node_id..."

  # Stage certs for this node (DB server + client root certs)
  cp "$CERTS_DIR/ca.crt"               "$CLUSTER_DIR/$node_id/certs/cockroach/ca.crt"
  cp "$CERTS_DIR/node-$node_id.crt"    "$CLUSTER_DIR/$node_id/certs/cockroach/node.crt"
  cp "$CERTS_DIR/node-$node_id.key"    "$CLUSTER_DIR/$node_id/certs/cockroach/node.key"
  cp "$CERTS_DIR/client.root.crt"      "$CLUSTER_DIR/$node_id/certs/cockroach/client.root.crt"
  cp "$CERTS_DIR/client.root.key"      "$CLUSTER_DIR/$node_id/certs/cockroach/client.root.key"

  # Relay client certs folder will be mounted to /app/certs (where your Go looks for ./certs/*)
  cp "$CERTS_DIR/ca.crt"               "$CLUSTER_DIR/$node_id/certs/relay/ca.crt"
  cp "$CERTS_DIR/client.relay.crt"     "$CLUSTER_DIR/$node_id/certs/relay/client.relay.crt"
  cp "$CERTS_DIR/client.relay.key"     "$CLUSTER_DIR/$node_id/certs/relay/client.relay.key"
  # >>> ADD root certs here too so your current Go code (expects client.root.*) works without changes:
  cp "$CERTS_DIR/client.root.crt"      "$CLUSTER_DIR/$node_id/certs/relay/client.root.crt"
  cp "$CERTS_DIR/client.root.key"      "$CLUSTER_DIR/$node_id/certs/relay/client.root.key"

  # Set proper permissions for certificates
  chmod 600 "$CLUSTER_DIR/$node_id/certs/cockroach"/*.key "$CLUSTER_DIR/$node_id/certs/relay"/*.key
  chmod 644 "$CLUSTER_DIR/$node_id/certs/cockroach"/*.crt "$CLUSTER_DIR/$node_id/certs/relay"/*.crt

  # relay config
  cat > "$CLUSTER_DIR/$node_id/config/config.yaml" << EOF
GENERAL: {}

LOGGING:
  LEVEL: info
  FILE:
  FORMAT: json
  MAX_SIZE: 100
  MAX_BACKUPS: 10
  MAX_AGE: 30

METRICS:
  ENABLED: true
  PORT: 8181

RELAY:
  NAME: "$node_url"
  DESCRIPTION: "High-performance, reliable, scalable Nostr relay"
  CONTACT: "support@shugur.com"
  ICON: "https://github.com/Shugur-Network/relay/raw/main/logo.png"
  BANNER: "https://github.com/Shugur-Network/relay/raw/main/banner.png"
  WS_ADDR: ":8080"
  PUBLIC_URL: "wss://$node_url"
  EVENT_CACHE_SIZE: 50000
  SEND_BUFFER_SIZE: 8192
  WRITE_TIMEOUT: 30s
  IDLE_TIMEOUT: 300s
  THROTTLING:
    MAX_CONTENT_LENGTH: 65536
    MAX_CONNECTIONS: 2000
    BAN_THRESHOLD: 5
    BAN_DURATION: 300
    RATE_LIMIT:
      ENABLED: true
      MAX_EVENTS_PER_SECOND: 200
      MAX_REQUESTS_PER_SECOND: 400
      BURST_SIZE: 100
      PROGRESSIVE_BAN: true
      MAX_BAN_DURATION: 24h

RELAY_POLICY:
  BLACKLIST:
    PUBKEYS: []
  WHITELIST:
    PUBKEYS: []

DATABASE:
  SERVER: "localhost"
  PORT: 26257
EOF

  # Caddyfile
  cat > "$CLUSTER_DIR/$node_id/config/Caddyfile" << EOF
$node_url {
    handle /api/* {
        reverse_proxy localhost:8080
    }
    handle {
        reverse_proxy localhost:8080 {
            header_up Host {host}
            header_up X-Real-IP {remote}
            header_up X-Forwarded-For {remote}
            header_up X-Forwarded-Proto {scheme}
        }
    }
    @internal {
        remote_ip 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16 127.0.0.1
    }
    handle_path /metrics {
        handle @internal {
            reverse_proxy localhost:8181
        }
        handle {
            respond "Access Denied" 403
        }
    }
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        Referrer-Policy "strict-origin-when-cross-origin"
        X-Cluster-Node "$node_id"
        -Server
        -X-Powered-By
    }
    encode gzip zstd
    log {
        output file /var/log/caddy/access.log {
            roll_size 100mb
            roll_keep 10
        }
        format json
    }
}

:8090 {
    handle /health {
        respond "OK" 200
    }
}
EOF

  # Build join list (using RPC port 26258)
  local join_addresses=""
  for j in "${!CLUSTER_NODES[@]}"; do
    [[ $j -gt 0 ]] && join_addresses+=","
    join_addresses+="${CLUSTER_NODES[$j]}:26258"
  done

  # Compose (SECURE Cockroach; relay volume provides /app/certs used by your Go code)
  cat > "$CLUSTER_DIR/$node_id/config/docker-compose.yml" << EOF
services:
  cockroachdb:
    image: cockroachdb/cockroach:latest
    container_name: cockroachdb
    hostname: cockroachdb
    command: start --certs-dir=/cockroach/certs --http-addr=0.0.0.0:8080 --listen-addr=0.0.0.0:26258 --sql-addr=0.0.0.0:26257 --advertise-addr=${CLUSTER_NODES[$idx]}:26258 --advertise-sql-addr=${CLUSTER_NODES[$idx]}:26257 --join=$join_addresses --cache=25% --max-sql-memory=25%
    volumes:
      - cockroach_data:/cockroach/cockroach-data
      - ./logs/cockroachdb:/cockroach/logs
      - ./certs/cockroach:/cockroach/certs:ro
    ports:
      - "26257:26257"  # CockroachDB SQL
      - "26258:26258"  # CockroachDB RPC
      - "9090:8080"    # CockroachDB Admin UI
    environment:
      - COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING=true
    healthcheck:
      test: ["CMD", "/cockroach/cockroach", "sql", "--certs-dir=/cockroach/certs", "--host=localhost:26257", "--execute=SELECT 1;"]
      interval: 60s
      timeout: 30s
      retries: 10
      start_period: 120s
    restart: unless-stopped
    networks:
      - cluster_network

  relay:
    image: ghcr.io/shugur-network/relay:latest
    container_name: relay
    hostname: relay
    user: "1001:1001"  # Ensure relay runs as UID 1001 to match certificate ownership
    # Your Go code constructs its own DSN and ignores these SSL envs.
    # The important bit is the volume below that mounts /app/certs with ca.crt + client.root.*.
    environment:
      - SHUGUR_ENV=production
      - SHUGUR_DB_HOST=localhost
      - SHUGUR_DB_PORT=26257
      - SHUGUR_DB_USER=relay
      - SHUGUR_LOG_LEVEL=info
      - SHUGUR_LOG_FORMAT=json
      - SHUGUR_METRICS_ENABLED=true
      - SHUGUR_WS_PORT=8080
      - SHUGUR_METRICS_PORT=8181
      - SHUGUR_MAX_CONNECTIONS=2000
      - SHUGUR_CLUSTER_NODE=$node_id
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./logs/relay:/app/logs
      - ./certs/relay:/app/certs:ro   # contains ca.crt + client.root.crt|key (for your current Go code)
    depends_on:
      cockroachdb:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://localhost:8080/api/info || curl -fsI http://localhost:8080/api/info"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
    restart: unless-stopped
    network_mode: host

  caddy:
    image: caddy:latest
    container_name: caddy
    hostname: caddy
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
      - ./logs/caddy:/var/log/caddy
    depends_on:
      - relay
    healthcheck:
      test: ["CMD", "caddy", "validate", "--config", "/etc/caddy/Caddyfile"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    restart: unless-stopped
    network_mode: host

volumes:
  cockroach_data:
    driver: local
  caddy_data:
    driver: local
  caddy_config:
    driver: local

networks:
  cluster_network:
    driver: bridge
EOF

  log_info "✅ Configuration & cert staging created for $node_id"
}

create_all_node_configs() {
  log_cluster "Creating node-specific configurations..."
  create_deployment_structure
  for i in "${!CLUSTER_NODES[@]}"; do
    create_node_config "$i"
  done
  log_cluster "All node configurations created"
}

# ---------- deploy ----------
deploy_files_to_nodes() {
  log_cluster "Deploying files to all nodes..."
  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"

    log_debug "Deploying to $node_id ($node)..."
    
    # Create directories and ensure proper ownership BEFORE copying files
    ssh_exec "$node" "
      # Clean up any existing directories with wrong ownership
      sudo rm -rf $REMOTE_DIR 2>/dev/null || rm -rf $REMOTE_DIR 2>/dev/null || true
      
      # Create directories as ubuntu user
      mkdir -p $REMOTE_DIR/{certs/{cockroach,relay},logs/{relay,cockroachdb,caddy}}
      
      # Ensure ubuntu user owns the entire directory structure
      CURRENT_USER=\$(id -u)
      CURRENT_GROUP=\$(id -g)
      sudo chown -R \$CURRENT_USER:\$CURRENT_GROUP $REMOTE_DIR 2>/dev/null || chown -R \$CURRENT_USER:\$CURRENT_GROUP $REMOTE_DIR 2>/dev/null || true
    " "preparing directories with correct ownership"

    scp_copy "$CLUSTER_DIR/$node_id/config/." "$node" "$REMOTE_DIR/" "copying config"
    scp_copy "$CLUSTER_DIR/$node_id/certs/cockroach/." "$node" "$REMOTE_DIR/certs/cockroach/" "copying cockroach certs"
    scp_copy "$CLUSTER_DIR/$node_id/certs/relay/."     "$node" "$REMOTE_DIR/certs/relay/"     "copying relay certs"

    # Set proper ownership and permissions for certificates
    # The relay container runs as UID 1001, cockroach certs need ubuntu user access
    ssh_exec "$node" "
      # Get current user info for cockroach certs
      CURRENT_USER=\$(id -u)
      CURRENT_GROUP=\$(id -g)
      
      # Ensure UID 1001 exists on the system for relay certificate ownership
      if ! id 1001 >/dev/null 2>&1; then
        sudo useradd -u 1001 -r -s /bin/false relay-certs 2>/dev/null || true
      fi
      
      # Set ownership for cockroach certs to current user (ubuntu)
      sudo chown -R \$CURRENT_USER:\$CURRENT_GROUP $REMOTE_DIR/certs/cockroach/ 2>/dev/null || chown -R \$CURRENT_USER:\$CURRENT_GROUP $REMOTE_DIR/certs/cockroach/ 2>/dev/null || true
      
      # Set ownership for relay certs to UID 1001 (relay container user)
      sudo chown -R 1001:1001 $REMOTE_DIR/certs/relay/ 2>/dev/null || chown -R 1001:1001 $REMOTE_DIR/certs/relay/ 2>/dev/null || true
      
      # Set proper permissions for certificate security
      # Private keys: 600 (read/write for owner only)
      find $REMOTE_DIR/certs/ -name '*.key' -exec chmod 600 {} \; 2>/dev/null || true
      # Certificates: 644 (readable by all, writable by owner)
      find $REMOTE_DIR/certs/ -name '*.crt' -exec chmod 644 {} \; 2>/dev/null || true
      # Directories: 755 (accessible to all)
      find $REMOTE_DIR/certs/ -type d -exec chmod 755 {} \; 2>/dev/null || true
      
      # Verify permissions for debugging
      echo 'Certificate directory permissions:'
      ls -la $REMOTE_DIR/certs/relay/ || echo 'Relay certs directory not accessible'
    " "setting cert ownership: cockroach certs to ubuntu user, relay certs to UID 1001"
    log_debug "✅ Files deployed to $node_id"
  done
  log_cluster "✅ Deployment complete"
}

# ---------- DB init (SECURE) ----------
initialize_cockroachdb_cluster() {
  log_cluster "Initializing SECURE CockroachDB cluster..."
  
  # Step 1: Start ONLY the first node initially
  local first_node="${CLUSTER_NODES[0]}"
  log_debug "Starting first CockroachDB node (node1)..."
  ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose up -d cockroachdb" "starting first cockroachdb node"

  log_cluster "Waiting for first node to be ready (5 seconds)..."
  sleep 5

  # Step 2: Initialize the cluster on the first node (MUST be done before connectivity check)
  log_cluster "Initializing cluster on first node (secure)..."
  ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose exec -T cockroachdb /cockroach/cockroach init --certs-dir=/cockroach/certs --host=localhost:26258" "initializing cluster"

  log_cluster "Waiting for cluster initialization to complete..."
  sleep 5

  # Step 3: Check if first node is responding after initialization
  log_debug "Testing CockroachDB connectivity on first node..."
  for attempt in {1..10}; do
    if ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose exec -T cockroachdb /cockroach/cockroach sql --certs-dir=/cockroach/certs --host=localhost:26257 --execute='SELECT 1;'" "testing connectivity"; then
      log_debug "✅ CockroachDB responding on first node"
      break
    elif [[ $attempt -eq 10 ]]; then
      log_error "CockroachDB not responding after 10 attempts"
      log_cluster "Checking CockroachDB logs on first node..."
      ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose logs cockroachdb | tail -20" "checking logs" || true
      exit 1
    else
      log_debug "Attempt $attempt/10 failed, waiting 5 seconds..."
      sleep 5
    fi
  done

  # Step 4: Create database and user on the initialized cluster
  log_cluster "Creating database and user (secure)..."
  ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose exec -T cockroachdb /cockroach/cockroach sql --certs-dir=/cockroach/certs --host=localhost:26257 --execute=\"CREATE DATABASE IF NOT EXISTS shugur;\"" "create database"
  ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose exec -T cockroachdb /cockroach/cockroach sql --certs-dir=/cockroach/certs --host=localhost:26257 --execute=\"CREATE USER IF NOT EXISTS relay; GRANT ALL ON DATABASE shugur TO relay;\"" "create user/grant"

  # Step 5: Now start the remaining nodes which will join the initialized cluster
  if [[ ${#CLUSTER_NODES[@]} -gt 1 ]]; then
    log_cluster "Starting remaining CockroachDB nodes to join the cluster..."
    for i in "${!CLUSTER_NODES[@]}"; do
      if [[ $i -eq 0 ]]; then continue; fi  # Skip first node (already started)
      
      local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"
      log_debug "Starting CockroachDB on $node_id..."
      ssh_exec "$node" "cd $REMOTE_DIR && docker compose up -d cockroachdb" "starting cockroachdb"
      sleep 5  # Give each node time to start and join
    done
    
    log_cluster "Waiting for all nodes to join the cluster (10 seconds)..."
    sleep 10
  fi

  # Step 6: Verify cluster status
  log_cluster "Verifying cluster status..."
  ssh_exec "$first_node" "cd $REMOTE_DIR && docker compose exec -T cockroachdb /cockroach/cockroach node status --certs-dir=/cockroach/certs --host=localhost:26257" "checking node status"

  log_cluster "✅ Secure CockroachDB cluster initialized"
}

# ---------- start services ----------
start_relay_services() {
  log_cluster "Starting Relay services..."
  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"
    log_debug "Starting relay on $node_id..."
    ssh_exec "$node" "cd $REMOTE_DIR && docker compose up -d relay" "starting relay"
    sleep 5
  done
  log_cluster "✅ Relay services started"
}

start_caddy_services() {
  log_cluster "Starting Caddy services..."
  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"
    log_debug "Starting caddy on $node_id..."
    ssh_exec "$node" "cd $REMOTE_DIR && docker compose up -d caddy" "starting caddy"
    sleep 3
  done
  log_cluster "✅ Caddy services started"
}

# ---------- debugging helper ----------
debug_certificates() {
  log_cluster "🔍 Certificate debugging information..."
  
  if [[ -d "$CERTS_DIR" ]]; then
    log_debug "Certificate files in $CERTS_DIR:"
    ls -la "$CERTS_DIR/"
    echo ""
    
    # Check CA certificate
    if [[ -f "$CERTS_DIR/ca.crt" ]]; then
      log_debug "CA Certificate details:"
      openssl x509 -in "$CERTS_DIR/ca.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After|X509v3)" || true
      echo ""
    fi
    
    # Check first node certificate
    local first_cert="$CERTS_DIR/node-node1.crt"
    if [[ -f "$first_cert" ]]; then
      log_debug "Node1 Certificate SAN details:"
      openssl x509 -in "$first_cert" -text -noout | grep -A5 "X509v3 Subject Alternative Name" || true
      echo ""
    fi
    
    # Check client certificates
    for user in root relay; do
      if [[ -f "$CERTS_DIR/client.$user.crt" ]]; then
        log_debug "Client $user Certificate CN:"
        openssl x509 -in "$CERTS_DIR/client.$user.crt" -text -noout | grep "Subject:" || true
      fi
    done
  else
    log_warn "Certificate directory $CERTS_DIR not found"
  fi
  echo ""
}

# ---------- debugging remote logs ----------
debug_remote_logs() {
  local node="$1"
  local node_id="$2"
  
  log_debug "🔍 Checking logs on $node_id ($node)..."
  
  # Check if containers are running
  ssh_exec "$node" "cd $REMOTE_DIR && docker compose ps" "checking container status" || true
  
  # Check CockroachDB logs
  log_debug "CockroachDB logs (last 20 lines):"
  ssh_exec "$node" "cd $REMOTE_DIR && docker compose logs --tail=20 cockroachdb" "cockroachdb logs" || true
  
  # Check certificate permissions
  log_debug "Certificate file permissions:"
  ssh_exec "$node" "ls -la $REMOTE_DIR/certs/cockroach/" "cert permissions" || true
  
  echo ""
}
verify_cluster_health() {
  log_cluster "Verifying cluster health..."
  local healthy=0 total=${#CLUSTER_NODES[@]}

  for i in "${!CLUSTER_NODES[@]}"; do
    local node="${CLUSTER_NODES[$i]}" node_id="node$((i+1))"
    log_debug "Checking $node_id local relay..."
    if sshpass -e ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$SSH_USER@$node" \
         "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/info 2>/dev/null | grep -q '^200$'"; then
      log_debug "✅ $node_id relay healthy"
      ((healthy++))
    else
      log_warn  "❌ $node_id relay health check failed"
      debug_remote_logs "$node" "$node_id"
    fi

    sshpass -e ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$SSH_USER@$node" \
      "cd $REMOTE_DIR && docker compose ps --format 'table {{.Service}}\t{{.Status}}' | grep -E 'running|healthy'" >/dev/null 2>&1 \
      || log_warn "Some services on $node_id may not be running"
  done

  local first_node="${CLUSTER_NODES[0]}"
  log_debug "Checking CockroachDB cluster status (secure)..."
  if ! sshpass -e ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$SSH_USER@$first_node" \
    "cd $REMOTE_DIR && docker compose exec -T cockroachdb /cockroach/cockroach node status --certs-dir=/cockroach/certs --host=localhost:26257" \
    >/dev/null 2>&1; then
    log_warn "CockroachDB cluster status check had issues"
    debug_remote_logs "$first_node" "node1"
  fi

  log_cluster "Health check: $healthy/$total nodes reported healthy"
  if [[ $healthy -eq $total ]]; then
    log_cluster "🎉 All nodes healthy"
  else
    log_warn "Some nodes may need attention - check logs above"
  fi
}

test_external_connectivity() {
  log_cluster "Testing external connectivity..."
  local working=0 total=${#CLUSTER_NODES[@]}

  for i in "${!CLUSTER_NODES[@]}"; do
    local node_url="${CLUSTER_URLS[$i]}" node_id="node$((i+1))"
    log_debug "Testing $node_id ($node_url)..."

    local http_status https_status relay_status
    http_status=$(curl -s -o /dev/null -w "%{http_code}" "http://$node_url/" 2>/dev/null || echo "000")
    https_status=$(curl -s -o /dev/null -w "%{http_code}" "https://$node_url/" 2>/dev/null || echo "000")
    relay_status=$(curl -s -o /dev/null -w "%{http_code}" "https://$node_url/api/info" 2>/dev/null || echo "000")

    [[ "$relay_status" == "200" ]] && { log_debug "✅ relay HTTPS /api/info OK ($relay_status)"; ((working++)); } || log_warn "❌ relay HTTPS /api/info FAILED ($relay_status)"
    [[ "$http_status" =~ ^(301|308)$ ]] && log_debug "✅ HTTP redirect OK ($http_status)" || log_warn "❌ HTTP redirect FAILED ($http_status)"
    [[ "$https_status" == "200" ]] && log_debug "✅ HTTPS OK ($https_status)" || log_warn "❌ HTTPS FAILED ($https_status) - check DNS/firewall"
  done

  log_cluster "External connectivity: $working/$total nodes reachable via HTTPS"
}

# ---------- local cleanup after installation ----------
cleanup_local_directories() {
  log_cluster "Cleaning up local working directories..."
  
  # Preserve CA certificate and key for future node additions
  if [[ -f "$CERTS_DIR/ca.crt" && -f "$CERTS_DIR/ca.key" ]]; then
    log_debug "Preserving CA certificate and key for future node additions..."
    mkdir -p "./cluster-ca-backup"
    cp "$CERTS_DIR/ca.crt" "./cluster-ca-backup/" 2>/dev/null || true
    cp "$CERTS_DIR/ca.key" "./cluster-ca-backup/" 2>/dev/null || true
    cp "$CERTS_DIR/client.root.key" "./cluster-ca-backup/" 2>/dev/null || true
    cp "$CERTS_DIR/client.relay.key" "./cluster-ca-backup/" 2>/dev/null || true
    log_debug "✅ CA certificate backed up to ./cluster-ca-backup/"
    log_debug "✅ CA private key backed up to ./cluster-ca-backup/"
  fi
  
  # Remove certificates directory (contains temporary CA keys and node certificates)
  if [[ -d "$CERTS_DIR" ]]; then
    log_debug "Removing working certificates directory: $CERTS_DIR"
    rm -rf "$CERTS_DIR"
    log_debug "✅ Working certificates directory removed"
  else
    log_debug "No local certificates directory found to clean"
  fi
  
  # Remove cluster deployment staging directory
  if [[ -d "$CLUSTER_DIR" ]]; then
    log_debug "Removing local deployment staging directory: $CLUSTER_DIR"
    rm -rf "$CLUSTER_DIR"
    log_debug "✅ Local deployment staging directory removed"
  else
    log_debug "No local deployment directory found to clean"
  fi
  
  # Remove any leftover certificate generation files
  rm -f ./index.txt* ./serial.txt* 2>/dev/null || true
  rm -f ./*.pem 2>/dev/null || true
  rm -f ./*.cnf 2>/dev/null || true
  rm -f ./*.csr 2>/dev/null || true
  
  log_cluster "✅ Local directories cleaned - CA certificates preserved in ./cluster-ca-backup/ for future node additions"
}

# ---------- summary ----------
show_completion_message() {
  echo ""
  echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════════════════════╗${NC}"
  echo -e "${GREEN}║                      🎉 SECURE DISTRIBUTED CLUSTER INSTALLATION COMPLETE! 🎉              ║${NC}"
  echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════════════════════╝${NC}"
  echo ""
  echo -e "${BLUE}🌐 Relay Access Points (per-node Caddy):${NC}"
  for i in "${!CLUSTER_NODES[@]}"; do
    local node_url="${CLUSTER_URLS[$i]}"
    echo -e "   ${GREEN}•${NC} Node $((i+1)) (${CLUSTER_NODES[$i]}):"
    echo -e "     - WebSocket: ${YELLOW}wss://$node_url${NC}"
    echo -e "     - Relay info: ${YELLOW}https://$node_url/api/info${NC}"
  done
  echo ""
  echo -e "${BLUE}🔧 Management & Monitoring:${NC}"
  echo -e "   ${GREEN}•${NC} Cockroach Admin UI: ${YELLOW}http://${CLUSTER_NODES[0]}:9090/${NC}"
  echo -e "   ${GREEN}•${NC} Local health (on node): ${YELLOW}curl http://localhost:8080/api/info${NC}"
  echo ""
  echo -e "${BLUE}📋 Next Steps:${NC}"
  echo -e "   ${GREEN}•${NC} Relay certs owned by UID 1001 for seamless container access"
  echo -e "   ${GREEN}•${NC} Local working directories automatically cleaned after installation"
  echo -e "   ${GREEN}•${NC} CA certificates preserved in ./cluster-ca-backup/ for adding nodes"
  echo -e "   ${GREEN}•${NC} To add nodes later: use ./scripts/add-cluster-node.sh"
  echo ""
  echo -e "${YELLOW}💡 Installation complete! CA preserved for future node additions.${NC}"
  echo ""
}

# ---------- main ----------
main() {
  check_privileges
  log_cluster "Step 1: Checking prerequisites"
  check_prerequisites

  log_cluster "Step 2: SSH Authentication Setup"
  get_ssh_credentials

  log_cluster "Step 3: Configuring Cluster Nodes"
  configure_cluster_nodes

  log_cluster "Step 4: Testing SSH Connectivity"
  test_ssh_connectivity

  log_cluster "Step 5: Installing Docker on Nodes"
  install_docker_on_nodes

  log_cluster "Step 6: Generating TLS Certificates (following CockroachDB best practices)"
  generate_certificates
  
  # Debug certificates after generation
  debug_certificates

  log_cluster "Step 7: Creating Node Configs"
  create_all_node_configs

  log_cluster "Step 8: Deploying Files to Nodes"
  deploy_files_to_nodes

  log_cluster "Step 9: Initializing CockroachDB (secure)"
  initialize_cockroachdb_cluster

  log_cluster "Step 10: Starting Relay Services"
  start_relay_services

  log_cluster "Step 11: Starting Caddy Services"
  start_caddy_services

  log_cluster "Step 12: Verifying Cluster Health"
  sleep 30  # Increased wait time for services to stabilize
  verify_cluster_health

  log_cluster "Step 13: Testing External Connectivity"
  test_external_connectivity

  log_cluster "Step 14: Cleaning Up Local Directories"
  cleanup_local_directories

  show_completion_message
}

if [[ $# -gt 0 ]]; then
  case $1 in
    --help|-h)
      echo "Shugur Relay SECURE Cluster Installer"
      echo
      echo "Usage: sudo $0 [options]"
      echo
      echo "Options:"
      echo "  --help, -h         Show this help message"
      echo "  --debug            Debug certificates and remote logs (requires existing installation)"
      echo
      echo "Installs:"
      echo "  • Multi-node CockroachDB (TLS, --certs-dir) following official best practices"
      echo "  • Per-node Relay + Caddy (HTTPS)"
      echo "  • Provides /app/certs with CA + client.root.* for your current Go code"
      echo "  • Uses proper OpenSSL certificate generation per CockroachDB documentation"
      echo
      echo "Requirements:"
      echo "  • Run with sudo/root"
      echo "  • SSH access to all nodes; remote passwordless sudo for Docker install"
      echo "  • Ubuntu/Debian nodes assumed"
      echo "  • Open ports: 22, 26257-26258, 8080, 80, 443"
      echo
      exit 0
      ;;
    --debug)
      log_cluster "🔍 Debug mode - checking certificates and logs..."
      if [[ -d "$CERTS_DIR" ]]; then
        debug_certificates
      else
        log_warn "No local certificates found in $CERTS_DIR"
      fi
      
      # If nodes are configured, check remote logs
      if [[ ${#CLUSTER_NODES[@]} -gt 0 ]] || [[ -f "./cluster-nodes.conf" ]]; then
        get_ssh_credentials
        # Load node configuration if available
        if [[ -f "./cluster-nodes.conf" ]]; then
          source "./cluster-nodes.conf"
        fi
        for i in "${!CLUSTER_NODES[@]}"; do
          debug_remote_logs "${CLUSTER_NODES[$i]}" "node$((i+1))"
        done
      else
        log_info "No cluster configuration found. Run normal installation first."
      fi
      exit 0
      ;;
    *)
      log_error "Unknown option: $1"
      echo "Use --help for usage."
      exit 1
      ;;
  esac
fi

main
