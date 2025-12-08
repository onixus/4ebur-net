package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/onixus/4ebur-net/internal/cache"
	"github.com/onixus/4ebur-net/internal/cert"
	"github.com/onixus/4ebur-net/pkg/pool"
)

// ProxyServer is the main MITM proxy server
type ProxyServer struct {
	certManager *cert.CertManager
	transport   *http.Transport
	httpCache   *cache.HTTPCache
	mu          sync.RWMutex
}

// NewProxyServer creates a new proxy server instance
func NewProxyServer() (*ProxyServer, error) {
	certMgr, err := cert.NewCertManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create cert manager: %w", err)
	}

	// Get configuration from environment
	maxIdleConns := getEnvInt("MAX_IDLE_CONNS", 1000)
	maxIdleConnsPerHost := getEnvInt("MAX_IDLE_CONNS_PER_HOST", 100)
	maxConnsPerHost := getEnvInt("MAX_CONNS_PER_HOST", 100)
	cacheSize := getEnvInt64("CACHE_SIZE_MB", 100) * 1024 * 1024
	cacheMaxAge := getEnvDuration("CACHE_MAX_AGE", 5*time.Minute)

	// Create optimized HTTP transport
	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		MaxConnsPerHost:     maxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // For MITM proxy
		},
		DisableCompression: true, // Maximum throughput
		ForceAttemptHTTP2:  true, // Enable HTTP/2
	}

	// Create HTTP cache
	httpCache := cache.NewHTTPCache(cacheSize, cacheMaxAge)

	// Log cache configuration
	log.Printf("💾 Cache enabled: %dMB, max-age: %v", cacheSize/(1024*1024), cacheMaxAge)

	return &ProxyServer{
		certManager: certMgr,
		transport:   transport,
		httpCache:   httpCache,
	}, nil
}

// ServeHTTP handles incoming HTTP requests
func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
	} else {
		p.handleHTTP(w, r)
	}
}

// handleHTTP handles regular HTTP requests with caching
func (p *ProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL)

	// Try to get from cache
	cacheKey := cache.GenerateKey(r)
	if entry, found := p.httpCache.Get(cacheKey); found {
		log.Printf("💾 Cache HIT: %s", r.URL)
		if err := entry.WriteToResponse(w); err != nil {
			log.Printf("✗ Error writing cached response: %v", err)
		}
		return
	}

	log.Printf("💿 Cache MISS: %s", r.URL)

	// Create new request to target
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Send request
	resp, err := p.transport.RoundTrip(req)
	if err != nil {
		log.Printf("✗ Error forwarding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Check if cacheable
	if cache.IsCacheable(r, resp) {
		// Create cache entry
		if entry, err := cache.CreateCacheEntry(resp, p.httpCache.Stats); err == nil {
			p.httpCache.Set(cacheKey, entry)
			log.Printf("💾 Cached response for: %s (size: %d bytes)", r.URL, entry.Size)
		}
	}

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Add cache miss header
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(resp.StatusCode)

	// Copy response body with pooled buffer
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)

	_, err = io.CopyBuffer(w, resp.Body, buf.Bytes()[:cap(buf.Bytes())])
	if err != nil && err != io.EOF {
		log.Printf("✗ Error copying response: %v", err)
	}

	log.Printf("← %d %s", resp.StatusCode, r.URL)
}

// handleConnect handles HTTPS CONNECT requests for MITM
func (p *ProxyServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 CONNECT %s", r.Host)

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("✗ Hijack error: %v", err)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Printf("✗ Failed to send 200: %v", err)
		return
	}

	// Extract hostname
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	// Get certificate for host
	tlsCert, err := p.certManager.GetCertificate(host)
	if err != nil {
		log.Printf("✗ Failed to get certificate for %s: %v", host, err)
		return
	}

	// Wrap connection with TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
	}

	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	// Perform TLS handshake
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("✗ TLS handshake failed: %v", err)
		return
	}

	// Read HTTP request from TLS connection
	reader := bufio.NewReader(tlsConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("✗ Failed to read request: %v", err)
		return
	}

	// Fix request URL
	req.URL.Scheme = "https"
	req.URL.Host = r.Host

	// Check cache for HTTPS requests
	cacheKey := cache.GenerateKey(req)
	if entry, found := p.httpCache.Get(cacheKey); found {
		log.Printf("💾 Cache HIT (HTTPS): %s", req.URL)
		// Write cached response
		var buf bytes.Buffer
		resp := &http.Response{
			StatusCode: entry.StatusCode,
			Header:     entry.Headers,
			Body:       io.NopCloser(bytes.NewReader(entry.Body)),
		}
		resp.Write(&buf)
		tlsConn.Write(buf.Bytes())
		return
	}

	log.Printf("💿 Cache MISS (HTTPS): %s", req.URL)

	// Forward request
	resp, err := p.transport.RoundTrip(req)
	if err != nil {
		log.Printf("✗ Error forwarding HTTPS request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check if cacheable and cache it
	if cache.IsCacheable(req, resp) {
		if entry, err := cache.CreateCacheEntry(resp, p.httpCache.Stats); err == nil {
			p.httpCache.Set(cacheKey, entry)
			log.Printf("💾 Cached HTTPS response for: %s (size: %d bytes)", req.URL, entry.Size)
		}
	}

	// Write response
	if err := resp.Write(tlsConn); err != nil {
		log.Printf("✗ Error writing response: %v", err)
	}
}

// GetCacheStats returns cache statistics
func (p *ProxyServer) GetCacheStats() (hits, misses uint64, size int64, entries int, hitRate float64) {
	hits, misses, size, entries = p.httpCache.Stats()
	hitRate = p.httpCache.HitRate()
	return
}

// getEnvInt gets an integer from environment variable with default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvInt64 gets an int64 from environment variable with default
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration gets a duration from environment variable with default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
