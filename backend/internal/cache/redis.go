package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/anterajatech/warehouse-api/internal/config"
	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned when a key is not found in the cache.
var ErrCacheMiss = errors.New("cache: key not found")

// Cache wraps a Redis client. It is an optional optimization (nilai tambah);
// the service layer must still function correctly when caching is disabled.
type Cache struct {
	client *redis.Client
	enabled bool
}

// New returns a Cache instance. When Redis is disabled in config, all operations
// become no-ops so the rest of the application is unaffected.
func New(cfg *config.RedisConfig) *Cache {
	c := &Cache{enabled: cfg.Enabled}
	if !cfg.Enabled {
		return c
	}

	c.client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return c
}

// Ping verifies connectivity to Redis. It is a no-op when caching is disabled.
func (c *Cache) Ping(ctx context.Context) error {
	if !c.enabled {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

// Get retrieves and JSON-decodes a value. Returns ErrCacheMiss when the key
// does not exist.
func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	if !c.enabled {
		return ErrCacheMiss
	}

	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get: %w", err)
	}

	if err := json.Unmarshal(val, dest); err != nil {
		return fmt.Errorf("cache unmarshal: %w", err)
	}
	return nil
}

// Set JSON-encodes a value and stores it with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

// Delete removes a key. No-op when caching is disabled.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if !c.enabled {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// Enabled reports whether caching is active.
func (c *Cache) Enabled() bool {
	return c.enabled
}

// InvalidateItemCache clears all cached item-related keys for an item.
func (c *Cache) InvalidateItemCache(ctx context.Context, itemID int64) {
	if !c.enabled {
		return
	}
	c.Delete(ctx, fmt.Sprintf("item:%d", itemID))
	c.Delete(ctx, "items:list")
}
