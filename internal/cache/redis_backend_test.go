package cache

import (
	"testing"
	"time"
)

// Note: These tests require a running Redis instance
// Skip them in CI if Redis is not available

func TestRedisBackendBasic(t *testing.T) {
	t.Skip("Skipping Redis tests - requires Redis instance")

	// Create Redis backend
	backend, err := NewRedisBackend("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer backend.Close()

	// Test Set/Get
	entry := &CacheEntry{
		Key:      "test-key",
		Value:    []byte("test-value"),
		ExpireAt: time.Now().Add(1 * time.Hour),
	}

	err = backend.Set("test-key", entry, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	retrieved, err := backend.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get cache entry: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved entry is nil")
	}

	if string(retrieved.Value) != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", string(retrieved.Value))
	}

	// Test Delete
	err = backend.Delete("test-key")
	if err != nil {
		t.Fatalf("Failed to delete cache entry: %v", err)
	}

	retrieved, err = backend.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get deleted entry: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil for deleted entry")
	}
}

func TestRedisBackendStats(t *testing.T) {
	t.Skip("Skipping Redis tests - requires Redis instance")

	backend, err := NewRedisBackend("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer backend.Close()

	stats, err := backend.Stats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats is nil")
	}

	if _, ok := stats["db_size"]; !ok {
		t.Error("Stats should contain db_size")
	}
}
