# Performance Guide

This guide provides recommendations for optimizing Shugur Relay performance in production environments.

## Capacity Planning

### Hardware Recommendations

#### Standalone Deployment

| Load Level | CPU | RAM | Storage | Network |
|------------|-----|-----|---------|---------|
| **Light** (< 1K events/day) | 2 cores | 4 GB | 50 GB SSD | 100 Mbps |
| **Medium** (< 100K events/day) | 4 cores | 8 GB | 200 GB SSD | 1 Gbps |
| **Heavy** (< 1M events/day) | 8 cores | 16 GB | 500 GB NVMe | 10 Gbps |
| **Enterprise** (> 1M events/day) | 16+ cores | 32+ GB | 1+ TB NVMe | 10+ Gbps |

#### Distributed Deployment (per node)

| Component | CPU | RAM | Storage | Notes |
|-----------|-----|-----|---------|-------|
| **Relay Node** | 4-8 cores | 8-16 GB | 50 GB SSD | Stateless, can scale horizontally |
| **Database Node** | 8-16 cores | 16-64 GB | 500 GB+ NVMe | Storage grows with data retention |
| **Load Balancer** | 2-4 cores | 4-8 GB | 20 GB SSD | Nginx/HAProxy/Caddy |

### Network Requirements

- **Latency**: < 10ms between database nodes
- **Bandwidth**: 100 Mbps minimum per 1000 concurrent connections
- **IPv6**: Recommended for global accessibility

## Configuration Optimization

### Relay Configuration

```yaml
# High-performance relay configuration
RELAY:
  EVENT_CACHE_SIZE: 50000        # Larger cache for better hit rates
  SEND_BUFFER_SIZE: 32768        # Larger buffers for throughput
  WRITE_TIMEOUT: 30s             # Reduced timeout for faster cleanup
  IDLE_TIMEOUT: 300s             # Keep connections alive longer
  
  THROTTLING:
    MAX_CONNECTIONS: 5000        # Increase for high traffic
    MAX_CONTENT_LENGTH: 16384    # Allow larger events if needed
    
    RATE_LIMIT:
      MAX_EVENTS_PER_SECOND: 100 # Adjust based on capacity
      MAX_REQUESTS_PER_SECOND: 200
      BURST_SIZE: 50             # Allow larger bursts
      
DATABASE:
  SERVER: "cockroachdb"          # Database hostname
  PORT: 26257                    # Database port
  
METRICS:
  ENABLED: true                  # Essential for monitoring
  PORT: 8181                     # Production metrics port
```

**Note**: Database connection pooling and SSL settings are managed automatically by the application based on certificate availability and environment.

### Database Optimization

#### CockroachDB Settings

```sql
-- Memory and cache settings
SET CLUSTER SETTING cluster.preserve_downgrade_option = '';
SET CLUSTER SETTING sql.stats.automatic_collection.enabled = true;
SET CLUSTER SETTING sql.stats.histogram_collection.enabled = true;

-- Performance tuning
SET CLUSTER SETTING kv.range_merge.queue_interval = '1s';
SET CLUSTER SETTING kv.snapshot_rebalance.max_rate = '128MiB';
SET CLUSTER SETTING kv.snapshot_recovery.max_rate = '128MiB';

-- Connection pooling
SET CLUSTER SETTING server.max_connections_per_gateway = 500;
```

#### Storage Layout

```sql
-- Optimize table storage for events
ALTER TABLE shugur.events CONFIGURE ZONE USING
  range_min_bytes = 134217728,  -- 128MB
  range_max_bytes = 536870912,  -- 512MB
  gc.ttlseconds = 604800;       -- 7 days GC

-- Separate hot and cold data
CREATE TABLE shugur.events_archive AS 
SELECT * FROM shugur.events WHERE created_at < extract(epoch from now() - interval '30 days');
```

## Operating System Optimization

### Linux Kernel Parameters

```bash
# /etc/sysctl.conf optimizations
# Network performance
net.core.somaxconn = 65536
net.core.netdev_max_backlog = 30000
net.ipv4.tcp_max_syn_backlog = 65536
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 87380 16777216

# File descriptors
fs.file-max = 1000000

# Virtual memory
vm.swappiness = 1
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5

# Apply changes
sysctl -p
```

### System Limits

```bash
# /etc/security/limits.conf
* soft nofile 1000000
* hard nofile 1000000
* soft nproc 1000000
* hard nproc 1000000

# /etc/systemd/system.conf
DefaultLimitNOFILE=1000000
DefaultLimitNPROC=1000000
```

### CPU Optimization

```bash
# Set CPU governor to performance
echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

# Disable CPU frequency scaling
systemctl disable cpufreq

# NUMA optimization (if applicable)
echo never > /sys/kernel/mm/transparent_hugepage/enabled
```

## Monitoring and Metrics

### Key Performance Indicators

1. **Throughput Metrics**
   - Events per second: `rate(relay_events_received_total[1m])`
   - Connections per second: `rate(relay_connections_total[1m])`
   - Database queries per second: `rate(relay_db_queries_total[1m])`

2. **Latency Metrics**
   - Event processing time: `relay_event_processing_duration`
   - Database query time: `relay_db_query_duration`
   - WebSocket response time: `relay_websocket_response_duration`

