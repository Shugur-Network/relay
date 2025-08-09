# Manual Installation: Docker Distributed

This method provides a step-by-step guide for setting up a **multi-node, distributed** relay with Docker. It mirrors the functionality of the automated script but grants you full control over every aspect of the setup. This is ideal for production environments where you need to manage the configuration manually.

## Overview

The process is divided into three main parts:

1. **Local Certificate Generation**: You will act as the Certificate Authority (CA) on your local machine, creating all necessary TLS certificates for the cluster.
2. **Per-Node Configuration**: You will connect to each of your servers, create the required directory structure, and copy over the certificates and configuration files.
3. **Cluster Bootstrap**: You will start the services in a specific order to initialize the distributed database and bring the entire relay cluster online.

---

## Part 1: Certificate Generation (On Your Local Machine)

First, you need to generate TLS certificates for secure communication between the CockroachDB nodes and for the relay to connect securely to the database.

### 1. Prepare the Workspace

Create a dedicated directory on your local computer to hold the certificates.

```bash
mkdir -p shugur-certs/safe-dir
cd shugur-certs
```

### 2. Create the Certificate Authority (CA)

The CA will be used to sign all other certificates.

Create a configuration file for your CA named `ca.cnf`:

```ini
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
```

Now, generate the CA key and certificate:

```bash
# Create the CA key in the safe directory
openssl genrsa -out safe-dir/ca.key 2048
chmod 400 safe-dir/ca.key

# Create the CA certificate
openssl req -new -x509 -config ca.cnf -key safe-dir/ca.key -out ca.crt -days 3650 -batch

# Initialize database and serial files for signing
touch index.txt
echo '01' > serial.txt
```

### 3. Generate Node Certificates

Repeat this process for **each server** in your cluster.

For each node, create a config file `node-X.cnf` (e.g., `node-1.cnf`), replacing `YOUR_NODE_IP_OR_DOMAIN` with the actual public IP or domain name of that server.

```ini
# OpenSSL node configuration file for node X
[ req ]
prompt=no
distinguished_name = distinguished_name
req_extensions = extensions

[ distinguished_name ]
organizationName = Cockroach

[ extensions ]
subjectAltName = critical,DNS:node,DNS:localhost,IP:127.0.0.1,DNS:YOUR_NODE_IP_OR_DOMAIN,IP:YOUR_NODE_IP_OR_DOMAIN
```

Now, generate, sign, and verify the certificate for that node (e.g., for `node-1`):

```bash
# Generate the key
openssl genrsa -out node-1.key 2048
chmod 400 node-1.key

# Create the Certificate Signing Request (CSR)
openssl req -new -config node-1.cnf -key node-1.key -out node-1.csr -batch

# Sign the certificate with your CA
openssl ca -config ca.cnf -keyfile safe-dir/ca.key -cert ca.crt -policy signing_policy -extensions signing_node_req -out node-1.crt -outdir . -in node-1.csr -batch
```

**Remember to repeat this for `node-2`, `node-3`, etc.**

### 4. Generate Client Certificates

Create certificates for the `root` and `relay` users to connect to the database.

Create a generic `client.cnf` file:

```ini
[ req ]
prompt=no
distinguished_name = distinguished_name
req_extensions = extensions

[ distinguished_name ]
organizationName = Cockroach
commonName = root

[ extensions ]
subjectAltName = DNS:root
```

Now, generate certificates for both users:

```bash
# For the 'root' user
openssl genrsa -out client.root.key 2048
chmod 400 client.root.key
openssl req -new -config client.cnf -key client.root.key -out client.root.csr -batch
openssl ca -config ca.cnf -keyfile safe-dir/ca.key -cert ca.crt -policy signing_policy -extensions signing_client_req -out client.root.crt -in client.root.csr -batch

# For the 'relay' user
sed -i 's/commonName = root/commonName = relay/' client.cnf
sed -i 's/subjectAltName = DNS:root/subjectAltName = DNS:relay/' client.cnf
openssl genrsa -out client.relay.key 2048
chmod 400 client.relay.key
openssl req -new -config client.cnf -key client.relay.key -out client.relay.csr -batch
openssl ca -config ca.cnf -keyfile safe-dir/ca.key -cert ca.crt -policy signing_policy -extensions signing_client_req -out client.relay.crt -in client.relay.csr -batch
```

At the end of this part, your `shugur-certs` directory should contain the CA, node certificates/keys, and client certificates/keys.

---

## Part 2: Per-Node Server Configuration

You must perform these steps on **every server** that will be part of your distributed relay.

### 1. Install Prerequisites

Ensure Docker and Docker Compose are installed on each server.

