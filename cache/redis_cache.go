package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"theary_test/models"

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

// GetTree retrieves the tree from cache if available
func (c *RedisCache) GetTree() ([]*models.Node, bool) {
	ctx := context.Background()
	data, err := c.client.Get(ctx, "tree").Result()
	if err != nil {
		return nil, false
	}

	var nodes []*models.Node
	if err := json.Unmarshal([]byte(data), &nodes); err != nil {
		return nil, false
	}

	return nodes, true
}

// SetTree stores the tree in cache
func (c *RedisCache) SetTree(tree []*models.Node) {
	ctx := context.Background()
	data, err := json.Marshal(tree)
	if err != nil {
		return
	}

	c.client.Set(ctx, "tree", data, c.ttl)
}

// InvalidateCache removes the tree from cache
func (c *RedisCache) InvalidateCache() {
	ctx := context.Background()
	c.client.Del(ctx, "tree")
}

// SetCacheTTL sets the cache time-to-live duration
func (c *RedisCache) SetCacheTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
