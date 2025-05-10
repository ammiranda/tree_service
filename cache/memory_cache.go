package cache

import (
	"fmt"
	"sync"
	"time"
)

// MemoryCache implements CacheProvider using in-memory storage
type MemoryCache struct {
	mu       sync.RWMutex
	data     map[string]*PaginatedTreeResponse
	ttl      time.Duration
	expiries map[string]time.Time
}

// NewMemoryCache creates a new in-memory cache provider
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		ttl:      5 * time.Minute,
		data:     make(map[string]*PaginatedTreeResponse),
		expiries: make(map[string]time.Time),
	}
}

// Initialize performs any necessary setup for the cache provider
func (c *MemoryCache) Initialize() error {
	return nil
}

// getCacheKey generates a cache key for the given page and pageSize
func getCacheKey(page, pageSize int) string {
	return fmt.Sprintf("tree:%d:%d", page, pageSize)
}

// GetPaginatedTree retrieves the paginated tree from cache if available
func (c *MemoryCache) GetPaginatedTree(page, pageSize int) (*PaginatedTreeResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := getCacheKey(page, pageSize)
	expiry, exists := c.expiries[key]
	if !exists || time.Now().After(expiry) {
		return nil, false
	}

	if response, ok := c.data[key]; ok {
		return response, true
	}

	return nil, false
}

// SetPaginatedTree stores the paginated tree in cache
func (c *MemoryCache) SetPaginatedTree(page, pageSize int, response *PaginatedTreeResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getCacheKey(page, pageSize)
	c.data[key] = response
	c.expiries[key] = time.Now().Add(c.ttl)
}

// InvalidateCache removes all cached data
func (c *MemoryCache) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*PaginatedTreeResponse)
	c.expiries = make(map[string]time.Time)
}

// SetCacheTTL sets the cache time-to-live duration
func (c *MemoryCache) SetCacheTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
	// Update all existing expiries
	now := time.Now()
	for key := range c.data {
		c.expiries[key] = now.Add(ttl)
	}
}