```bash
# Install Docker Engine
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

### 2. Create Directory Structure

Create the necessary directories on each server.

```bash
mkdir -p ~/shugur-relay/{config,certs/{cockroach,relay},logs/{relay,cockroachdb,caddy}}
```

### 3. Copy Certificates

From your local machine, securely copy the generated certificates to each server.

For **Node 1** (replace `USER` and `NODE_1_IP`):

```bash
# General certs
scp ca.crt client.root.* client.relay.* USER@NODE_1_IP:~/shugur-relay/
# Node-specific cert
scp node-1.crt node-1.key USER@NODE_1_IP:~/shugur-relay/
```

Repeat for **Node 2**, **Node 3**, etc., using their respective `node-X` certificates.

Once copied, organize the certificates on **each server**:

```bash
cd ~/shugur-relay

# CockroachDB certs
mv ca.crt certs/cockroach/
mv node-*.crt certs/cockroach/node.crt # Rename to generic 'node.crt'
mv node-*.key certs/cockroach/node.key # Rename to generic 'node.key'
mv client.root.crt certs/cockroach/
mv client.root.key certs/cockroach/

# Relay client certs
cp certs/cockroach/ca.crt certs/relay/
cp certs/cockroach/client.root.crt certs/relay/
cp certs/cockroach/client.root.key certs/relay/
mv client.relay.crt certs/relay/
mv client.relay.key certs/relay/

# Set secure permissions
chmod 600 certs/cockroach/*.key certs/relay/*.key
```

### 4. Create Configuration Files

On each server, you need three configuration files inside `~/shugur-relay/config/`.

**`config.yaml`**:

```yaml
# ~/shugur-relay/config/config.yaml
GENERAL: {}
LOGGING:
  LEVEL: info
DATABASE:
  SERVER: "localhost"
  PORT: 26257
RELAY:
  NAME: "Distributed Relay Node (CHANGE ME)"
  DESCRIPTION: "A node in the Shugur Relay distributed network."
  CONTACT: "operator@example.com"
  # ... other settings
```

**`Caddyfile`** (replace `your-domain-for-this-node.com`):

```caddy
# ~/shugur-relay/config/Caddyfile
your-domain-for-this-node.com {
    handle {
        reverse_proxy localhost:8080
    }
    # ... other Caddy settings from the script
}
```

**`docker-compose.yml`** (This is the most critical file):
Replace `NODE_1_IP`, `NODE_2_IP`, `NODE_3_IP` with the actual IPs of ALL your servers. The `--join` list must be identical on all nodes.

```yaml
# ~/shugur-relay/config/docker-compose.yml
services:
  cockroachdb:
    image: cockroachdb/cockroach:latest
    container_name: cockroachdb
    command: start --certs-dir=/cockroach/certs --listen-addr=0.0.0.0:26258 --sql-addr=0.0.0.0:26257 --advertise-addr=$(hostname -i):26258 --join=NODE_1_IP:26258,NODE_2_IP:26258,NODE_3_IP:26258
    volumes:
      - ../cockroach_data:/cockroach/cockroach-data
      - ../certs/cockroach:/cockroach/certs:ro
    ports:
      - "26257:26257"
      - "26258:26258"
      - "9090:8080" # Admin UI

  relay:
    image: ghcr.io/shugur-network/relay:latest
    # ... relay service definition from script ...
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ../certs/relay:/app/certs:ro
    network_mode: host
    depends_on:
      - cockroachdb

  caddy:
    image: caddy:latest
    # ... caddy service definition from script ...
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
    network_mode: host
    depends_on:
      - relay

volumes:
  cockroach_data:
```

---

## Part 3: Cluster Deployment and Initialization

This part requires careful sequencing.

### 1. Start the First Node

On **Node 1 only**, start the database service:

```bash
cd ~/shugur-relay/config
docker compose up -d cockroachdb
```

### 2. Initialize the Cluster

Wait a few moments, then on **Node 1 only**, initialize the cluster:

```bash
docker compose exec cockroachdb /cockroach/cockroach init --certs-dir=/cockroach/certs --host=localhost:26258
```

### 3. Start Remaining Database Nodes

On **all other nodes (Node 2, Node 3, etc.)**, start their database services. They will automatically join the cluster initialized by Node 1.

```bash
cd ~/shugur-relay/config
docker compose up -d cockroachdb
```

### 4. Create Database and User

Back on **Node 1**, create the database and user for the relay:

```bash
docker compose exec cockroachdb /cockroach/cockroach sql --certs-dir=/cockroach/certs --host=localhost:26257 --execute="CREATE DATABASE IF NOT EXISTS shugur_relay; CREATE USER IF NOT EXISTS relay; GRANT ALL ON DATABASE shugur_relay TO relay;"
```

### 5. Start All Services

On **all nodes** (including Node 1), bring up the rest of the services (relay and Caddy):

```bash
cd ~/shugur-relay/config
docker compose up -d
```

Your distributed relay cluster is now running! You can verify the status on each node with `docker compose ps` and check the CockroachDB UI at `http://<any_node_ip>:9090`.
