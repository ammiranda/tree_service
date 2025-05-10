package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements CacheProvider using Redis
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache provider
func NewRedisCache() *RedisCache {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &RedisCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

// Initialize performs any necessary setup for the cache provider
func (c *RedisCache) Initialize() error {
	ctx := context.Background()
	_, err := c.client.Ping(ctx).Result()
	return err
}

// getRedisKey generates a cache key for the given page and pageSize
func getRedisKey(page, pageSize int) string {
	return fmt.Sprintf("tree:%d:%d", page, pageSize)
}

// GetPaginatedTree retrieves the paginated tree from cache if available
func (c *RedisCache) GetPaginatedTree(page, pageSize int) (*PaginatedTreeResponse, bool) {
	ctx := context.Background()
	key := getRedisKey(page, pageSize)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var response PaginatedTreeResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, false
	}

	return &response, true
}

// SetPaginatedTree stores the paginated tree in cache
func (c *RedisCache) SetPaginatedTree(page, pageSize int, response *PaginatedTreeResponse) {
	ctx := context.Background()
	key := getRedisKey(page, pageSize)

	data, err := json.Marshal(response)
	if err != nil {
		return
	}

	c.client.Set(ctx, key, data, c.ttl)
}

// InvalidateCache removes all cached data
func (c *RedisCache) InvalidateCache() {
	ctx := context.Background()
	// Use scan to find and delete all tree:* keys
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(ctx, cursor, "tree:*", 100).Result()
		if err != nil {
			return
		}

		if len(keys) > 0 {
			c.client.Del(ctx, keys...)
		}

		if cursor == 0 {
			break
		}
	}
}

// SetCacheTTL sets the cache time-to-live duration
func (c *RedisCache) SetCacheTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
