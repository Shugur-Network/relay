# Troubleshooting Guide

This guide helps resolve common issues when deploying and operating Shugur Relay.

## Common Issues

### 1. Relay Won't Start

#### Symptoms

- Service fails to start
- Connection refused errors
- Immediate exit after startup

#### Diagnosis

```bash
# Check service status
sudo systemctl status shugur-relay

# View recent logs
sudo journalctl -u shugur-relay -n 50

# Check configuration
shugur-relay --config /path/to/config.yaml --validate
```

#### Common Causes

- **Database Connection Failed**: Check CockroachDB is running and accessible
- **Port Already in Use**: Another service is using port 8080
- **Invalid Configuration**: YAML syntax errors or invalid values
- **Permission Issues**: Insufficient file permissions

#### Solutions

```bash
# Check if port is in use
sudo netstat -tlnp | grep :8080

# Validate YAML syntax
yamllint config.yaml

# Test database connection
cockroach sql --certs-dir=/path/to/certs --host=localhost:26257 --execute="SELECT 1;"

# Fix permissions
sudo chown -R shugur:shugur /opt/shugur-relay
sudo chmod 600 /opt/shugur-relay/certs/*.key
```

### 2. Database Connection Issues

#### Symptoms

- "connection refused" errors
- "certificate verify failed" errors
- Slow queries or timeouts

#### Diagnosis

```bash
# Check CockroachDB status
sudo systemctl status cockroachdb

# Test connection manually
cockroach sql --certs-dir=/var/lib/cockroach/certs --host=localhost:26257

# Check database logs
sudo journalctl -u cockroachdb -f
```

#### Common Causes

- **CockroachDB Not Running**: Service stopped or crashed
- **Certificate Issues**: Invalid or expired certificates
- **Network Issues**: Firewall blocking database ports
- **Resource Constraints**: Insufficient memory or disk space

#### Solutions

```bash
# Restart CockroachDB
sudo systemctl restart cockroachdb

# Regenerate certificates
sudo -u cockroach cockroach cert create-node localhost $(hostname) \
  --certs-dir=/var/lib/cockroach/certs \
  --ca-key=/var/lib/cockroach/certs/ca.key

# Check firewall
sudo ufw status
sudo ufw allow 26257/tcp

# Monitor resources
htop
df -h
```

### 3. WebSocket Connection Problems

#### Symptoms

- Clients can't connect
- Frequent disconnections
- Timeout errors

#### Diagnosis

```bash
# Test WebSocket directly
wscat -c ws://localhost:8080

# Check connection limits
netstat -an | grep :8080 | wc -l

# Check relay statistics
curl http://localhost:8080/api/stats
```

#### Common Causes

- **Reverse Proxy Issues**: Caddy/Nginx misconfiguration
- **Firewall Blocking**: Port 8080 not accessible
- **Connection Limits**: Too many concurrent connections
- **SSL/TLS Issues**: Certificate problems with WSS

#### Solutions

```bash
# Test without reverse proxy
curl -H "Upgrade: websocket" http://localhost:8080

# Check reverse proxy config
sudo caddy validate --config /etc/caddy/Caddyfile

# Increase connection limits
ulimit -n 65536

# Fix SSL certificates
sudo caddy reload
```

### 4. High Memory Usage

#### Symptoms

- OOM (Out of Memory) kills
- Slow performance
- Swap usage increasing

#### Diagnosis

```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head

# Monitor relay memory
top -p $(pgrep shugur-relay)

# Check relay statistics
curl http://localhost:8080/api/stats
```

#### Solutions

```yaml
# Reduce cache size in config.yaml
RELAY:
  EVENT_CACHE_SIZE: 5000  # Reduce from default 10000

# Tune database connections
DATABASE:
  MAX_OPEN_CONNS: 10      # Reduce from default 25
  MAX_IDLE_CONNS: 2       # Reduce from default 5
```

### 5. Slow Query Performance

#### Symptoms

- Delayed event retrieval
- Client timeouts
- High database CPU usage

#### Diagnosis

```bash
# Check database performance
cockroach sql --certs-dir=/certs --host=localhost:26257 --execute="
SHOW QUERIES;
"

# Monitor query times
cockroach sql --certs-dir=/certs --host=localhost:26257 --execute="
SELECT * FROM [SHOW CLUSTER SETTING sql.stats.automatic_collection.enabled];
"

# Check indexes
cockroach sql --certs-dir=/certs --host=localhost:26257 --execute="
SHOW INDEXES FROM shugur_relay.events;
"
```

#### Solutions

```sql
-- Create additional indexes for common queries
CREATE INDEX CONCURRENTLY idx_events_kind_created_at 
ON shugur_relay.events (kind, created_at DESC);

CREATE INDEX CONCURRENTLY idx_events_pubkey_created_at 
ON shugur_relay.events (pubkey, created_at DESC);

-- Update table statistics
ANALYZE shugur_relay.events;
```

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Connection Count**: `relay_connections_active`
2. **Event Rate**: `relay_events_received_total`
3. **Error Rate**: `relay_errors_total`
4. **Database Performance**: `relay_db_query_duration`
5. **Memory Usage**: `process_resident_memory_bytes`

### Sample Prometheus Alerts

```yaml
groups:
- name: shugur-relay
  rules:
  - alert: RelayDown
    expr: up{job="shugur-relay"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Shugur Relay is down"

  - alert: HighErrorRate
    expr: rate(relay_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"

  - alert: DatabaseConnectionFailed
    expr: relay_db_connections_failed_total > 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database connection failures"
```

## Performance Tuning

### Operating System

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize network settings
echo "net.core.somaxconn = 65536" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" >> /etc/sysctl.conf
sysctl -p
```

### Database Optimization

```sql
-- Increase cache size
SET CLUSTER SETTING kv.snapshot_rebalance.max_rate = '64MiB';
SET CLUSTER SETTING kv.snapshot_recovery.max_rate = '64MiB';

-- Optimize memory settings
SET CLUSTER SETTING sql.defaults.default_int_size = 4;
```

### Application Tuning

```yaml
# config.yaml optimizations
RELAY:
  SEND_BUFFER_SIZE: 16384      # Increase buffer size
  EVENT_CACHE_SIZE: 20000      # Increase cache if you have RAM
  
DATABASE:
  MAX_OPEN_CONNS: 50           # Increase for high traffic
  CONN_MAX_LIFETIME: 15m       # Reduce connection lifetime
```

## Logging and Debugging

### Enable Debug Logging

```yaml
LOGGING:
  LEVEL: debug
  FORMAT: json
  FILE: /var/log/shugur-relay/debug.log
```

### Useful Log Queries

```bash
# Find connection errors
grep "connection" /var/log/shugur-relay/relay.log | grep ERROR

# Monitor event processing
grep "event_stored" /var/log/shugur-relay/relay.log | tail -100

# Check rate limiting
grep "rate_limit" /var/log/shugur-relay/relay.log
```

## Getting Help

### Before Seeking Support

1. Check this troubleshooting guide
2. Review the configuration documentation
3. Check system resources (CPU, memory, disk)
4. Collect relevant logs and metrics

### Information to Include

- Shugur Relay version (`shugur-relay --version`)
- Operating system and version
- Configuration file (remove sensitive data)
- Error messages and logs
- System resource usage
- Network topology (if distributed setup)

### Support Channels

- **GitHub Issues**: <https://github.com/Shugur-Network/Relay/issues>
- **Documentation**: <https://github.com/Shugur-Network/Relay/docs>
- **Community**: Check README for community links
