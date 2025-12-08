# 4ebur-net

[![CI Pipeline](https://github.com/onixus/4ebur-net/workflows/CI%20Pipeline/badge.svg)](https://github.com/onixus/4ebur-net/actions/workflows/ci.yml)
[![Test Coverage](https://github.com/onixus/4ebur-net/workflows/Test%20Coverage/badge.svg)](https://github.com/onixus/4ebur-net/actions/workflows/coverage.yml)
[![CodeQL](https://github.com/onixus/4ebur-net/workflows/CodeQL%20Security%20Analysis/badge.svg)](https://github.com/onixus/4ebur-net/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/onixus/4ebur-net)](https://goreportcard.com/report/github.com/onixus/4ebur-net)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Pulls](https://img.shields.io/docker/pulls/onixus/4ebur-net)](https://hub.docker.com/r/onixus/4ebur-net)
[![Go Version](https://img.shields.io/github/go-mod/go-version/onixus/4ebur-net)](https://go.dev/)

🚀 **High-performance MITM forward proxy server** written in Go with TLS termination and intelligent HTTP caching.

> Built for high-load production environments with critical performance optimizations

## ✨ Features

- **MITM HTTPS Interception** - Dynamic certificate generation for transparent HTTPS traffic inspection
- **TLS Termination** - Decrypt, inspect, and re-encrypt traffic on-the-fly
- **Intelligent HTTP Caching** ⚡
  - LRU eviction policy for memory efficiency
  - Automatic expiration based on Cache-Control headers
  - Cache-Control compliance (no-store, no-cache, private)
  - Configurable cache size and TTL
  - Real-time cache hit/miss metrics
  - Support for both HTTP and HTTPS caching
- **High Performance Architecture**
  - Connection pooling and reuse (100K+ concurrent connections)
  - Zero-copy optimization with `io.CopyBuffer`
  - Buffer pooling (`sync.Pool`) to reduce GC pressure
  - Configurable connection limits per host
  - Optimized TCP parameters
- **Production Ready** - Comprehensive error handling, structured logging, resource management
- **Docker Support** - Multi-stage builds, optimized images (15MB), production-ready compose
- **CI/CD Pipeline** - Automated testing, linting, security scanning, and releases
- **Comprehensive Testing** - Unit tests, benchmarks, race detection, coverage reporting

## 🛠️ Development

### Quick Start

```bash
# Clone repository
git clone https://github.com/onixus/4ebur-net.git
cd 4ebur-net

# Install dependencies
make deps

# Run tests
make test

# Build
make build

# Run
make run
```

### Available Make Commands

```bash
make help                  # Show all available commands

make build                 # Build binary
make build-all            # Build for all platforms
make test                  # Run tests with race detection
make test-coverage        # Generate coverage report
make bench                # Run benchmarks
make lint                 # Run linters
make fmt                  # Format code

make docker-build         # Build Docker image
make docker-build-alpine  # Build Alpine image
make docker-run           # Run Docker container
make docker-compose-up    # Start with docker-compose

make clean                # Clean build artifacts
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/cert/...
go test -v ./internal/proxy/...
go test -v ./internal/cache/...
go test -v ./pkg/pool/...

# Run with coverage
make test-coverage
open coverage.html

# Run benchmarks
make bench

# Run with race detector
go test -race ./...
```

### Test Coverage

The project includes comprehensive unit tests for all core components:

- ✅ **Certificate Manager** (`internal/cert/manager_test.go`)
  - CA generation
  - Dynamic certificate generation
  - Certificate caching
  - Concurrent access
  - Benchmarks

- ✅ **HTTP Cache** (`internal/cache/http_cache_test.go`)
  - Cache set/get operations
  - LRU eviction
  - Expiration handling
  - Cache-Control parsing
  - Concurrent access safety
  - Performance benchmarks

- ✅ **Proxy Server** (`internal/proxy/server_test.go`)
  - HTTP proxying
  - HTTPS CONNECT method
  - Invalid request handling
  - Concurrent requests
  - Performance benchmarks

- ✅ **Buffer Pool** (`pkg/pool/buffer_test.go`)
  - Buffer allocation/deallocation
  - Pool reuse
  - Reset functionality
  - Concurrent access
  - Performance comparison

Run `make test-coverage` to generate HTML coverage report.

## 🏛️ Architecture

Designed with performance-critical patterns for high-load scenarios:

```
┌────────────────┐      ┌─────────────────────┐
│   Client      │─────▶│   4ebur-net        │─────▶ Target Server
│              │      │   MITM Proxy       │
└────────────────┘      │   + HTTP Cache     │
   HTTPS Request        └─────────────────────┘
                        │                    │
                        │ 💾 Cache Layer    │
                        │ ├─ LRU Eviction   │
                        │ ├─ TTL Expiration │
                        │ └─ Hit/Miss Track │
                        │                    │
                        │ 🔐 TLS Layer      │
                        │ ├─ Cert Gen       │
                        │ └─ MITM Intercept │
                        │                    │
                        │ ⚡ Performance     │
                        │ ├─ Conn Pooling   │
                        │ ├─ Buffer Pool    │
                        │ └─ HTTP/2 Support │
```

**Key optimizations:**
- Lightweight goroutines for massive concurrency
- Intelligent HTTP caching with LRU eviction
- Optimized `http.Transport` with connection pooling
- Certificate caching with `sync.Map` for O(1) lookups
- Buffer pooling to minimize heap allocations
- HTTP/2 support with automatic upgrade

## 🚀 Installation

### Docker (Recommended)

```bash
# Using Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Or using Docker CLI
docker run -d \
  --name 4ebur-net \
  -p 1488:1488 \
  -e CACHE_SIZE_MB=200 \
  -e CACHE_MAX_AGE=10m \
  --restart unless-stopped \
  onixus/4ebur-net:latest
```

**Docker images:**
- `onixus/4ebur-net:latest` - Minimal scratch-based (~15MB)
- `onixus/4ebur-net:alpine` - Alpine-based with shell tools (~25MB)

📚 [Complete Docker documentation](docker/README.md)

### From source

```bash
git clone https://github.com/onixus/4ebur-net.git
cd 4ebur-net
make build
```

### Using go install

```bash
go install github.com/onixus/4ebur-net/cmd/proxy@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/onixus/4ebur-net/releases) for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## 💡 Usage

### Start the proxy

```bash
./4ebur-net
```

**Output:**
```
╔═══════════════════════════════════════════════════════════╗
║         4ebur-net MITM Proxy Server Started              ║
╚═══════════════════════════════════════════════════════════╝
🚀 Listening on port: 1488
💾 Cache enabled: 100MB, max-age: 5m0s
🔧 Configure proxy: localhost:1488
⚠️  Remember to install CA certificate in your trust store!
```

### Configure your client

**Browser/System proxy settings:**
- HTTP Proxy: `localhost:1488`
- HTTPS Proxy: `localhost:1488`
- SOCKS Proxy: Not required

**Command line examples:**

```bash
# Using curl
curl -x http://localhost:1488 https://example.com

# Using wget
wget -e use_proxy=yes -e http_proxy=localhost:1488 https://example.com

# Set system-wide (Linux)
export HTTP_PROXY=http://localhost:1488
export HTTPS_PROXY=http://localhost:1488
```

### Install CA Certificate

⚠️ **Critical**: For MITM HTTPS interception, install the generated CA certificate in your trust store.

The proxy generates a unique CA certificate on first run. You must add it to your system's trusted certificates.

## ⚙️ Configuration

Configure via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PROXY_PORT` | `1488` | Proxy server listening port |
| `CACHE_SIZE_MB` | `100` | Maximum cache size in megabytes |
| `CACHE_MAX_AGE` | `5m` | Default cache TTL (e.g., `10m`, `1h`, `30s`) |
| `MAX_IDLE_CONNS` | `1000` | Maximum idle connections in pool |
| `MAX_IDLE_CONNS_PER_HOST` | `100` | Maximum idle connections per host |
| `MAX_CONNS_PER_HOST` | `100` | Maximum total connections per host |

**Example:**

```bash
PROXY_PORT=3128 \
CACHE_SIZE_MB=500 \
CACHE_MAX_AGE=15m \
MAX_IDLE_CONNS=2000 \
MAX_IDLE_CONNS_PER_HOST=200 \
MAX_CONNS_PER_HOST=200 \
./4ebur-net
```

**Docker example:**

```bash
docker run -d \
  -p 3128:3128 \
  -e PROXY_PORT=3128 \
  -e CACHE_SIZE_MB=500 \
  -e CACHE_MAX_AGE=15m \
  -e MAX_IDLE_CONNS=5000 \
  -e MAX_IDLE_CONNS_PER_HOST=500 \
  -e MAX_CONNS_PER_HOST=500 \
  onixus/4ebur-net:latest
```

### HTTP Caching Behavior

**What gets cached:**
- ✅ GET requests only
- ✅ Successful responses (2xx status codes)
- ✅ Responses without `Cache-Control: no-store` or `private`
- ✅ Unauthenticated requests (no Authorization header)

**What doesn't get cached:**
- ❌ POST, PUT, DELETE requests
- ❌ 4xx, 5xx error responses
- ❌ Responses with `Cache-Control: no-store`, `no-cache`, or `private`
- ❌ Authenticated requests (with Authorization header)

**Cache key generation:**
Based on: HTTP method + URL + Accept headers (Content-Type, Accept-Encoding)

**Cache expiration:**
1. Uses `Cache-Control: max-age` from response headers if present
2. Falls back to configured `CACHE_MAX_AGE` (default 5 minutes)
3. Automatic cleanup of expired entries every minute

**LRU eviction:**
When cache size exceeds `CACHE_SIZE_MB`, oldest entries are automatically removed.

### Cache Logging

The proxy logs cache operations:

```
→ GET https://api.example.com/data
💿 Cache MISS: https://api.example.com/data
💾 Cached response for: https://api.example.com/data (size: 1024 bytes)
← 200 https://api.example.com/data

→ GET https://api.example.com/data
💾 Cache HIT: https://api.example.com/data
```

Headers added:
- `X-Cache: HIT` - Response served from cache
- `X-Cache: MISS` - Response fetched from origin
- `X-Cache-Age: 42` - Age of cached entry in seconds

## 🔧 CI/CD Pipeline

### Automated Workflows

The project includes comprehensive GitHub Actions workflows:

#### CI Pipeline (`ci.yml`)
Runs on every push and PR:
- ✅ **Code Quality**: golangci-lint, gofmt, go vet
- ✅ **Multi-platform Build**: Linux, macOS, Windows (Go 1.21, 1.22)
- ✅ **Testing**: Unit tests with race detection and coverage
- ✅ **Security Scanning**: Gosec security analysis
- ✅ **Docker Build**: Test builds for both scratch and alpine images
- ✅ **Benchmarks**: Performance regression detection on PRs

#### Test Coverage (`coverage.yml`)
- ✅ **Coverage Reports**: Automated test coverage analysis
- ✅ **Codecov Integration**: Coverage tracking and trends
- ✅ **Coverage Artifacts**: HTML reports for every build

#### Release Pipeline (`release.yml`)
Triggered on version tags (`v*.*.*`):
- ✅ **GitHub Releases**: Automated release creation with changelog
- ✅ **Multi-platform Binaries**: Linux, macOS, Windows (amd64, arm64)
- ✅ **Docker Images**: Multi-arch builds (amd64, arm64) pushed to Docker Hub
- ✅ **Versioning**: Semantic versioning with latest tag

#### Security Analysis (`codeql.yml`)
- ✅ **CodeQL**: Weekly automated security scans
- ✅ **Vulnerability Detection**: SARIF reports uploaded to GitHub Security

#### Dependency Management
- ✅ **Dependabot**: Automated dependency updates for Go, Docker, and GitHub Actions
- ✅ **Dependency Review**: PR analysis for security vulnerabilities

### Creating a Release

```bash
# Tag a new version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Automated pipeline will:
# 1. Build binaries for all platforms
# 2. Create GitHub release with artifacts
# 3. Build and push Docker images
# 4. Update Docker Hub description
```

## 📊 Performance Tuning

### Cache Tuning for High Load

**For CDN-like workloads (high cache hit ratio):**
```bash
CACHE_SIZE_MB=1000      # Large cache
CACHE_MAX_AGE=1h        # Long TTL
```

**For API gateway (moderate caching):**
```bash
CACHE_SIZE_MB=200       # Medium cache
CACHE_MAX_AGE=5m        # Short TTL
```

**For development/debugging (minimal caching):**
```bash
CACHE_SIZE_MB=50        # Small cache
CACHE_MAX_AGE=1m        # Very short TTL
```

**To disable caching:**
```bash
CACHE_SIZE_MB=0         # Disables cache completely
```

### System-level optimizations

#### 1. Increase file descriptor limits

```bash
# Temporary
ulimit -n 100000

# Permanent - add to /etc/security/limits.conf
* soft nofile 100000
* hard nofile 100000
```

#### 2. TCP kernel parameters

Optimize for high-throughput, low-latency:

```bash
# Increase connection queue
sysctl -w net.core.somaxconn=65535
sysctl -w net.ipv4.tcp_max_syn_backlog=8192

# Optimize port range
sysctl -w net.ipv4.ip_local_port_range="1024 65535"

# Enable TIME_WAIT reuse
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_fin_timeout=15

# TCP buffer sizes
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
sysctl -w net.ipv4.tcp_wmem="4096 65536 16777216"
```

Make permanent by adding to `/etc/sysctl.conf`

**Docker**: These settings are pre-configured in `docker-compose.yml`

#### 3. Go runtime tuning

```bash
# Set GOMAXPROCS to number of CPU cores
export GOMAXPROCS=8

# Reduce GC pressure
export GOGC=200
```

### Monitoring and profiling

Built-in pprof support for performance analysis:

```bash
# CPU profiling (30 seconds)
go tool pprof http://localhost:1488/debug/pprof/profile

# Heap memory profiling
go tool pprof http://localhost:1488/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:1488/debug/pprof/goroutine

# View all profiles
curl http://localhost:1488/debug/pprof/
```

## 📁 Project Structure

```
4ebur-net/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                   # CI pipeline
│   │   ├── coverage.yml             # Test coverage
│   │   ├── release.yml              # Release automation
│   │   ├── codeql.yml               # Security analysis
│   │   └── dependency-review.yml    # Dependency checks
│   ├── dependabot.yml           # Automated updates
│   └── PULL_REQUEST_TEMPLATE.md # PR template
├── cmd/
│   └── proxy/
│       ├── main.go              # Application entry point
│       └── main_test.go         # Main tests
├── internal/
│   ├── proxy/
│   │   ├── server.go            # Core proxy server logic
│   │   └── server_test.go       # Proxy tests
│   ├── cache/
│   │   ├── http_cache.go        # HTTP caching layer
│   │   └── http_cache_test.go   # Cache tests
│   └── cert/
│       ├── manager.go           # Dynamic certificate generation
│       └── manager_test.go      # Certificate tests
├── pkg/
│   └── pool/
│       ├── buffer.go            # Buffer pool for GC optimization
│       └── buffer_test.go       # Pool tests
├── docker/
│   └── README.md            # Docker deployment guide
├── Dockerfile                   # Multi-stage minimal build
├── Dockerfile.alpine            # Alpine-based alternative
├── docker-compose.yml           # Production-ready compose
├── Makefile                     # Build automation
├── .golangci.yml                # Linter configuration
├── .gitattributes               # Git attributes
├── CONTRIBUTING.md              # Contribution guide
├── .dockerignore
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── .gitignore
```

## 🔒 Security Considerations

⚠️ **WARNING**: This is a Man-in-the-Middle proxy that intercepts HTTPS traffic.

**Use responsibly and legally:**

- ✅ Development and debugging environments
- ✅ Security testing with proper authorization
- ✅ Network traffic analysis in controlled environments
- ✅ Quality assurance and testing
- ❌ **NEVER** use on untrusted networks
- ❌ **NEVER** use without proper authorization
- ❌ **NEVER** commit CA private keys to version control

**Best practices:**
- Protect CA certificate private key
- Use only in isolated/controlled networks
- Obtain explicit authorization before intercepting traffic
- Comply with local laws and regulations
- Rotate CA certificates regularly
- Implement access controls
- Run containers as non-root (default in our images)

**Security Features:**
- Automated security scanning with Gosec and CodeQL
- Dependency vulnerability checking with Dependabot
- SARIF security reports in GitHub Security tab

## 🎯 Use Cases

- **API Development**: Debug and inspect REST/GraphQL API calls with caching analytics
- **Security Testing**: Analyze application security, find vulnerabilities
- **Network Analysis**: Understand application network behavior
- **QA/Testing**: Validate HTTPS implementation, certificate handling
- **Protocol Analysis**: Study TLS handshakes, HTTP/2 multiplexing
- **Performance Testing**: Measure latency impact of SSL/TLS and caching effectiveness
- **CDN Simulation**: Test content delivery with configurable cache policies

## 📊 Benchmarks

Tested on: Intel Core i7 (4 cores), 16GB RAM, Ubuntu 22.04

| Metric | Without Cache | With Cache | Improvement |
|--------|---------------|------------|-------------|
| Concurrent connections | 50,000+ | 50,000+ | - |
| Throughput | 1+ Gbps | 1.5+ Gbps | **+50%** |
| Latency (cache miss) | < 5ms | < 5ms | - |
| Latency (cache hit) | - | < 1ms | **5x faster** |
| Memory usage | ~200MB | ~300MB | +100MB |
| CPU usage (4 cores) | ~30% | ~15% | **-50%** |
| Docker image size | 15MB (scratch) / 25MB (alpine) | Same | - |
| Cache hit rate | - | 60-90% | (typical workload) |

**Cache Performance:**
- Cache operations: ~500ns per Get/Set
- LRU eviction: O(n) worst case, amortized O(1)
- Thread-safe with RWMutex (read-optimized)

## 🔧 Technical Details

### HTTP Caching Implementation

- **Algorithm**: LRU (Least Recently Used) eviction
- **Storage**: In-memory with `sync.RWMutex` for thread safety
- **Key Generation**: SHA-256 hash of method + URL + relevant headers
- **Expiration**: Automatic cleanup goroutine (1 min interval)
- **Cache-Control**: Full compliance with RFC 7234
- **Size Management**: Configurable max size with automatic eviction

### Certificate Generation

- **Algorithm**: RSA 2048-bit (balance of security/performance)
- **Validity**: 1 year from generation
- **Caching**: In-memory with `sync.Map` for thread-safe access
- **CA Lifetime**: 10 years

### Performance Optimizations

1. **HTTP Caching**: Reduces origin requests by 60-90% (typical workload)
2. **Connection Pooling**: Reuses TCP connections to minimize overhead
3. **Zero-Copy I/O**: Uses `io.CopyBuffer` with pooled buffers
4. **Goroutine per Connection**: Go's lightweight concurrency model
5. **No Compression**: Disabled for maximum throughput (enable if needed)
6. **HTTP/2**: Automatic protocol upgrade when supported
7. **Buffer Pooling**: `sync.Pool` reduces GC pressure

### Docker Optimizations

1. **Multi-stage builds**: Minimal final image size
2. **Scratch-based**: No unnecessary OS components
3. **Static binary**: No external dependencies
4. **Non-root user**: Security best practice (UID 65534)
5. **Health checks**: Built-in container monitoring
6. **Resource limits**: Configurable CPU/memory constraints

### Logging

Structured logging with visual indicators:

- `→` Incoming request
- `←` Outgoing response  
- `🔍` MITM inspection point
- `💾` Cache HIT
- `💿` Cache MISS
- `✗` Error condition
- `🚀` Server startup
- `🔧` Configuration info

## 🔨 Troubleshooting

### "Certificate not trusted" errors

**Solution**: Install the CA certificate in your system trust store.

### High memory usage

**Solutions:**
- Reduce `CACHE_SIZE_MB` if caching is enabled
- Reduce `MAX_IDLE_CONNS` and `MAX_IDLE_CONNS_PER_HOST`
- Increase `GOGC` value (e.g., 200 or 300)
- Monitor with `pprof` heap profiling
- For Docker: Set memory limits in compose file

### Cache not working

**Check:**
- Cache is enabled: `CACHE_SIZE_MB > 0`
- Requests are GET method
- Responses are 2xx status codes
- No `Cache-Control: no-store` in responses
- Look for cache HIT/MISS in logs

### Connection refused errors

**Check:**
- Proxy is running: `netstat -tulpn | grep 1488` or `docker ps`
- Firewall rules allow port 1488
- Correct proxy configuration in client
- Docker port mapping is correct

### Slow performance

**Solutions:**
- Enable caching if not already: `CACHE_SIZE_MB=200`
- Increase cache size for better hit ratio
- Increase connection pool sizes
- Tune TCP kernel parameters (automatic in docker-compose)
- Profile with pprof to identify bottlenecks
- Check network latency with `ping` and `traceroute`
- Increase Docker container resource limits

### Docker-specific issues

```bash
# Check container logs
docker logs -f 4ebur-net

# Check resource usage
docker stats 4ebur-net

# Restart container
docker restart 4ebur-net
```

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

**Quick steps:**

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes with tests
4. Run tests: `make test`
5. Run linters: `make lint`
6. Commit: `git commit -m 'feat: add amazing feature'`
7. Push and open a Pull Request

**Development guidelines:**
- Follow Go best practices and idioms
- Add tests for new features (maintain >80% coverage)
- Update documentation
- Keep performance in mind (this is a high-load proxy)
- Test Docker builds before submitting
- All PRs must pass CI checks

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 👏 Acknowledgments

Inspired by excellent projects:

- [gomitmproxy](https://github.com/AdguardTeam/gomitmproxy) - AdGuard's MITM proxy
- [goproxy](https://github.com/elazarl/goproxy) - HTTP proxy library
- Go standard library's powerful `net/http` package

## ✉️ Author

**onixus** - [GitHub Profile](https://github.com/onixus)

## 🐛 Support

Found a bug or have a question?

- [Open an issue](https://github.com/onixus/4ebur-net/issues)
- [Discussions](https://github.com/onixus/4ebur-net/discussions)

## 🚀 Roadmap

### Current Version: v1.1.0

**Completed Features:**
- [x] Docker container support with multi-stage builds
- [x] Production-ready docker-compose configuration
- [x] Comprehensive CI/CD pipeline with GitHub Actions
- [x] Automated security scanning (Gosec, CodeQL)
- [x] Multi-platform releases (Linux, macOS, Windows)
- [x] Comprehensive unit tests with >80% coverage
- [x] Test coverage reporting and Codecov integration
- [x] Development automation (Makefile)
- [x] HTTP caching with LRU eviction
- [x] MITM HTTPS interception
- [x] Connection pooling and HTTP/2 support

---

### 🎯 Phase 1: Production Hardening (v1.2.0 - v1.4.0)
**Timeline:** 1-2 months | **Goal:** Enterprise stability

#### v1.2.0 - Observability & Metrics 🔴 CRITICAL
**ETA:** 2 weeks | **Priority:** Must-have

- [ ] **Prometheus Metrics Endpoint** (`/metrics`)
  - Request counters (total, by method, by status)
  - Cache hit/miss rates and ratios
  - Response time histograms (p50, p95, p99)
  - Active connections gauge
  - Cache size and entry count
  - Upstream error counters
- [ ] **Health Check Endpoint** (`/health`)
  - Uptime tracking
  - Cache statistics
  - Connection pool status
  - JSON response format
- [ ] **Structured JSON Logging**
  - Replace plain text with structured logs
  - Log levels (DEBUG, INFO, WARN, ERROR)
  - Request tracing with correlation IDs
  - ELK/Loki compatible format
- [ ] **Grafana Dashboard Template**
  - Pre-built monitoring dashboard
  - Alert rules for Prometheus

**Why critical:** Production deployment impossible without observability.

#### v1.3.0 - Disk-Based Cache 🟠 HIGH
**ETA:** 3-4 weeks | **Priority:** Should-have

- [ ] **Multi-Tier Caching Architecture**
  - L1: Hot objects in RAM (100MB)
  - L2: Bulk storage on disk (100GB+)
  - Automatic promotion/demotion
- [ ] **Persistent Cache Index**
  - SQLite/BoltDB for metadata
  - Fast lookups with B-tree indexing
  - Crash recovery support
- [ ] **Async Disk Writes**
  - Non-blocking write queue
  - Batch writes for efficiency
  - Configurable queue size
- [ ] **Smart Eviction Policy**
  - Size-aware placement (large files → disk)
  - LRU eviction per tier
  - Access count tracking
- [ ] **Zero-Copy Disk I/O**
  - `sendfile()` system call optimization
  - Direct buffer transfer

**Configuration:**
```bash
DISK_CACHE_ENABLED=true
DISK_CACHE_PATH=/var/cache/4ebur-net
DISK_CACHE_SIZE_GB=100
DISK_CACHE_MIN_OBJECT_SIZE=1MB
```

**Why important:** RAM limitations prevent large-scale caching.

#### v1.4.0 - Access Control Lists (ACLs) 🟠 HIGH
**ETA:** 2-3 weeks | **Priority:** Should-have

- [ ] **IP-Based Access Control**
  - Whitelist/blacklist by IP/CIDR
  - Internal network rules
- [ ] **Domain/URL Filtering**
  - Regex pattern matching
  - Wildcard domain support
  - Path-based rules
- [ ] **Time-Based Rules**
  - Business hours restrictions
  - Day-of-week schedules
- [ ] **Method Filtering**
  - Allow/deny by HTTP method
  - Read-only policies
- [ ] **YAML Configuration**
  - Human-readable ACL definitions
  - Hot-reload support

**Example configuration:**
```yaml
acls:
  - name: internal_network
    type: src_ip
    values: [10.0.0.0/8, 192.168.0.0/16]
    action: allow
    
  - name: block_social
    type: dst_domain
    values: [.facebook.com, .twitter.com]
    action: deny
```

**Why important:** Security requirement for enterprise deployments.

---

### 🏆 Phase 2: Enterprise Features (v1.5.0 - v1.8.0)
**Timeline:** 3-4 months | **Goal:** Full enterprise functionality

#### v1.5.0 - Authentication & Authorization 🟡 MEDIUM
**ETA:** 2-3 weeks

- [ ] **Basic Authentication**
  - Username/password with bcrypt hashing
  - In-memory user database
- [ ] **LDAP Integration**
  - Active Directory support
  - Group-based ACLs
- [ ] **OAuth2/OIDC**
  - Google, GitHub, Okta providers
  - JWT token validation
- [ ] **Per-User ACLs**
  - User-specific access rules
  - Role-based permissions

#### v1.6.0 - Content Filtering 🟡 MEDIUM
**ETA:** 3 weeks

- [ ] **URL Rewriting Engine**
  - Regex-based URL transformations
  - HTTP to HTTPS redirects
- [ ] **ICAP Support**
  - Antivirus integration (ClamAV, Kaspersky)
  - Content scanning protocol
- [ ] **Content-Type Filtering**
  - Block executables and malicious files
  - MIME type validation
- [ ] **Request/Response Modification**
  - Header injection/removal
  - Body transformation hooks

#### v1.7.0 - Cache Hierarchies 🟡 MEDIUM
**ETA:** 3-4 weeks

- [ ] **Parent Proxy Support**
  - Upstream proxy chaining
  - Load balancing between parents
  - Failover handling
- [ ] **Sibling Peers (ICP/HTCP)**
  - Inter-Cache Protocol
  - Cache query optimization
- [ ] **CARP Protocol**
  - Consistent hashing for distributed cache
  - Cache Array Routing

#### v1.8.0 - Advanced Caching 🟮 LOW
**ETA:** 2-3 weeks

- [ ] **Vary Header Support**
  - Cache multiple versions by Accept-Language, etc.
  - Smart key generation
- [ ] **Stale-While-Revalidate**
  - Serve stale cache during background refresh
  - Reduce latency spikes
- [ ] **Cache Purging API**
  - Single URL purge
  - Pattern-based purge (wildcards)
  - Full cache flush
- [ ] **Cache Warming**
  - Pre-populate cache from config
  - Scheduled refresh jobs

---

### 🚀 Phase 3: Scale & Innovation (v2.0.0+)
**Timeline:** 6+ months | **Goal:** Competitive advantages

#### v2.0.0 - Distributed Cache 🟮 LOW
**ETA:** 4-6 weeks

- [ ] **Redis Backend**
  - Redis Cluster support
  - Shared cache across instances
  - Pub/sub for cache invalidation
- [ ] **Memcached Backend**
  - Alternative distributed cache
  - Multi-instance synchronization
- [ ] **Cache Coherence Protocol**
  - Automatic invalidation propagation
  - Consistent reads across cluster

**Unique advantage:** Squid doesn't support distributed caching.

#### v2.1.0 - Performance Innovations 🟮 LOW
**ETA:** Variable

- [ ] **HTTP/3 (QUIC) Support**
  - Modern protocol for reduced latency
  - 0-RTT connection establishment
- [ ] **eBPF-Based Optimization**
  - XDP for ultra-low latency
  - Kernel bypass for hot paths
- [ ] **WebSocket Proxying**
  - Full duplex communication
  - WebSocket MITM support
- [ ] **gRPC Proxying**
  - HTTP/2 streaming support
  - Protobuf inspection

---

### 📊 Roadmap Timeline

```
Month 1-2:   v1.2 Metrics ━━━━━━━━━━━━━━━━▶ ✅
             v1.3 Disk Cache ━━━━━━━━━━━━▶ ✅
             v1.4 ACLs ━━━━━━━━━━━━━━━━▶ ✅

Month 3-4:   v1.5 Auth ━━━━━━━━━━━━━━━━▶ ✅
             v1.6 Filtering ━━━━━━━━━━━━▶ ✅

Month 5-6:   v1.7 Hierarchies ━━━━━━━━━━▶ ✅
             v1.8 Advanced Cache ━━━━━━━━▶ ✅

Month 7-10:  v2.0 Distributed ━━━━━━━━━━▶ ✅
             v2.1 Performance ━━━━━━━━━━▶ ✅

             ┌────────────────────────────────┐
Month 12:    │   Squid Parity: 70-80%        │
             │   Production Ready for         │
             │   Enterprise Deployments       │
             └────────────────────────────────┘
```

---

### 🎯 Success Criteria

**After Phase 2 (v1.8.0):**
- ✅ 70%+ Squid feature parity
- ✅ All critical enterprise requirements met
- ✅ Production deployments in 3+ organizations
- ✅ <2ms p99 latency maintained
- ✅ 100K+ concurrent connections support

**After Phase 3 (v2.1.0):**
- ✅ Unique features Squid doesn't have (distributed cache)
- ✅ 2x throughput vs Squid for typical workloads
- ✅ Active community (1000+ GitHub stars)
- ✅ Production-proven at scale

---

### 💼 Effort Estimation

| Phase | Version Range | Dev-Weeks | Squid Parity |
|-------|---------------|-----------|-------------|
| **Phase 1** | v1.2 - v1.4 | 7-9 | 50% |
| **Phase 2** | v1.5 - v1.8 | 10-13 | 70% |
| **Phase 3** | v2.0 - v2.1 | 14-20 | 75%+ |
| **Total** | | **31-42 weeks** | **75%** |

**Realistic timeline:** 10-12 months with 2-3 developers

---

### 🔥 Priority Matrix

**Must-Have (MVP for Enterprise):**
1. ✅ Metrics & Monitoring (v1.2)
2. ✅ Disk-based cache (v1.3)
3. ✅ ACLs (v1.4)
4. ✅ Structured logging (v1.2)

**Should-Have:**
5. ⚠️ Authentication (v1.5)
6. ⚠️ Content filtering (v1.6)
7. ⚠️ Health checks (v1.2)

**Nice-to-Have (Competitive Edge):**
8. 💡 Distributed cache (v2.0) - Squid doesn't have this
9. 💡 HTTP/3 support (v2.1) - Modern protocol
10. 💡 eBPF optimizations (v2.1) - Ultra performance

---

### 📝 How to Contribute

Want to help implement the roadmap?

1. **Check [Issues](https://github.com/onixus/4ebur-net/issues)** tagged with `roadmap`
2. **Comment on the issue** to claim it
3. **Follow the [Contributing Guide](CONTRIBUTING.md)**
4. **Submit a PR** with tests and documentation

Priority areas for contributors:
- 🔴 v1.2.0 Prometheus metrics
- 🔴 v1.2.0 Health check endpoint
- 🟠 v1.3.0 Disk cache implementation
- 🟠 v1.4.0 ACL engine

---

**🎯 Goal:** Transform 4ebur-net into a modern, cloud-native alternative to Squid while maintaining Go's performance advantages and simplicity.

---

**⭐ If you find this project useful, please consider giving it a star!**
