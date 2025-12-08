package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CacheEntry represents a cached HTTP response
type CacheEntry struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	CachedAt   time.Time
	ExpireAt   time.Time
	Size       int64
}

// HTTPCache manages HTTP response caching
type HTTPCache struct {
	mu         sync.RWMutex
	entries    map[string]*CacheEntry
	maxSize    int64 // Maximum cache size in bytes
	currentSize int64 // Current cache size
	maxAge     time.Duration
	hitCount   uint64
	missCount  uint64
}

// NewHTTPCache creates a new HTTP cache
func NewHTTPCache(maxSize int64, maxAge time.Duration) *HTTPCache {
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // 100MB default
	}
	if maxAge == 0 {
		maxAge = 5 * time.Minute // 5 minutes default
	}

	c := &HTTPCache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    maxSize,
		maxAge:     maxAge,
	}

	// Start cleanup goroutine
	go c.cleanupExpired()

	return c
}

// Get retrieves a cached response
func (c *HTTPCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		c.missCount++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpireAt) {
		c.missCount++
		return nil, false
	}

	c.hitCount++
	return entry, true
}

// Set stores a response in cache
func (c *HTTPCache) Set(key string, entry *CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if c.currentSize+entry.Size > c.maxSize {
		c.evictLRU(entry.Size)
	}

	// Remove old entry if exists
	if old, exists := c.entries[key]; exists {
		c.currentSize -= old.Size
	}

	// Add new entry
	c.entries[key] = entry
	c.currentSize += entry.Size

	return nil
}

// Delete removes an entry from cache
func (c *HTTPCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		c.currentSize -= entry.Size
		delete(c.entries, key)
	}
}

// Clear removes all entries
func (c *HTTPCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.currentSize = 0
}

// Stats returns cache statistics
func (c *HTTPCache) Stats() (hits, misses uint64, size int64, entries int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.hitCount, c.missCount, c.currentSize, len(c.entries)
}

// HitRate returns cache hit rate
func (c *HTTPCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hitCount + c.missCount
	if total == 0 {
		return 0.0
	}
	return float64(c.hitCount) / float64(total)
}

// evictLRU evicts least recently used entries
func (c *HTTPCache) evictLRU(needed int64) {
	var oldest *CacheEntry
	var oldestKey string

	for needed > 0 && len(c.entries) > 0 {
		// Find oldest entry
		for key, entry := range c.entries {
			if oldest == nil || entry.CachedAt.Before(oldest.CachedAt) {
				oldest = entry
				oldestKey = key
			}
		}

		if oldest != nil {
			c.currentSize -= oldest.Size
			needed -= oldest.Size
			delete(c.entries, oldestKey)
			oldest = nil
		}
	}
}

// cleanupExpired removes expired entries periodically
func (c *HTTPCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpireAt) {
				c.currentSize -= entry.Size
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// GenerateKey creates a cache key from request
func GenerateKey(r *http.Request) string {
	// Include method, URL, and relevant headers
	var b strings.Builder
	b.WriteString(r.Method)
	b.WriteString(":")
	b.WriteString(r.URL.String())

	// Include headers that affect response
	if accept := r.Header.Get("Accept"); accept != "" {
		b.WriteString(":")
		b.WriteString(accept)
	}
	if encoding := r.Header.Get("Accept-Encoding"); encoding != "" {
		b.WriteString(":")
		b.WriteString(encoding)
	}

	// Hash the key for shorter storage
	hash := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(hash[:])
}

// IsCacheable checks if response can be cached
func IsCacheable(r *http.Request, resp *http.Response) bool {
	// Only cache GET requests
	if r.Method != http.MethodGet {
		return false
	}

	// Check status code (only cache successful responses)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	// Check Cache-Control headers
	cacheControl := resp.Header.Get("Cache-Control")
	if strings.Contains(cacheControl, "no-store") || strings.Contains(cacheControl, "no-cache") {
		return false
	}
	if strings.Contains(cacheControl, "private") {
		return false
	}

	// Check for Authorization header (don't cache authenticated requests by default)
	if r.Header.Get("Authorization") != "" {
		return false
	}

	return true
}

// GetMaxAge extracts max-age from Cache-Control header
func GetMaxAge(resp *http.Response, defaultAge time.Duration) time.Duration {
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "" {
		return defaultAge
	}

	// Parse max-age directive
	for _, directive := range strings.Split(cacheControl, ",") {
		directive = strings.TrimSpace(directive)
		if strings.HasPrefix(directive, "max-age=") {
			maxAgeStr := strings.TrimPrefix(directive, "max-age=")
			if seconds, err := strconv.ParseInt(maxAgeStr, 10, 64); err == nil {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	return defaultAge
}

// CreateCacheEntry creates a cache entry from response
func CreateCacheEntry(resp *http.Response, maxAge time.Duration) (*CacheEntry, error) {
	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body.Close()

	// Replace body for further use
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// Get actual max-age from headers
	actualMaxAge := GetMaxAge(resp, maxAge)

	now := time.Now()
	entry := &CacheEntry{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
		CachedAt:   now,
		ExpireAt:   now.Add(actualMaxAge),
		Size:       int64(len(body)),
	}

	return entry, nil
}

// WriteToResponse writes cache entry to response writer
func (e *CacheEntry) WriteToResponse(w http.ResponseWriter) error {
	// Copy headers
	for key, values := range e.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Add cache hit header
	w.Header().Set("X-Cache", "HIT")
	w.Header().Set("X-Cache-Age", fmt.Sprintf("%.0f", time.Since(e.CachedAt).Seconds()))

	// Write status code
	w.WriteHeader(e.StatusCode)

	// Write body
	_, err := w.Write(e.Body)
	return err
}
