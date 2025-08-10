# Configuration Parameters

Shugur Relay can be configured via a YAML file (`config.yaml`) or through environment variables.

- **YAML file**: By default, the application looks for a `config.yaml` in the current directory. You can specify a different path using the `--config` flag when running the application.
- **Environment Variables**: All parameters can be set using environment variables. They must be prefixed with `SHUGUR_`, and nested keys are separated by an underscore. For example, `RELAY.NAME` becomes `SHUGUR_RELAY_NAME`, and `DATABASE.SERVER` becomes `SHUGUR_DATABASE_SERVER`.

**Note**: Some Docker deployment scripts may use simplified environment variable names (like `SHUGUR_DB_HOST` or `SHUGUR_LOG_LEVEL`) for convenience. However, the canonical environment variables documented below are what the application actually reads.

The order of precedence is: **Environment Variables > Configuration File > Defaults**.

---

## `LOGGING`

Controls the logging output of the relay.

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `LEVEL` | `SHUGUR_LOGGING_LEVEL` | The minimum level of logs to output. Options: `debug`, `info`, `warn`, `error`, `fatal`. | `debug` |
| `FILE` | `SHUGUR_LOGGING_FILE` | The path to a file where logs should be written. If empty, logs are written to standard output. | (empty) |
| `FORMAT` | `SHUGUR_LOGGING_FORMAT` | The format of the logs. Options: `console` (human-readable) or `json` (machine-readable). | `console` |
| `MAX_SIZE` | `SHUGUR_LOGGING_MAX_SIZE` | The maximum size in megabytes of the log file before it gets rotated. | `20` |
| `MAX_BACKUPS` | `SHUGUR_LOGGING_MAX_BACKUPS` | The maximum number of old log files to retain. | `10` |
| `MAX_AGE` | `SHUGUR_LOGGING_MAX_AGE` | The maximum number of days to retain old log files. | `14` |

---

## `METRICS`

Configuration for Prometheus metrics collection.

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `ENABLED` | `SHUGUR_METRICS_ENABLED` | Set to `true` to enable metrics collection. | `true` |
| `PORT` | `SHUGUR_METRICS_PORT` | The port number for metrics configuration (used in deployment scripts). | `2112` |

**Note**: Metrics are collected internally using Prometheus client libraries. The `PORT` setting is primarily used by deployment configurations and reverse proxy setups. Production deployments typically use port `8181` for metrics access through the reverse proxy.

---

## `RELAY`

Core configuration for the relay's identity and behavior (NIP-11).

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `NAME` | `SHUGUR_RELAY_NAME` | The public name of the relay. Max 30 characters. | `shugur-relay` |
| `DESCRIPTION` | `SHUGUR_RELAY_DESCRIPTION` | A short description of the relay. Max 200 characters. | `High-performance, reliable, scalable Nostr relay...` |
| `CONTACT` | `SHUGUR_RELAY_CONTACT` | An email address for the relay operator. | `support@shugur.com` |
| `ICON` | `SHUGUR_RELAY_ICON` | A URL to an icon for the relay. | (Shugur logo URL) |
| `BANNER` | `SHUGUR_RELAY_BANNER` | A URL to a banner image for the relay (optional). | (empty) |
| `WS_ADDR` | `SHUGUR_RELAY_WS_ADDR` | The listening address and port for the WebSocket server. | `:8080` |
| `PUBLIC_URL` | `SHUGUR_RELAY_PUBLIC_URL` | The publicly accessible WebSocket URL (e.g., `wss://relay.example.com`). | `wss://relay.shugur.net` |
| `IDLE_TIMEOUT` | `SHUGUR_RELAY_IDLE_TIMEOUT` | The duration after which an inactive connection is closed (e.g., `300s`). | `300s` |
| `WRITE_TIMEOUT` | `SHUGUR_RELAY_WRITE_TIMEOUT` | The timeout for writing data to a client (e.g., `60s`). | `60s` |
| `SEND_BUFFER_SIZE` | `SHUGUR_RELAY_SEND_BUFFER_SIZE` | The size of the per-connection send buffer in bytes. | `8192` |
| `EVENT_CACHE_SIZE` | `SHUGUR_RELAY_EVENT_CACHE_SIZE` | The number of recent events to keep in an in-memory cache for faster retrieval. | `10000` |