3. **Resource Utilization**
   - CPU usage: `rate(process_cpu_seconds_total[1m])`
   - Memory usage: `process_resident_memory_bytes`
   - Disk I/O: `rate(node_disk_io_time_seconds_total[1m])`

4. **Error Rates**
   - Connection failures: `rate(relay_connections_failed_total[1m])`
   - Event rejections: `rate(relay_events_rejected_total[1m])`
   - Database errors: `rate(relay_db_errors_total[1m])`

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Shugur Relay Performance",
    "panels": [
      {
        "title": "Events per Second",
        "targets": [
          {
            "expr": "rate(relay_events_received_total[1m])"
          }
        ]
      },
      {
        "title": "Active Connections",
        "targets": [
          {
            "expr": "relay_connections_active"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "targets": [
          {
            "expr": "process_resident_memory_bytes / 1024 / 1024"
          }
        ]
      }
    ]
  }
}
```

## Load Testing

### Test Setup

```bash
# Install artillery for load testing
npm install -g artillery

# Create load test configuration
cat > artillery-config.yml << EOF
config:
  target: 'ws://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 300  
      arrivalRate: 100
      name: "Load test"
    - duration: 60
      arrivalRate: 200
      name: "Spike test"

scenarios:
  - name: "Subscribe and publish"
    weight: 100
    engine: ws
    
before:
  - log: "Starting load test"

after:
  - log: "Load test complete"
EOF

# Run load test
artillery run artillery-config.yml
```

### Benchmark Results

Typical performance on recommended hardware:

| Metric | Standalone | Distributed (3 nodes) |
|--------|------------|----------------------|
| **Events/sec** | 1,000-5,000 | 10,000-50,000 |
| **Concurrent Connections** | 1,000-5,000 | 10,000-50,000 |
| **Query Latency (p95)** | < 50ms | < 25ms |
| **Memory Usage** | 500MB-2GB | 1GB-4GB per node |

## Scaling Strategies

### Horizontal Scaling

1. **Add Relay Nodes**

   ```bash
   # Deploy additional relay instances
   docker-compose up --scale relay=3
   ```

2. **Load Balancing**

   ```nginx
   # Nginx configuration
   upstream relay_backend {
       server relay1:8080;
       server relay2:8080;
       server relay3:8080;
   }
   
   server {
       location / {
           proxy_pass http://relay_backend;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection "upgrade";
       }
   }
   ```

3. **Database Scaling**

   ```sql
   -- Add new database nodes
   cockroach start --join=existing-node:26257 ...
   ```

### Vertical Scaling

1. **Increase Resource Limits**

   ```yaml
   # Docker Compose
   services:
     relay:
       deploy:
         resources:
           limits:
             cpus: '8'
             memory: 16G
   ```

2. **Optimize Configuration**

   ```yaml
   RELAY:
     EVENT_CACHE_SIZE: 100000  # Scale with available RAM
     SEND_BUFFER_SIZE: 65536   # Larger buffers for more throughput
   ```

## Troubleshooting Performance Issues

### High CPU Usage

1. **Identify bottlenecks**

   ```bash
   # Profile the application
   go tool pprof http://localhost:6060/debug/pprof/profile
   
   # Check system CPU usage
   top -H -p $(pgrep shugur-relay)
   ```

2. **Common causes and solutions**
   - High event validation: Implement caching
   - Excessive logging: Reduce log level
   - Database queries: Optimize indexes

### High Memory Usage

1. **Memory profiling**

   ```bash
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

2. **Optimization strategies**
   - Reduce cache sizes
   - Implement event TTL
   - Use connection pooling

### Slow Database Queries

1. **Query analysis**

   ```sql
   SHOW QUERIES;
   SHOW TRACE FOR SELECT * FROM events WHERE kind = 1 LIMIT 100;
   ```

2. **Index optimization**

   ```sql
   CREATE INDEX CONCURRENTLY ON events (kind, created_at DESC) 
   STORING (id, pubkey, content);
   ```

## Best Practices

1. **Monitoring**: Implement comprehensive monitoring from day one
2. **Capacity Planning**: Plan for 3x expected load
3. **Testing**: Regular load testing in staging environment
4. **Updates**: Keep software updated for performance improvements
5. **Backup**: Regular performance benchmarks to detect regressions
6. **Documentation**: Document any custom optimizations

> **🚀 Performance Tip**: Start with conservative settings and gradually increase based on actual usage patterns. Monitor resource utilization and adjust accordingly.

> **📊 Monitoring Tip**: Set up alerts for key metrics like connection count, event processing rate, and database response times to catch issues early.

Remember: Premature optimization is the root of all evil. Always measure before optimizing!

## Next Steps

- Review your current configuration using the [Configuration Guide](./CONFIGURATION.md)
- Monitor your relay's performance using the metrics and logging features
- Scale your deployment as needed using the [Installation Guide](./installation/INSTALLATION.md)

## Related Documentation

- **[Installation Guide](./installation/INSTALLATION.md)**: Choose your deployment method
- **[Architecture Overview](./ARCHITECTURE.md)**: Understand the system design
- **[Configuration Guide](./CONFIGURATION.md)**: Configure your relay settings
- **[Troubleshooting Guide](./TROUBLESHOOTING.md)**: Resolve performance issues
- **[API Reference](./API.md)**: WebSocket and HTTP endpoint documentation
