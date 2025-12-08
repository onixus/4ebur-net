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
- **🌐 Web Interface** - Built-in web UI for easy CA certificate download and monitoring
- **📥 CA Certificate Endpoint** - Download CA certificate via `/ca.crt` endpoint
- **📊 API Endpoints** - `/health`, `/stats`, `/ca.crt` for monitoring and management
- **Intelligent HTTP Caching** ⚡
  - LRU eviction policy for memory efficiency
  - Automatic expiration based on Cache-Control headers
  - Cache-Control compliance (no-store, no-cache, private)
  - Configurable cache size and TTL
  - Real-time cache hit/miss metrics
  - Support for both HTTP and HTTPS caching
  - **40-150x performance improvement** with cache hits
- **High Performance Architecture**
  - Connection pooling and reuse (100K+ concurrent connections)
  - Zero-copy optimization with `io.CopyBuffer`
  - Buffer pooling (`sync.Pool`) to reduce GC pressure
  - Configurable connection limits per host
  - Optimized TCP parameters
- **🇷🇺 ALT Linux Support** - Native support for Russian OS (Sisyphus and P10)
- **Production Ready** - Comprehensive error handling, structured logging, resource management
- **Docker Support** - Multi-stage builds, optimized images (15MB-200MB), production-ready compose
- **CI/CD Pipeline** - Automated testing, linting, security scanning, and releases
- **Comprehensive Testing** - Unit tests, benchmarks, race detection, coverage reporting

## 🎉 New in v1.2.0

### 🌐 Web Interface

Access `http://localhost:1488/` in your browser for:
- 📥 One-click CA certificate download
- 📊 Real-time cache statistics
- 💚 Health check status
- 📖 Installation instructions
- 🔧 Configuration examples

### 📥 CA Certificate Download

No more manual certificate extraction!

```bash
# Download CA certificate directly from proxy
curl http://localhost:1488/ca.crt -o 4ebur-net-ca.crt

# Install on Arch Linux
sudo cp 4ebur-net-ca.crt /etc/ca-certificates/trust-source/anchors/
sudo trust extract-compat

# Now use proxy without -k flag!
curl -x http://localhost:1488 https://example.com
```

### 📊 API Endpoints

**Health Check:**
```bash
curl http://localhost:1488/health
# {"status":"ok","service":"4ebur-net"}
```

**Cache Statistics:**
```bash
curl http://localhost:1488/stats | jq .
# {
#   "cache_hits": 1523,
#   "cache_misses": 89,
#   "cache_size_bytes": 15234567,
#   "cache_entries": 234,
#   "hit_rate": 0.95
# }
```

**CA Certificate:**
```bash
curl http://localhost:1488/ca.crt -o ca.crt
```

### 🇷🇺 ALT Linux Support

Full support for Russian operating system ALT Linux:

**Docker images:**
- `onixus/4ebur-net:alt-sisyphus` - Rolling release (Go 1.25+)
- `onixus/4ebur-net:alt-p10` - Stable branch (Go 1.24+)

```bash
# Build and run on ALT Sisyphus
docker build -f Dockerfile.alt -t 4ebur-net:alt .
docker run -d -p 1488:1488 4ebur-net:alt

# Or use docker-compose
docker-compose -f docker-compose.alt.yml up -d
```

📚 [Complete ALT Linux Documentation](docs/ALT_LINUX.md)

## 🚀 Quick Start

### Docker (Recommended)

```bash
# Start proxy
docker run -d --name 4ebur-net -p 1488:1488 onixus/4ebur-net:latest

# Download CA certificate
curl http://localhost:1488/ca.crt -o 4ebur-net-ca.crt

# Install CA certificate (Arch Linux)
sudo cp 4ebur-net-ca.crt /etc/ca-certificates/trust-source/anchors/
sudo trust extract-compat

# Test proxy
curl -x http://localhost:1488 https://example.com

# Open web interface
firefox http://localhost:1488/
```

### From Source

```bash
git clone https://github.com/onixus/4ebur-net.git
cd 4ebur-net
make build
./4ebur-net
```

