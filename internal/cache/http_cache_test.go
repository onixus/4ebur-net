package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPCache(t *testing.T) {
	cache := NewHTTPCache(1024*1024, 5*time.Minute)
	if cache == nil {
		t.Fatal("NewHTTPCache returned nil")
	}

	if cache.maxSize != 1024*1024 {
		t.Errorf("Expected maxSize 1MB, got %d", cache.maxSize)
	}

	if cache.maxAge != 5*time.Minute {
		t.Errorf("Expected maxAge 5m, got %v", cache.maxAge)
	}
}

func TestCacheSetGet(t *testing.T) {
	cache := NewHTTPCache(1024*1024, 5*time.Minute)

	entry := &CacheEntry{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"text/html"}},
		Body:       []byte("test body"),
		CachedAt:   time.Now(),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       9,
	}

	key := "test-key"
	err := cache.Set(key, entry)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, found := cache.Get(key)
	if !found {
		t.Fatal("Entry not found in cache")
	}

	if retrieved.StatusCode != entry.StatusCode {
		t.Errorf("StatusCode mismatch: expected %d, got %d", entry.StatusCode, retrieved.StatusCode)
	}

	if !bytes.Equal(retrieved.Body, entry.Body) {
		t.Error("Body mismatch")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewHTTPCache(1024*1024, 1*time.Second)

	entry := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("test"),
		CachedAt:   time.Now(),
		ExpireAt:   time.Now().Add(100 * time.Millisecond),
		Size:       4,
	}

	key := "expire-test"
	cache.Set(key, entry)

	// Should be found immediately
	if _, found := cache.Get(key); !found {
		t.Error("Entry should be found before expiration")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiration
	if _, found := cache.Get(key); found {
		t.Error("Entry should be expired")
	}
}

func TestCacheEviction(t *testing.T) {
	// Small cache that can hold only 2 entries
	cache := NewHTTPCache(20, 5*time.Minute)

	entry1 := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("entry1"),
		CachedAt:   time.Now(),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       6,
	}

	entry2 := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("entry2"),
		CachedAt:   time.Now().Add(1 * time.Second),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       6,
	}

	entry3 := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("entry3"),
		CachedAt:   time.Now().Add(2 * time.Second),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       10,
	}

	cache.Set("key1", entry1)
	cache.Set("key2", entry2)

	// Adding entry3 should evict entry1 (oldest)
	cache.Set("key3", entry3)

	if _, found := cache.Get("key1"); found {
		t.Error("Entry1 should have been evicted")
	}

	if _, found := cache.Get("key3"); !found {
		t.Error("Entry3 should be in cache")
	}
}

func TestGenerateKey(t *testing.T) {
	req1 := httptest.NewRequest("GET", "http://example.com/path", nil)
	req2 := httptest.NewRequest("GET", "http://example.com/path", nil)
	req3 := httptest.NewRequest("POST", "http://example.com/path", nil)

	key1 := GenerateKey(req1)
	key2 := GenerateKey(req2)
	key3 := GenerateKey(req3)

	// Same requests should have same keys
	if key1 != key2 {
		t.Error("Same requests should have same keys")
	}

	// Different methods should have different keys
	if key1 == key3 {
		t.Error("Different methods should have different keys")
	}
}

func TestIsCacheable(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		statusCode int
		headers    map[string]string
		cacheable  bool
	}{
		{
			name:       "GET 200 OK",
			method:     "GET",
			statusCode: 200,
			headers:    map[string]string{},
			cacheable:  true,
		},
		{
			name:       "POST request",
			method:     "POST",
			statusCode: 200,
			cacheable:  false,
		},
		{
			name:       "404 response",
			method:     "GET",
			statusCode: 404,
			cacheable:  false,
		},
		{
			name:       "Cache-Control: no-store",
			method:     "GET",
			statusCode: 200,
			headers:    map[string]string{"Cache-Control": "no-store"},
			cacheable:  false,
		},
		{
			name:       "Cache-Control: private",
			method:     "GET",
			statusCode: 200,
			headers:    map[string]string{"Cache-Control": "private"},
			cacheable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "http://example.com", nil)
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}

			for k, v := range tt.headers {
				resp.Header.Set(k, v)
			}

			if got := IsCacheable(req, resp); got != tt.cacheable {
				t.Errorf("IsCacheable() = %v, want %v", got, tt.cacheable)
			}
		})
	}
}

func TestCacheStats(t *testing.T) {
	cache := NewHTTPCache(1024*1024, 5*time.Minute)

	entry := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("test"),
		CachedAt:   time.Now(),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       4,
	}

	cache.Set("key1", entry)

	// Hit
	cache.Get("key1")

	// Miss
	cache.Get("nonexistent")

	hits, misses, _, _ := cache.Stats()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}

	hitRate := cache.HitRate()
	if hitRate != 0.5 {
		t.Errorf("Expected hit rate 0.5, got %f", hitRate)
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache := NewHTTPCache(100*1024*1024, 5*time.Minute)
	entry := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("benchmark test body"),
		CachedAt:   time.Now(),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       19,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := GenerateKey(httptest.NewRequest("GET", "http://example.com", nil))
		cache.Set(key, entry)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewHTTPCache(100*1024*1024, 5*time.Minute)
	entry := &CacheEntry{
		StatusCode: 200,
		Body:       []byte("benchmark test body"),
		CachedAt:   time.Now(),
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Size:       19,
	}
	key := "benchmark-key"
	cache.Set(key, entry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(key)
	}
}
