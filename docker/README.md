# Docker Deployment Guide

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start the proxy
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the proxy
docker-compose down
```

### Using Docker CLI

```bash
# Build the image
docker build -t onixus/4ebur-net:latest .

# Run the container
docker run -d \
  --name 4ebur-net \
  -p 8080:8080 \
  -e MAX_IDLE_CONNS=2000 \
  --restart unless-stopped \
  onixus/4ebur-net:latest

# View logs
docker logs -f 4ebur-net

# Stop and remove
docker stop 4ebur-net && docker rm 4ebur-net
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PROXY_PORT` | `8080` | Proxy listening port |
| `MAX_IDLE_CONNS` | `1000` | Maximum idle connections |
| `MAX_IDLE_CONNS_PER_HOST` | `100` | Max idle connections per host |
| `MAX_CONNS_PER_HOST` | `100` | Max total connections per host |
| `TZ` | `UTC` | Timezone |

### Resource Limits

Edit `docker-compose.yml` to adjust:

```yaml
deploy:
  resources:
    limits:
      cpus: '4.0'      # Maximum CPU cores
      memory: 2G       # Maximum memory
    reservations:
      cpus: '1.0'      # Reserved CPU cores
      memory: 512M     # Reserved memory
```

## Advanced Usage

### Custom Network Configuration

```bash
# Create custom network
docker network create --driver bridge proxy-net

# Run with custom network
docker run -d \
  --name 4ebur-net \
  --network proxy-net \
  -p 8080:8080 \
  onixus/4ebur-net:latest
```

### Persist CA Certificates

```bash
# Create volume for certificates
docker volume create 4ebur-certs

# Run with volume
docker run -d \
  --name 4ebur-net \
  -p 8080:8080 \
  -v 4ebur-certs:/certs \
  onixus/4ebur-net:latest
```

### Production Deployment

```bash
# Use Alpine-based image for more features
docker build -f Dockerfile.alpine -t onixus/4ebur-net:alpine .

# Run with production settings
docker run -d \
  --name 4ebur-net-prod \
  -p 8080:8080 \
  --restart always \
  --memory="2g" \
  --cpus="2.0" \
  --ulimit nofile=100000:100000 \
  -e MAX_IDLE_CONNS=5000 \
  -e MAX_IDLE_CONNS_PER_HOST=500 \
  -e MAX_CONNS_PER_HOST=500 \
  --log-driver json-file \
  --log-opt max-size=50m \
  --log-opt max-file=5 \
  onixus/4ebur-net:latest
```

## Docker Hub

### Pull from Docker Hub

```bash
docker pull onixus/4ebur-net:latest
```

### Build and Push

```bash
# Build multi-platform images
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t onixus/4ebur-net:latest \
  -t onixus/4ebur-net:1.0.0 \
  --push .
```

## Monitoring

### Health Checks

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' 4ebur-net

# View health check logs
docker inspect --format='{{json .State.Health}}' 4ebur-net | jq
```

### Resource Usage

```bash
# Real-time stats
docker stats 4ebur-net

# Container top processes
docker top 4ebur-net
```

### Logs

```bash
# Follow logs
docker logs -f 4ebur-net

# Last 100 lines
docker logs --tail 100 4ebur-net

# Logs since timestamp
docker logs --since 2025-12-07T12:00:00 4ebur-net
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker logs 4ebur-net

# Inspect container
docker inspect 4ebur-net

# Check port availability
netstat -tulpn | grep 8080
```

### High memory usage

```bash
# Check memory stats
docker stats --no-stream 4ebur-net

# Adjust memory limit
docker update --memory 2g --memory-swap 2g 4ebur-net
```

### Permission issues

```bash
# Run as specific user
docker run -d \
  --name 4ebur-net \
  --user 1000:1000 \
  -p 8080:8080 \
  onixus/4ebur-net:latest
```

## Security

### Run as non-root

The container runs as user `65534` (nobody) by default for security.

### Read-only filesystem

```bash
docker run -d \
  --name 4ebur-net \
  --read-only \
  --tmpfs /tmp \
  -p 8080:8080 \
  onixus/4ebur-net:latest
```

### Drop capabilities

```bash
docker run -d \
  --name 4ebur-net \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \
  -p 8080:8080 \
  onixus/4ebur-net:latest
```

## Image Sizes

| Image | Size | Description |
|-------|------|-------------|
| `onixus/4ebur-net:latest` | ~15MB | Minimal scratch-based image |
| `onixus/4ebur-net:alpine` | ~25MB | Alpine-based with shell tools |

## Performance Tuning

### System limits for host

```bash
# Increase file descriptors
echo 'fs.file-max = 500000' >> /etc/sysctl.conf

# TCP tuning
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 8192' >> /etc/sysctl.conf

# Apply changes
sysctl -p
```

### Docker daemon settings

Edit `/etc/docker/daemon.json`:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 100000,
      "Soft": 100000
    }
  }
}
```