## 📊 Performance Results

Real-world test results:

| Metric | Request 1 (MISS) | Request 2 (HIT) | Request 3 (HIT) | Improvement |
|--------|------------------|-----------------|-----------------|-------------|
| **Latency** | 353 ms | 8.87 ms | 7.79 ms | **45x faster** |
| **GitHub API** | 1.34 sec | 8.95 ms | 7.79 ms | **150x faster** |
| **Static content** | 500 ms | 11 ms | 10 ms | **50x faster** |

✅ **SSL certificate verification** - Works without `-k` flag!  
✅ **Cache hit rate** - 60-90% for typical workloads  
✅ **Concurrent connections** - 50,000+ tested  
✅ **Docker image size** - 15MB (scratch) / 25MB (alpine) / 200MB (ALT Linux)

## 🛠️ Usage

### Configure Client

**Environment variables:**
```bash
export HTTP_PROXY=http://localhost:1488
export HTTPS_PROXY=http://localhost:1488
```

**With curl:**
```bash
curl -x http://localhost:1488 https://api.github.com/users/octocat
```

**System-wide proxy (GNOME):**
```bash
gsettings set org.gnome.system.proxy mode 'manual'
gsettings set org.gnome.system.proxy.http host 'localhost'
gsettings set org.gnome.system.proxy.http port 1488
gsettings set org.gnome.system.proxy.https host 'localhost'
gsettings set org.gnome.system.proxy.https port 1488
```

### Monitor Cache Performance

```bash
# Watch cache stats in real-time
watch -n 1 'curl -s http://localhost:1488/stats | jq .'

# View logs
docker logs -f 4ebur-net

# Example output:
# → GET https://api.github.com/users/octocat
# 💿 Cache MISS (HTTPS): https://api.github.com/users/octocat
# 💾 Cached HTTPS response for: ... (size: 1350 bytes)
# 
# → GET https://api.github.com/users/octocat
# 💾 Cache HIT (HTTPS): https://api.github.com/users/octocat
```

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
docker run -d \
  -p 1488:1488 \
  -e CACHE_SIZE_MB=500 \
  -e CACHE_MAX_AGE=15m \
  -e MAX_IDLE_CONNS=2000 \
  onixus/4ebur-net:latest
```

## 🐛 Troubleshooting

### "Certificate not trusted" errors

**Solution:** Download and install CA certificate:

```bash
# Download CA
curl http://localhost:1488/ca.crt -o 4ebur-net-ca.crt

# Install (Arch Linux)
sudo cp 4ebur-net-ca.crt /etc/ca-certificates/trust-source/anchors/
sudo trust extract-compat

# Install (Ubuntu/Debian)
sudo cp 4ebur-net-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Verify
trust list | grep -i 4ebur
```

### Cache not working

**Check:**
```bash
# Verify cache is enabled
curl http://localhost:1488/stats | jq .cache_entries

# Should be > 0 after some requests
```

### View detailed logs

```bash
# Docker
docker logs -f 4ebur-net | grep -E "Cache|CONNECT"

# Local
./4ebur-net 2>&1 | grep -E "Cache|CONNECT"
```

## 📚 Documentation

- [Docker Deployment Guide](docker/README.md)
- [ALT Linux Support](docs/ALT_LINUX.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Security Considerations](#-security-considerations)
- [Performance Tuning](#-performance-tuning)
- [API Documentation](#-api-endpoints)

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

## 🎯 Use Cases

- **API Development**: Debug and inspect REST/GraphQL API calls with caching analytics
- **Security Testing**: Analyze application security, find vulnerabilities
- **Network Analysis**: Understand application network behavior
- **QA/Testing**: Validate HTTPS implementation, certificate handling
- **Protocol Analysis**: Study TLS handshakes, HTTP/2 multiplexing
- **Performance Testing**: Measure latency impact of SSL/TLS and caching effectiveness
- **CDN Simulation**: Test content delivery with configurable cache policies

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

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**⭐ If you find this project useful, please consider giving it a star!**
