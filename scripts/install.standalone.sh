#!/bin/bash
# Shugur Relay - Standalone Installation Script (polished)
# Complete one-command installer: Docker (if needed) + CockroachDB + Relay + Caddy
# DB and schema are auto-managed by the relay on first start.

set -Eeuo pipefail

# ---------- cleanup trap ----------
cleanup_on_exit() {
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    log_error "Installation failed with exit code $exit_code"
    log_info "Attempting cleanup of temporary files..."
    
    # Clean up installation files even on failure
    local files_to_remove=("Caddyfile" "config.yaml" "docker-compose.standalone.yml")
    for file in "${files_to_remove[@]}"; do
      if [[ -f "$file" ]]; then
        log_debug "Removing installation file: $file"
        rm -f "$file" 2>/dev/null || true
      fi
    done
    
    # Remove logs directory if it was created during installation  
    if [[ -d "./logs" ]]; then
      log_debug "Removing logs directory: ./logs"
      rm -rf "./logs" 2>/dev/null || true
    fi
    
    log_info "Cleanup completed. Please check the error above and retry installation."
  fi
}

# Set up cleanup trap for script failures
trap cleanup_on_exit EXIT

# ---------- utils ----------
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log_info(){ echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn(){ echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error(){ echo -e "${RED}[ERROR]${NC} $1"; }
log_debug(){ echo -e "${BLUE}[DEBUG]${NC} $1"; }
trap 'log_error "Failed at line $LINENO"; exit 1' ERR

# ---------- preflight ----------
check_sudo() {
  if [[ ${EUID} -ne 0 ]]; then
    echo -e "${RED}[ERROR]${NC} This script must be run with sudo/root"
    echo "Usage: sudo $0"
    exit 1
  fi
}

show_banner() {
  cat <<'BANNER'

🚀 Shugur Relay - Standalone Installation
==========================================
This will set up:
 • Docker (if missing)
 • CockroachDB (single-node)
 • Shugur Relay (prebuilt image)
 • Caddy (reverse proxy + HTTPS)

⚠️  Requires sudo/root

BANNER
}

detect_os() {
  if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    OS_ID="$ID"
    OS_NAME="$NAME"
  else
    OS_ID="unknown"
    OS_NAME="unknown"
  fi
  log_info "Detected OS: ${OS_NAME} (${OS_ID})"
}

check_docker() {
  if command -v docker >/dev/null 2>&1; then
    log_info "Docker is already installed"
    return 0
  fi
  log_warn "Docker not found; will install"
  return 1
}

install_docker() {
  log_info "Installing Docker..."
  case "$OS_ID" in
    ubuntu|debian)
      apt-get update -y
      apt-get install -y ca-certificates curl gnupg lsb-release
      install -d -m 0755 /etc/apt/keyrings
      curl -fsSL https://download.docker.com/linux/$OS_ID/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
      echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS_ID \
        $(. /etc/os-release; echo "$VERSION_CODENAME") stable" \
        > /etc/apt/sources.list.d/docker.list
      apt-get update -y
      apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
      systemctl enable --now docker
      if [[ -n "${SUDO_USER:-}" ]]; then
        usermod -aG docker "$SUDO_USER" || true
      fi
      ;;
    rhel|centos|rocky|almalinux|fedora)
      if command -v dnf >/dev/null 2>&1; then PM=dnf; else PM=yum; fi
      $PM -y install yum-utils ca-certificates curl
      $PM config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
      $PM -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin
      systemctl enable --now docker
      if [[ -n "${SUDO_USER:-}" ]]; then
        usermod -aG docker "$SUDO_USER" || true
      fi
      ;;
    *)
      log_error "Unsupported/unknown OS for auto Docker install (${OS_ID}). Install Docker manually."
      exit 1
      ;;
  esac
  log_info "Docker installed."
}

