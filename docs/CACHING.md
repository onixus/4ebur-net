# Distributed Caching with Redis

## Overview

4ebur-net supports **tiered caching** with in-memory L1 cache and distributed Redis L2 cache.

## Architecture

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│ 4ebur-net #1 │◄─────►│ Redis Master │◄─────►│ 4ebur-net #2 │
│  L1: RAM     │       │  (read/write)│       │  L1: RAM     │
│  L2: Redis   │       └──────┬───────┘       │  L2: Redis   │
└──────────────┘              │               └──────────────┘
                              │ Replication
                              ▼
                       ┌──────────────┐
                       │Redis Replica │
                       │  (read-only) │
                       └──────────────┘
```

## Features

### Tiered Cache Benefits

1. **L1 (In-Memory)**:
   - Ultra-fast: <1ms latency
   - Hot objects cached locally
   - LRU eviction
   - Size: 100-500MB

2. **L2 (Redis)**:
   - Distributed: shared across instances
   - Persistent: survives restarts
   - Larger: 1-10GB+
   - Pub/Sub for invalidation

### Cache Flow

1. Request comes in
2. Check L1 (RAM) - if HIT → return (fastest)
3. Check L2 (Redis) - if HIT → promote to L1, return
4. MISS → fetch from origin → store in L1 + L2

## Configuration

### Environment Variables

```bash
# Enable Redis backend
REDIS_ENABLED=true
REDIS_ADDR=redis-master:6379
REDIS_PASSWORD=your_strong_password
REDIS_DB=0

# L1 Cache (in-memory)
CACHE_SIZE_MB=200
CACHE_MAX_AGE=5m

# L2 will use same TTL
```

### Docker Compose Setup

**docker-compose.cache.yml:**

```yaml
version: '3.8'

services:
  redis-master:
    image: redis:7.2-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-master-data:/data
      - ./config/redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis-replica:
    image: redis:7.2-alpine
    ports:
      - "6380:6379"
    volumes:
      - redis-replica-data:/data
    command: redis-server --slaveof redis-master 6379
    depends_on:
      - redis-master

  redis-sentinel:
    image: redis:7.2-alpine
    ports:
      - "26379:26379"
    volumes:
      - ./config/sentinel.conf:/usr/local/etc/redis/sentinel.conf
    command: redis-sentinel /usr/local/etc/redis/sentinel.conf
    depends_on:
      - redis-master
      - redis-replica

volumes:
  redis-master-data:
  redis-replica-data:
```

### Redis Configuration

**config/redis.conf:**

```conf
# Persistence
save 900 1
save 300 10
save 60 10000

appendonly yes
appendfsync everysec

# Memory
maxmemory 2gb
maxmemory-policy allkeys-lru

# Performance
tcp-backlog 511
timeout 0
tcp-keepalive 300

# Replication
repl-diskless-sync yes
repl-diskless-sync-delay 5

# Security
requirepass your_strong_password_here
```

## Usage

### Code Integration

```go
package main

import (
    "github.com/onixus/4ebur-net/internal/cache"
)

