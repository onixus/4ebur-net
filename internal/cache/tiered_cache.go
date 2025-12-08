package cache

import (
	"sync"
	"time"
)

// TieredCache implements multi-tier caching (L1: Memory, L2: Redis)
type TieredCache struct {
	l1         *HTTPCache
	l2         *RedisBackend
	maxAge     time.Duration
	mu         sync.RWMutex
	hitCount   uint64
	missCount  uint64
	l1Hits     uint64
	l2Hits     uint64
}

// NewTieredCache creates a new tiered cache
func NewTieredCache(l1 *HTTPCache, l2 *RedisBackend, maxAge time.Duration) *TieredCache {
	return &TieredCache{
		l1:     l1,
		l2:     l2,
		maxAge: maxAge,
	}
}

// Get retrieves entry from L1 (fast) then L2 (slower)
func (c *TieredCache) Get(key string) (*CacheEntry, bool) {
	// Try L1 cache (in-memory, fast)
	if entry, found := c.l1.Get(key); found {
		c.mu.Lock()
		c.hitCount++
		c.l1Hits++
		c.mu.Unlock()
		return entry, true
	}

	// Try L2 cache (Redis, slower but distributed)
	if c.l2 != nil {
		entry, err := c.l2.Get(key)
		if err == nil && entry != nil {
			// Promote to L1 for faster future access
			c.l1.Set(key, entry)

			c.mu.Lock()
			c.hitCount++
			c.l2Hits++
			c.mu.Unlock()

			return entry, true
		}
	}

	// Cache miss
	c.mu.Lock()
	c.missCount++
	c.mu.Unlock()

	return nil, false
}

// Set stores entry in both L1 and L2
func (c *TieredCache) Set(key string, entry *CacheEntry) error {
	// Always store in L1
	c.l1.Set(key, entry)

	// Store in L2 if available
	if c.l2 != nil {
		ttl := time.Until(entry.ExpireAt)
		if ttl > 0 {
			return c.l2.Set(key, entry, ttl)
		}
	}

	return nil
}

// Delete removes entry from both tiers
func (c *TieredCache) Delete(key string) {
	c.l1.Delete(key)

	if c.l2 != nil {
		c.l2.Delete(key)
	}
}

// Clear removes all entries from both tiers
func (c *TieredCache) Clear() {
	c.l1.Clear()

	if c.l2 != nil {
		c.l2.Clear()
	}
}

// Stats returns tiered cache statistics
func (c *TieredCache) Stats() (hits, misses, l1Hits, l2Hits uint64, hitRate float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hits = c.hitCount
	misses = c.missCount
	l1Hits = c.l1Hits
	l2Hits = c.l2Hits

	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return
}

// InvalidatePattern invalidates cache entries matching pattern
// Publishes invalidation to other instances via Redis Pub/Sub
func (c *TieredCache) InvalidatePattern(pattern string) error {
	// Clear local caches
	// Note: This is simplified - production would use pattern matching
	c.Clear()

	// Publish to other instances
	if c.l2 != nil {
		return c.l2.PublishInvalidation(pattern)
	}

	return nil
}