ensure_docker_running() {
  if ! docker info >/dev/null 2>&1; then
    log_warn "Docker daemon not running; attempting to start..."
    systemctl start docker || true
    sleep 2
    docker info >/dev/null 2>&1 || { log_error "Docker daemon is not running"; exit 1; }
  fi
}

# ---------- config helpers ----------
get_server_url() {
  echo "" >&2
  echo "🌐 Server Configuration" >&2
  echo "Enter your server domain (or 'localhost' or an IP for testing):" >&2
  read -r server_url
  [[ -z "$server_url" ]] && server_url="localhost"
  echo "$server_url"
}

is_ip() {
  local ip=$1
  [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]
}

create_caddyfile() {
  local server_url="$1"
  log_info "Creating Caddyfile for: $server_url"

  # For localhost or IPs, avoid public ACME; use HTTP or internal CA.
  if [[ "$server_url" == "localhost" ]]; then
    cat > Caddyfile <<'EOF'
http://localhost {
    handle /api/* {
        reverse_proxy relay:8080
    }
    handle {
        reverse_proxy relay:8080 {
            header_up Host {host}
            header_up X-Real-IP {remote}
            header_up X-Forwarded-For {remote}
            header_up X-Forwarded-Proto {scheme}
        }
    }
    encode gzip zstd
    header {
        Strict-Transport-Security "max-age=0"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        Referrer-Policy "strict-origin-when-cross-origin"
        -Server
        -X-Powered-By
    }
    log {
        output file /var/log/caddy/access.log {
            roll_size 100mb
            roll_keep 10
        }
        format json
    }
}
EOF
  elif is_ip "$server_url"; then
    cat > Caddyfile <<EOF
$server_url {
    tls internal
    handle /api/* {
        reverse_proxy relay:8080
    }
    handle {
        reverse_proxy relay:8080 {
            header_up Host {host}
            header_up X-Real-IP {remote}
            header_up X-Forwarded-For {remote}
            header_up X-Forwarded-Proto {scheme}
        }
    }
    encode gzip zstd
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        Referrer-Policy "strict-origin-when-cross-origin"
        -Server
        -X-Powered-By
    }
    log {
        output file /var/log/caddy/access.log {
            roll_size 100mb
            roll_keep 10
        }
        format json
    }
}
EOF
  else
    cat > Caddyfile <<EOF
$server_url {
    handle /api/* {
        reverse_proxy relay:8080
    }
    handle {
        reverse_proxy relay:8080 {
            header_up Host {host}
            header_up X-Real-IP {remote}
            header_up X-Forwarded-For {remote}
            header_up X-Forwarded-Proto {scheme}
        }
    }
    encode gzip zstd
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        Referrer-Policy "strict-origin-when-cross-origin"
        -Server
        -X-Powered-By
    }
    log {
        output file /var/log/caddy/access.log {
            roll_size 100mb
            roll_keep 10
        }
        format json
    }
}
EOF
  fi
  log_info "Caddyfile created."
}

create_config_file() {
  local config_file="./config.yaml"
  log_info "Creating relay config: $config_file"
  cat > "$config_file" <<'EOF'
# Shugur Relay Configuration (standalone)

GENERAL: {}

LOGGING:
  LEVEL: info
  FILE:
  FORMAT: json
  MAX_SIZE: 20
  MAX_BACKUPS: 10
  MAX_AGE: 14

METRICS:
  ENABLED: true
  PORT: 8181

RELAY:
  NAME: "shugur-relay"
  DESCRIPTION: "High-performance, reliable, scalable Nostr relay for decentralized communication."
  CONTACT: "admin@shugur.com"
  ICON: "https://github.com/Shugur-Network/relay/raw/main/logo.png"
  BANNER: "https://github.com/Shugur-Network/relay/raw/main/banner.png"
  WS_ADDR: ":8080"
  PUBLIC_URL: ""
  EVENT_CACHE_SIZE: 10000
  SEND_BUFFER_SIZE: 8192
  WRITE_TIMEOUT: 60s
  IDLE_TIMEOUT: 300s
  THROTTLING:
    MAX_CONTENT_LENGTH: 2048
    MAX_CONNECTIONS: 1000
    BAN_THRESHOLD: 5
    BAN_DURATION: 5
    RATE_LIMIT:
      ENABLED: true
      MAX_EVENTS_PER_SECOND: 50
      MAX_REQUESTS_PER_SECOND: 100
      BURST_SIZE: 20
      PROGRESSIVE_BAN: true
      MAX_BAN_DURATION: 24h

RELAY_POLICY:
  BLACKLIST:
    PUBKEYS: []
  WHITELIST:
    PUBKEYS: []

DATABASE:
  SERVER: "cockroachdb"
  PORT: 26257
EOF
}

update_config_with_server_url() {
  local server_url="$1"
  local config_file="./config.yaml"
  local public_url
  if [[ "$server_url" == "localhost" ]]; then
    public_url="ws://localhost"
  elif is_ip "$server_url"; then
    public_url="wss://$server_url"
  else
    public_url="wss://$server_url"
  fi
  log_info "Setting PUBLIC_URL=$public_url and NAME=$server_url"
  sed -i "s|PUBLIC_URL: \"\"|PUBLIC_URL: \"$public_url\"|g" "$config_file"
  sed -i "s|NAME: \"shugur-relay\"|NAME: \"$server_url\"|g" "$config_file"
}

create_complete_compose_file() {
  local compose_file="docker-compose.standalone.yml"
  log_info "Writing compose file: $compose_file"
  cat > "$compose_file" <<'EOF'
# Shugur Relay - Complete Standalone (Cockroach + Relay + Caddy)

services:
  cockroachdb:
    image: cockroachdb/cockroach:latest
    container_name: cockroachdb
    command: start-single-node --insecure
    volumes:
      - cockroach_volume:/cockroach/cockroach-data
    ports:
      - "26257:26257"  # SQL
      - "26258:26258"  # RPC
      - "9090:8080"    # Admin UI (host 9090 -> container 8080)
    healthcheck:
      test: ["CMD-SHELL", "curl -fsS http://localhost:8080/health?ready=1 >/dev/null"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 30s
    networks:
      - relay_network

  relay:
    image: ghcr.io/shugur-network/relay:latest
    container_name: relay
    restart: unless-stopped
    environment:
      - SHUGUR_ENV=production
      - SHUGUR_DB_HOST=cockroachdb
      - SHUGUR_DB_PORT=26257
      - SHUGUR_DB_USER=root
      - SHUGUR_DB_SSL_MODE=disable
      - SHUGUR_LOG_LEVEL=info
      - SHUGUR_LOG_FORMAT=json
      - SHUGUR_METRICS_ENABLED=true
      - SHUGUR_WS_PORT=8080
      - SHUGUR_METRICS_PORT=8181
      - SHUGUR_MAX_CONNECTIONS=100
      - SHUGUR_RATE_LIMIT=20
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./logs:/app/logs
    depends_on:
      cockroachdb:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "(command -v wget >/dev/null && wget -q --spider http://localhost:8080/api/info) || (command -v curl >/dev/null && curl -fsI http://localhost:8080/api/info)"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
    networks:
      - relay_network

  caddy:
    image: caddy:latest
    container_name: caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
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
    networks:
      - relay_network

volumes:
  cockroach_volume:
    driver: local
  caddy_data:
    driver: local
  caddy_config:
    driver: local

networks:
  relay_network:
    driver: bridge
EOF
}

start_all_services() {
  local compose_file="docker-compose.standalone.yml"

  log_info "Starting CockroachDB..."
  docker compose -f "$compose_file" up -d cockroachdb

  log_info "Waiting for CockroachDB to be ready..."
  local max_attempts=30 attempt=0
  while (( attempt < max_attempts )); do
    if docker compose -f "$compose_file" exec -T cockroachdb /cockroach/cockroach sql --insecure --execute="SELECT 1;" >/dev/null 2>&1; then
      log_info "CockroachDB is ready."
      break
    fi
    attempt=$((attempt+1))
    log_debug "Attempt $attempt/$max_attempts..."
    sleep 2
  done
  if (( attempt == max_attempts )); then
    log_error "CockroachDB failed to become ready."
    exit 1
  fi

  log_info "Starting Relay..."
  docker compose -f "$compose_file" up -d relay

  log_info "Waiting for Relay (best-effort)..."
  max_attempts=30; attempt=0
  while (( attempt < max_attempts )); do
    if docker compose -f "$compose_file" exec -T relay sh -c '(command -v wget >/dev/null && wget -q --spider http://localhost:8080/api/info) || (command -v curl >/dev/null && curl -fsI http://localhost:8080/api/info)' >/dev/null 2>&1; then
      log_info "Relay is responding."
      break
    fi
    attempt=$((attempt+1))
    log_debug "Attempt $attempt/$max_attempts..."
    sleep 3
  done
  if (( attempt == max_attempts )); then
    log_warn "Relay health check timed out; continuing with Caddy startup."
  fi

  log_info "Starting Caddy..."
  docker compose -f "$compose_file" up -d caddy
  sleep 5
  if docker compose -f "$compose_file" ps caddy | grep -q "Up"; then
    log_info "Caddy started."
  else
    log_warn "Caddy may have issues; check: docker compose -f $compose_file logs caddy"
  fi

  log_info "All services are (attempted) up."
}

show_completion_message() {
  local server_url="$1"
  echo
  echo "🎉 Installation Complete"
  echo "========================"
  log_info "✅ CockroachDB (single-node) | ✅ Relay | ✅ Caddy"

  if [[ "$server_url" == "localhost" ]]; then
    log_info "Relay:            http://localhost"
    log_info "WebSocket:        ws://localhost"
  else
    log_info "Relay:            https://$server_url"
    log_info "WebSocket:        wss://$server_url"
  fi
  log_info "Cockroach Admin:  http://localhost:9090"
  echo
  log_info "📊 Management:"
  log_info "  docker compose -f docker-compose.standalone.yml logs -f"
  log_info "  docker compose -f docker-compose.standalone.yml restart"
  log_info "  docker compose -f docker-compose.standalone.yml down"
  log_info "  docker compose -f docker-compose.standalone.yml pull relay && docker compose -f docker-compose.standalone.yml restart relay"
  echo
  log_info "Security:"
  log_info "  • Localhost/IP uses HTTP or internal TLS; domains use public HTTPS."
  log_info "  • Metrics port (8181) is not exposed via Caddy."
  echo
  log_info "Repo: https://github.com/Shugur-Network/Relay"
  echo
  log_info "💡 Installation complete! Configuration files are ready in the current directory."
}

# ---------- main ----------
main() {
  check_sudo
  show_banner
  detect_os

  if ! check_docker; then
    install_docker
    log_info "Re-login may be required to use Docker without sudo."
  fi
  ensure_docker_running

  log_info "Step 1: Server configuration..."
  server_url=$(get_server_url)
  log_info "Using server URL: $server_url"

  log_info "Step 2: Prepare files..."
  mkdir -p logs logs/caddy
  create_caddyfile "$server_url"
  create_config_file
  update_config_with_server_url "$server_url"
  create_complete_compose_file

  log_info "Step 3: Start services..."
  start_all_services

  show_completion_message "$server_url"
}

# Support piped mode (URL from stdin)
if [ -t 0 ]; then
  main
else
  check_sudo
  server_url=$(cat); [[ -z "$server_url" ]] && server_url="localhost"
  show_banner
  detect_os
  if ! check_docker; then
    log_error "Docker not found. Install Docker first."
    exit 1
  fi
  ensure_docker_running
  mkdir -p logs logs/caddy
  create_caddyfile "$server_url"
  create_config_file
  update_config_with_server_url "$server_url"
  create_complete_compose_file
  start_all_services
  show_completion_message "$server_url"
fi
