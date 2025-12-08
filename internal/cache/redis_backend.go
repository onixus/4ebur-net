package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend implements distributed cache using Redis
type RedisBackend struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisBackend creates a new Redis cache backend
func NewRedisBackend(addr, password string, db int) (*RedisBackend, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     100, // Connection pool size
		MinIdleConns: 10,  // Minimum idle connections
		MaxRetries:   3,   // Retry failed operations
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisBackend{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get retrieves a cache entry from Redis
func (r *RedisBackend) Get(key string) (*CacheEntry, error) {
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	return &entry, nil
}

// Set stores a cache entry in Redis with TTL
func (r *RedisBackend) Set(key string, entry *CacheEntry, ttl time.Duration) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

// Delete removes a cache entry from Redis
func (r *RedisBackend) Delete(key string) error {
	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// Clear removes all cache entries (use with caution!)
func (r *RedisBackend) Clear() error {
	if err := r.client.FlushDB(r.ctx).Err(); err != nil {
		return fmt.Errorf("redis flush failed: %w", err)
	}
	return nil
}

// Stats returns Redis cache statistics
func (r *RedisBackend) Stats() (map[string]interface{}, error) {
	info, err := r.client.Info(r.ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("redis info failed: %w", err)
	}

	dbSize, err := r.client.DBSize(r.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis dbsize failed: %w", err)
	}

	return map[string]interface{}{
		"db_size": dbSize,
		"info":    info,
	}, nil
}

// PublishInvalidation publishes cache invalidation message to other instances
func (r *RedisBackend) PublishInvalidation(pattern string) error {
	if err := r.client.Publish(r.ctx, "cache:invalidate", pattern).Err(); err != nil {
		return fmt.Errorf("redis publish failed: %w", err)
	}
	return nil
}

// SubscribeInvalidations subscribes to cache invalidation messages
// The handler function is called for each invalidation message
func (r *RedisBackend) SubscribeInvalidations(handler func(string)) error {
	pubsub := r.client.Subscribe(r.ctx, "cache:invalidate")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		handler(msg.Payload)
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisBackend) Close() error {
	return r.client.Close()
}
