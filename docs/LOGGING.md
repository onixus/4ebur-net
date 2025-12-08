# Logging Infrastructure

## Overview

4ebur-net uses **Zerolog** for structured, high-performance logging with **Grafana Loki** for log aggregation and analysis.

## Stack Architecture

```
4ebur-net (Zerolog)  →  Promtail  →  Loki  →  Grafana
     JSON logs          Collector    Storage   Visualization
```

## Zerolog Features

### Why Zerolog?

- **Zero allocations**: No heap allocations = no GC pressure
- **Fast**: 0 allocs/op, 50ns per log operation
- **Structured**: JSON output for machine parsing
- **Levels**: DEBUG, INFO, WARN, ERROR
- **Context**: Automatic service/version fields

### Configuration

**Environment Variables:**

```bash
# Log level: debug, info, warn, error
LOG_LEVEL=info

# Log format: json (production) or pretty (development)
LOG_FORMAT=json

# Application version (optional)
VERSION=1.2.0
```

### Usage Examples

**Basic logging:**

```go
logger := logger.NewLogger()

logger.Info("Server started")
logger.Debug("Debug information")
logger.Warn("Warning message")
logger.Error(err, "Error occurred")
```

**HTTP Request logging:**

```go
logger.LogHTTPRequest(
    r.Method,
    r.URL.String(),
    r.RemoteAddr,
    statusCode,
    latency,
    cacheHit,
    r.UserAgent(),
)
```

**Output (JSON):**

```json
{
  "level": "info",
  "service": "4ebur-net",
  "version": "1.2.0",
  "time": "2025-12-08T19:45:23.123456789Z",
  "method": "GET",
  "url": "https://api.example.com/users",
  "remote_addr": "10.0.1.42:54321",
  "status": 200,
  "latency_ms": 42.5,
  "cache_hit": true,
  "user_agent": "Mozilla/5.0...",
  "message": "http_request"
}
```

## Grafana Loki Deployment

### Docker Compose Setup

**docker-compose.logging.yml:**

```yaml
version: '3.8'

services:
  # Promtail - log collector
  promtail:
    image: grafana/promtail:2.9.3
    volumes:
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock
      - ./config/promtail-config.yml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
    depends_on:
      - loki

  # Loki - log storage
  loki:
    image: grafana/loki:2.9.3
    ports:
      - "3100:3100"
    volumes:
      - loki-data:/loki
      - ./config/loki-config.yml:/etc/loki/local-config.yaml
    command: -config.file=/etc/loki/local-config.yaml

  # Grafana - visualization
  grafana:
    image: grafana/grafana:10.2.2
    ports:
      - "3000:3000"
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana-datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml
    depends_on:
      - loki

volumes:
  loki-data:
  grafana-data:
```

### Loki Configuration

**config/loki-config.yml:**

```yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
    kvstore:
      store: inmemory

# High-load optimizations
limits_config:
  ingestion_rate_mb: 100
  ingestion_burst_size_mb: 200
  per_stream_rate_limit: 10MB
  per_stream_rate_limit_burst: 20MB
  max_entries_limit_per_query: 10000
  max_query_length: 721h  # 30 days

# Retention
compactor:
  working_directory: /loki/compactor
  shared_store: filesystem
  retention_enabled: true
  retention_delete_delay: 2h

table_manager:
  retention_deletes_enabled: true
  retention_period: 720h  # 30 days
```

### Promtail Configuration

**config/promtail-config.yml:**

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push
    batch_wait: 1s
    batch_size: 1048576

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          - name: label
            values: ["logging=promtail"]
    
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        target_label: 'container'
      - source_labels: ['__meta_docker_container_log_stream']
        target_label: 'stream'
      - source_labels: ['__meta_docker_container_label_logging_jobname']
        target_label: 'job'
    
    # JSON parsing
    pipeline_stages:
      - json:
          expressions:
            level: level
            method: method
            status: status
            latency: latency_ms
            cache_hit: cache_hit
            url: url
      
      - labels:
          level:
          method:
          status:
      
      - timestamp:
          source: time
          format: RFC3339Nano
```

### Grafana Data Source

**config/grafana-datasources.yml:**

```yaml
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: true
    editable: false
```

## Running the Stack

```bash
# Start logging infrastructure
docker-compose -f docker-compose.logging.yml up -d

# View logs
docker-compose -f docker-compose.logging.yml logs -f

# Access Grafana
open http://localhost:3000
```

## LogQL Queries (Grafana)

### Basic queries:

```logql
# All logs from 4ebur-net
{job="4ebur-net"}

# Only errors
{job="4ebur-net"} |= "level=error"

# Cache hits
{job="4ebur-net"} | json | cache_hit="true"

# Slow requests (>1s)
{job="4ebur-net"} | json | latency_ms > 1000

# 4xx/5xx errors
{job="4ebur-net"} | json | status >= 400
```

### Metrics queries:

```logql
# Request rate
rate({job="4ebur-net"}[5m])

# Error rate
rate({job="4ebur-net"} |= "level=error"[5m])

# Average latency
avg_over_time({job="4ebur-net"} | json | unwrap latency_ms [5m])

# Cache hit ratio
sum(rate({job="4ebur-net"} | json | cache_hit="true" [5m])) 
/ 
sum(rate({job="4ebur-net"} | json [5m]))
```

## Performance

**Benchmarks (Zerolog):**

```
BenchmarkLogHTTPRequest-8    30000000    50 ns/op    0 B/op    0 allocs/op
```

**Loki Ingestion:**

- 5M+ logs/second
- <100ms query latency
- 10x less storage vs Elasticsearch

## Troubleshooting

### No logs in Grafana

1. Check Promtail is running: `docker logs promtail`
2. Verify Docker labels on proxy container
3. Check Loki connectivity: `curl http://localhost:3100/ready`

### High memory usage

- Reduce `retention_period` in loki-config.yml
- Lower `ingestion_rate_mb` limits
- Enable compression

### Slow queries

- Add more specific label filters
- Reduce time range
- Use `| json` parsing efficiently

## Production Best Practices

1. **Use JSON format** in production (`LOG_FORMAT=json`)
2. **Set appropriate log level** (`LOG_LEVEL=info`, not debug)
3. **Configure retention** (30 days default, adjust as needed)
4. **Monitor Loki metrics** (Prometheus integration)
5. **Use S3/GCS** for long-term storage (instead of filesystem)
6. **Enable compression** in Loki for storage efficiency
7. **Set up alerts** for error rates and slow queries

## See Also

- [Zerolog Documentation](https://github.com/rs/zerolog)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
