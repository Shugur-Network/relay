# Manual Installation: Docker Standalone

This method is for setting up a **single-node** Shugur Relay using Docker. It's ideal for development, testing, or small-scale production environments.

## Step 1: Prepare Your Server

1. Connect to your server.
2. Create a directory for your relay and navigate into it:

    ```bash
    mkdir ~/shugur-relay && cd ~/shugur-relay
    ```

## Step 2: Download Configuration Files

Download the `docker-compose.standalone.yml` and a default `config.yaml` file:

```bash
curl -O https://github.com/Shugur-Network/Relay/raw/main/docker/compose/docker-compose.standalone.yml
curl -O https://github.com/Shugur-Network/Relay/raw/main/config/development.yaml
mv development.yaml config.yaml
```

## Step 3: Customize Your Configuration

Open `config.yaml` with a text editor to adjust settings like the relay name, description, and retention policies.

```bash
nano config.yaml
```

## Step 4: Deploy the Relay

Start the services using Docker Compose:

```bash
docker compose -f docker-compose.standalone.yml up -d
```

## Step 5: Initialize the Database

After the containers are running, initialize the CockroachDB database:

```bash
docker compose -f docker-compose.standalone.yml exec cockroachdb ./cockroach sql --insecure --execute="CREATE DATABASE IF NOT EXISTS shugur_relay; CREATE USER IF NOT EXISTS relay; GRANT ALL ON DATABASE shugur_relay TO relay;"
```

Your standalone relay is now running. You can check its status with `docker compose ps`.