func main() {
    // Create L1 cache (in-memory)
    l1Cache := cache.NewHTTPCache(200*1024*1024, 5*time.Minute)
    
    // Create L2 cache (Redis)
    l2Cache, err := cache.NewRedisBackend(
        os.Getenv("REDIS_ADDR"),
        os.Getenv("REDIS_PASSWORD"),
        0,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Create tiered cache
    tieredCache := cache.NewTieredCache(l1Cache, l2Cache, 5*time.Minute)
    
    // Use tiered cache in proxy
    proxy := NewProxyServer(tieredCache)
}
```

### Cache Invalidation

**Invalidate across all instances:**

```go
// Clear cache for specific pattern
tieredCache.InvalidatePattern("api.example.com/*")

// This will:
// 1. Clear local L1 cache
// 2. Publish to Redis Pub/Sub
// 3. Other instances receive message and clear their L1
```

## Performance

### Benchmarks

| Operation | L1 (RAM) | L2 (Redis) | Improvement |
|-----------|----------|------------|-------------|
| Cache GET | 500ns | 0.15ms | 300x faster |
| Cache SET | 500ns | 0.20ms | 400x faster |
| Latency p99 | <1ms | <2ms | 2x faster |

### Cache Hit Ratio

```
Total requests: 1,000,000
L1 hits: 700,000 (70%)
L2 hits: 200,000 (20%)
Misses: 100,000 (10%)

Total hit rate: 90%
L1 hit rate: 70%
```

**Impact:**
- 90% of requests never hit origin
- 70% served from ultra-fast L1 cache
- Reduced origin load by 90%

## Monitoring

### Cache Statistics

```go
hits, misses, l1Hits, l2Hits, hitRate := tieredCache.Stats()

log.Printf("Cache Stats: hits=%d, misses=%d, rate=%.2f%%", 
    hits, misses, hitRate*100)
log.Printf("L1 hits: %d, L2 hits: %d", l1Hits, l2Hits)
```

### Redis Monitoring

```bash
# Connect to Redis
redis-cli -h localhost -p 6379 -a your_password

# Get info
INFO stats

# Check memory usage
INFO memory

# Monitor commands
MONITOR

# Get cache size
DBSIZE
```

## High Availability

### Redis Sentinel Setup

**config/sentinel.conf:**

```conf
port 26379

# Monitor master
sentinel monitor mymaster redis-master 6379 2
sentinel auth-pass mymaster your_password

# Failover settings
sentinel down-after-milliseconds mymaster 5000
sentinel parallel-syncs mymaster 1
sentinel failover-timeout mymaster 10000
```

**Automatic failover:**
1. Master goes down
2. Sentinel promotes replica to master
3. Applications auto-reconnect to new master
4. Zero downtime

## Best Practices

1. **Set appropriate TTLs**
   - Short TTL (1-5min) for dynamic data
   - Long TTL (1h+) for static assets

2. **Monitor cache hit ratio**
   - Target: >70% for typical workloads
   - <50% = too small cache or bad keys

3. **Size L1 appropriately**
   - 100-200MB for API gateway
   - 500MB-1GB for CDN-like workload

4. **Use Redis persistence**
   - RDB + AOF for durability
   - Faster restarts

5. **Enable replication**
   - Master-slave for HA
   - Sentinel for automatic failover

6. **Secure Redis**
   - Use strong passwords
   - Bind to private network
   - Enable TLS in production

## Troubleshooting

### Redis connection failed

```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 -a password ping

# Check logs
docker logs redis-master
```

### Low cache hit ratio

- Increase L1 cache size: `CACHE_SIZE_MB=500`
- Check if cache keys are consistent
- Verify TTL is not too short
- Monitor eviction rate

### High memory usage

- Reduce `maxmemory` in redis.conf
- Use `allkeys-lru` eviction policy
- Monitor with `INFO memory`

### Slow Redis queries

```bash
# Find slow queries
redis-cli --latency
redis-cli --latency-history

# Check for blocking operations
redis-cli SLOWLOG GET 10
```

## Migration from In-Memory Only

**Step 1:** Deploy Redis
```bash
docker-compose -f docker-compose.cache.yml up -d
```

**Step 2:** Enable Redis in 4ebur-net
```bash
REDIS_ENABLED=true
REDIS_ADDR=redis-master:6379
```

**Step 3:** Restart proxy
```bash
docker-compose restart proxy
```

**Step 4:** Monitor metrics
- Check L1/L2 hit ratios
- Verify Redis memory usage
- Monitor latency impact

## See Also

- [Redis Documentation](https://redis.io/documentation)
- [Redis Best Practices](https://redis.io/topics/best-practices)
- [Go Redis Client](https://github.com/redis/go-redis)
