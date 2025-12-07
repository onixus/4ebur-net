# 4ebur-net

[![CI Pipeline](https://github.com/onixus/4ebur-net/workflows/CI%20Pipeline/badge.svg)](https://github.com/onixus/4ebur-net/actions/workflows/ci.yml)
[![CodeQL](https://github.com/onixus/4ebur-net/workflows/CodeQL%20Security%20Analysis/badge.svg)](https://github.com/onixus/4ebur-net/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/onixus/4ebur-net)](https://goreportcard.com/report/github.com/onixus/4ebur-net)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Pulls](https://img.shields.io/docker/pulls/onixus/4ebur-net)](https://hub.docker.com/r/onixus/4ebur-net)
[![Go Version](https://img.shields.io/github/go-mod/go-version/onixus/4ebur-net)](https://go.dev/)

🚀 **High-performance MITM forward proxy server** written in Go with TLS termination support.

> Built for high-load production environments with critical performance optimizations

## ✨ Features

- **MITM HTTPS Interception** - Dynamic certificate generation for transparent HTTPS traffic inspection
- **TLS Termination** - Decrypt, inspect, and re-encrypt traffic on-the-fly
- **High Performance Architecture**
  - Connection pooling and reuse (100K+ concurrent connections)
  - Zero-copy optimization with `io.CopyBuffer`
  - Buffer pooling (`sync.Pool`) to reduce GC pressure
  - Configurable connection limits per host
  - Optimized TCP parameters
- **Production Ready** - Comprehensive error handling, structured logging, resource management
- **Docker Support** - Multi-stage builds, optimized images (15MB), production-ready compose
- **CI/CD Pipeline** - Automated testing, linting, security scanning, and releases

## 🏛️ Architecture

Designed with performance-critical patterns for high-load scenarios:

```
┌────────────────┐      ┌────────────────┐
│   Client      │─────▶│  4ebur-net   │─────▶ Target Server
│              │      │  MITM Proxy  │
└────────────────┘      └────────────────┘
   HTTPS Request        TLS Decrypt/Inspect        Forward Request
                        Certificate Generation
                        Connection Pooling
```

**Key optimizations:**
- Lightweight goroutines for massive concurrency
- Optimized `http.Transport` with intelligent connection pooling
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
  -e MAX_IDLE_CONNS=2000 \
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
go build -o 4ebur-net cmd/proxy/main.go
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
| `MAX_IDLE_CONNS` | `1000` | Maximum idle connections in pool |
| `MAX_IDLE_CONNS_PER_HOST` | `100` | Maximum idle connections per host |
| `MAX_CONNS_PER_HOST` | `100` | Maximum total connections per host |

**Example:**

```bash
PROXY_PORT=3128 \
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
  -e MAX_IDLE_CONNS=5000 \
  -e MAX_IDLE_CONNS_PER_HOST=500 \
  -e MAX_CONNS_PER_HOST=500 \
  onixus/4ebur-net:latest
```

## 📊 Performance Tuning

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

## 🔧 CI/CD Pipeline

### Automated Workflows

The project includes comprehensive GitHub Actions workflows:

#### CI Pipeline (`ci.yml`)
Runs on every push and PR:
- **Code Quality**: golangci-lint, gofmt, go vet
- **Multi-platform Build**: Linux, macOS, Windows (Go 1.21, 1.22)
- **Testing**: Unit tests with race detection and coverage
- **Security Scanning**: Gosec security analysis
- **Docker Build**: Test builds for both scratch and alpine images
- **Benchmarks**: Performance regression detection on PRs

#### Release Pipeline (`release.yml`)
Triggered on version tags (`v*.*.*`):
- **GitHub Releases**: Automated release creation with changelog
- **Multi-platform Binaries**: Linux, macOS, Windows (amd64, arm64)
- **Docker Images**: Multi-arch builds (amd64, arm64) pushed to Docker Hub
- **Versioning**: Semantic versioning with latest tag

#### Security Analysis (`codeql.yml`)
- **CodeQL**: Weekly automated security scans
- **Vulnerability Detection**: SARIF reports uploaded to GitHub Security

#### Dependency Management
- **Dependabot**: Automated dependency updates for Go, Docker, and GitHub Actions
- **Dependency Review**: PR analysis for security vulnerabilities

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

## 📁 Project Structure

```
4ebur-net/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                   # CI pipeline
│   │   ├── release.yml              # Release automation
│   │   ├── codeql.yml               # Security analysis
│   │   └── dependency-review.yml    # Dependency checks
│   ├── dependabot.yml           # Automated updates
│   └── PULL_REQUEST_TEMPLATE.md # PR template
├── cmd/
│   └── proxy/
│       └── main.go              # Application entry point
├── internal/
│   ├── proxy/
│   │   └── server.go            # Core proxy server logic
│   └── cert/
│       └── manager.go           # Dynamic certificate generation
├── pkg/
│   └── pool/
│       └── buffer.go            # Buffer pool for GC optimization
├── docker/
│   └── README.md            # Docker deployment guide
├── Dockerfile                   # Multi-stage minimal build
├── Dockerfile.alpine            # Alpine-based alternative
├── docker-compose.yml           # Production-ready compose
├── .golangci.yml                # Linter configuration
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

- **API Development**: Debug and inspect REST/GraphQL API calls
- **Security Testing**: Analyze application security, find vulnerabilities
- **Network Analysis**: Understand application network behavior
- **QA/Testing**: Validate HTTPS implementation, certificate handling
- **Protocol Analysis**: Study TLS handshakes, HTTP/2 multiplexing
- **Performance Testing**: Measure latency impact of SSL/TLS

## 📊 Benchmarks

Tested on: Intel Core i7 (4 cores), 16GB RAM, Ubuntu 22.04

| Metric | Value |
|--------|-------|
| Concurrent connections | 50,000+ |
| Throughput | 1+ Gbps |
| Latency overhead | < 5ms |
| Memory usage | ~200MB (50K connections) |
| CPU usage | ~30% (4 cores) |
| Docker image size | 15MB (scratch) / 25MB (alpine) |

## 🔧 Technical Details

### Certificate Generation

- **Algorithm**: RSA 2048-bit (balance of security/performance)
- **Validity**: 1 year from generation
- **Caching**: In-memory with `sync.Map` for thread-safe access
- **CA Lifetime**: 10 years

### Performance Optimizations

1. **Connection Pooling**: Reuses TCP connections to minimize overhead
2. **Zero-Copy I/O**: Uses `io.CopyBuffer` with pooled buffers
3. **Goroutine per Connection**: Go's lightweight concurrency model
4. **No Compression**: Disabled for maximum throughput (enable if needed)
5. **HTTP/2**: Automatic protocol upgrade when supported
6. **Buffer Pooling**: `sync.Pool` reduces GC pressure

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
- `✗` Error condition
- `🚀` Server startup
- `🔧` Configuration info

## 🔨 Troubleshooting

### "Certificate not trusted" errors

**Solution**: Install the CA certificate in your system trust store.

### High memory usage

**Solutions:**
- Reduce `MAX_IDLE_CONNS` and `MAX_IDLE_CONNS_PER_HOST`
- Increase `GOGC` value (e.g., 200 or 300)
- Monitor with `pprof` heap profiling
- For Docker: Set memory limits in compose file

### Connection refused errors

**Check:**
- Proxy is running: `netstat -tulpn | grep 1488` or `docker ps`
- Firewall rules allow port 1488
- Correct proxy configuration in client
- Docker port mapping is correct

### Slow performance

**Solutions:**
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

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

**Development guidelines:**
- Follow Go best practices and idioms
- Add tests for new features
- Update documentation
- Keep performance in mind (this is a high-load proxy)
- Test Docker builds before submitting
- All PRs must pass CI checks

**Code Quality:**
- Run `golangci-lint run` before committing
- Ensure `go fmt` and `go vet` pass
- Add unit tests with race detection
- Maintain or improve code coverage

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

- [x] Docker container support
- [x] Multi-stage optimized builds
- [x] Production-ready docker-compose
- [x] CI/CD pipeline with GitHub Actions
- [x] Automated security scanning
- [x] Multi-platform releases
- [ ] WebSocket proxying
- [ ] Request/response modification hooks
- [ ] Plugin system for custom logic
- [ ] Web UI for traffic inspection
- [ ] Traffic recording and replay
- [ ] Metrics export (Prometheus)
- [ ] gRPC support
- [ ] Load balancing capabilities
- [ ] Kubernetes deployment manifests

---

**⭐ If you find this project useful, please consider giving it a star!**