### `RELAY.THROTTLING`

Settings to prevent abuse and control resource usage.

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `MAX_CONTENT_LENGTH` | `SHUGUR_RELAY_THROTTLING_MAX_CONTENT_LENGTH` | The maximum size of an event in bytes. | `2048` |
| `MAX_CONNECTIONS` | `SHUGUR_RELAY_THROTTLING_MAX_CONNECTIONS` | The maximum number of concurrent WebSocket connections. | `1000` |
| `BAN_THRESHOLD` | `SHUGUR_RELAY_THROTTLING_BAN_THRESHOLD` | The number of violations (e.g., sending oversized events) before a client is banned. | `5` |
| `BAN_DURATION` | `SHUGUR_RELAY_THROTTLING_BAN_DURATION` | The duration of a ban in seconds for general violations. | `5` |

### `RELAY.THROTTLING.RATE_LIMIT`

Specific settings for event and request rate limiting.

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `ENABLED` | `SHUGUR_RELAY_THROTTLING_RATE_LIMIT_ENABLED` | Set to `true` to enable rate limiting. | `true` |
| `MAX_EVENTS_PER_SECOND` | `SHUGUR_RELAY_THROTTLING_RATE_LIMIT_MAX_EVENTS_PER_SECOND` | Max number of `EVENT` messages a client can send per second. | `50` |
| `MAX_REQUESTS_PER_SECOND` | `SHUGUR_RELAY_THROTTLING_RATE_LIMIT_MAX_REQUESTS_PER_SECOND` | Max number of `REQ` messages a client can send per second. | `100` |
| `BURST_SIZE` | `SHUGUR_RELAY_THROTTLING_RATE_LIMIT_BURST_SIZE` | The number of requests allowed to exceed the rate limit in a short burst. | `20` |
| `PROGRESSIVE_BAN` | `SHUGUR_RELAY_THROTTLING_RATE_LIMIT_PROGRESSIVE_BAN` | If `true`, repeated rate limit violations will result in progressively longer bans. | `true` |
| `MAX_BAN_DURATION` | `SHUGUR_RELAY_THROTTLING_RATE_LIMIT_MAX_BAN_DURATION` | The maximum duration for a progressive ban (e.g., `24h`). | `24h` |

---

## `RELAY_POLICY`

Defines rules for accepting or rejecting events based on public keys.

### `RELAY_POLICY.BLACKLIST`

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `PUBKEYS` | `SHUGUR_RELAY_POLICY_BLACKLIST_PUBKEYS` | A comma-separated list of hex-encoded public keys that are not allowed to publish events. | (empty list) |

### `RELAY_POLICY.WHITELIST`

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `PUBKEYS` | `SHUGUR_RELAY_POLICY_WHITELIST_PUBKEYS` | If this list is not empty, only these hex-encoded public keys are allowed to publish events. | (empty list) |

---

## `DATABASE`

Configuration for connecting to the CockroachDB database.

> **🔒 Security Note**: In production environments, always use secure connections with proper SSL certificates. See [Bare Metal Installation](./BARE-METAL.md) for certificate setup details.

| YAML Key | Environment Variable | Description | Default |
| :--- | :--- | :--- | :--- |
| `SERVER` | `SHUGUR_DATABASE_SERVER` | The hostname or IP address of the database server. | `localhost` |
| `PORT` | `SHUGUR_DATABASE_PORT` | The port number for the database server. | `26257` |

**Note**: The relay automatically constructs the database connection string using these parameters. Other connection details are handled automatically:

- **Database Name**: Always uses `shugur`
- **Authentication**: Uses `root` user for all connections
- **SSL Mode**: Automatically determined based on certificate availability (secure/insecure mode)
- **Connection Parameters**: Connection pooling and SSL settings are managed internally

For production deployments, ensure proper certificates are in place for secure connections. See [Bare Metal Installation](./BARE-METAL.md) for certificate setup details.

> **🔗 Related**: See [Installation Guide](./installation/INSTALLATION.md) for setup instructions and [Performance](./PERFORMANCE.md) for database optimization tips.
