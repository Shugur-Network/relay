# API Reference

Shugur Relay provides both WebSocket (Nostr protocol) and HTTP APIs for different use cases.

## WebSocket API (Nostr Protocol)

The primary interface for Nostr clients. Follows the Nostr protocol specification.

### Connection

```javascript
const ws = new WebSocket('wss://your-relay.com');
```

### Message Types

#### EVENT

Send an event to the relay:

```json
["EVENT", {
  "id": "event_id",
  "pubkey": "public_key_hex",
  "created_at": 1672531200,
  "kind": 1,
  "tags": [],
  "content": "Hello, Nostr!",
  "sig": "signature_hex"
}]
```

#### REQ

Subscribe to events:

```json
["REQ", "subscription_id", {
  "kinds": [1],
  "authors": ["pubkey_hex"],
  "since": 1672531200,
  "limit": 50
}]
```

#### CLOSE

Close a subscription:

```json
["CLOSE", "subscription_id"]
```

### Response Types

#### OK

Response to EVENT messages:

```json
["OK", "event_id", true, ""]
```

#### EOSE

End of stored events:

```json
["EOSE", "subscription_id"]
```

#### NOTICE

Error or informational messages:

```json
["NOTICE", "error message"]
```

## HTTP API

Administrative and monitoring endpoints.

### Health Check

**GET** `/health`

Returns relay health status.

**Response:**

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "2h30m15s"
}
```

### Relay Information (NIP-11)

**GET** `/` with `Accept: application/nostr+json`

Returns relay metadata per NIP-11.

**Response:**

```json
{
  "name": "Shugur Relay",
  "description": "High-performance, reliable, scalable Nostr relay",
  "pubkey": "relay_pubkey_hex",
  "contact": "admin@example.com",
  "supported_nips": [1, 2, 9, 11, 15, 16, 20, 22, 33, 40],
  "software": "https://github.com/Shugur-Network/Relay",
  "version": "1.0.0",
  "limitation": {
    "max_message_length": 8192,
    "max_subscriptions": 100,
    "max_filters": 10,
    "max_limit": 1000,
    "max_subid_length": 100,
    "payment_required": false,
    "auth_required": false
  }
}
```

### Metrics

**Note**: Prometheus metrics are collected internally but are not exposed via a direct HTTP endpoint in the application. In production deployments, metrics are typically accessed through the reverse proxy configuration (Caddy) which may expose them on a separate port for monitoring systems.

The relay collects the following metrics internally:

- Active WebSocket connections
- Messages received/sent
- Events processed and stored
- Database operations
- Error rates
- Request processing times

## Rate Limiting

The relay implements rate limiting to prevent abuse:

- **Events**: Default 10 events per second per connection
- **Requests**: Default 20 requests per second per connection
- **Burst**: Default 5 message burst allowed

When rate limits are exceeded, the relay will:

1. Send a NOTICE message
2. Temporarily ban the IP address
3. Close the connection for repeat offenders

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| `blocked` | Event blocked by policy | Event rejected by whitelist/blacklist |
| `pow` | Insufficient proof of work | Event doesn't meet PoW requirements |
| `duplicate` | Event already exists | Event with same ID already stored |
| `ephemeral` | Ephemeral event not stored | Kind 20000-29999 events not persisted |
| `invalid` | Event validation failed | Malformed event or invalid signature |
| `rate-limited` | Rate limit exceeded | Too many messages from client |

## Client Libraries

Compatible with standard Nostr client libraries:

- **JavaScript**: nostr-tools, nostr-js
- **Python**: python-nostr
- **Go**: go-nostr
- **Rust**: nostr-sdk
- **Swift**: nostr-swift

## Examples

> **💡 Tip**: These examples use JavaScript with nostr-tools, but the same concepts apply to other client libraries. Always handle errors and implement proper retry logic in production applications.

### Publishing a Note

```javascript
import { relayInit, getEventHash, signEvent } from 'nostr-tools'

const relay = relayInit('wss://your-relay.com')
await relay.connect()

const event = {
  kind: 1,
  created_at: Math.floor(Date.now() / 1000),
  tags: [],
  content: 'Hello from Shugur Relay!',
  pubkey: yourPublicKey,
}

event.id = getEventHash(event)
event.sig = signEvent(event, yourPrivateKey)

await relay.publish(event)
```

### Subscribing to Events

```javascript
const sub = relay.sub([
  {
    kinds: [1],
    limit: 50
  }
])

sub.on('event', event => {
  console.log('Received event:', event)
})

sub.on('eose', () => {
  console.log('End of stored events')
})

## Next Steps

- Review the [Configuration Guide](./CONFIGURATION.md) to understand how to configure the relay
- Check the [Performance Guide](./PERFORMANCE.md) for optimization recommendations
- See the [Troubleshooting Guide](./TROUBLESHOOTING.md) for common issues and solutions

## Related Documentation

- **[Installation Guide](./installation/INSTALLATION.md)**: Choose your deployment method
- **[Architecture Overview](./ARCHITECTURE.md)**: Understand the system design
- **[Configuration Guide](./CONFIGURATION.md)**: Configure your relay settings
- **[Performance Guide](./PERFORMANCE.md)**: Optimize for production workloads
- **[Troubleshooting Guide](./TROUBLESHOOTING.md)**: Resolve common issues
