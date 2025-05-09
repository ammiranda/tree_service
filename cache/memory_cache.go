package cache

import (
	"sync"
	"time"

	"github.com/ammiranda/tree_service/models"
)

// MemoryCache implements CacheProvider using in-memory storage
type MemoryCache struct {
	mu     sync.RWMutex
	data   []*models.Node
	ttl    time.Duration
	expiry time.Time
}

// NewMemoryCache creates a new in-memory cache provider
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		ttl: 5 * time.Minute,
	}
}

// Initialize performs any necessary setup for the cache provider
func (c *MemoryCache) Initialize() error {
	return nil
}

// GetTree retrieves the tree from cache if available
func (c *MemoryCache) GetTree() ([]*models.Node, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil || time.Now().After(c.expiry) {
		return nil, false
	}

	return c.data, true
}

// SetTree stores the tree in cache
func (c *MemoryCache) SetTree(tree []*models.Node) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = tree
	c.expiry = time.Now().Add(c.ttl)
}

// InvalidateCache removes the tree from cache
func (c *MemoryCache) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
	c.expiry = time.Time{}
}

// SetCacheTTL sets the cache time-to-live duration
func (c *MemoryCache) SetCacheTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
	if c.data != nil {
		c.expiry = time.Now().Add(ttl)
	}
}
