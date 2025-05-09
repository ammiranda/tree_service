package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/ammiranda/tree_service/models"
)

// MockCache is a cache provider that can be used for testing
type MockCache struct {
	mu              sync.RWMutex
	data            []*models.Node
	ttl             time.Duration
	expiry          time.Time
	GetTreeCalls    int
	SetTreeCalls    int
	InvalidateCalls int
	SetTTLCalls     int
	InitCalls       int
	ShouldFail      bool
}

// NewMockCache creates a new mock cache provider
func NewMockCache() *MockCache {
	return &MockCache{
		ttl: 5 * time.Minute,
	}
}

// Initialize performs any necessary setup for the cache provider
func (c *MockCache) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.InitCalls++
	if c.ShouldFail {
		return ErrCacheInitialization
	}
	return nil
}

// GetTree retrieves the tree from cache if available
func (c *MockCache) GetTree() ([]*models.Node, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetTreeCalls++

	if c.ShouldFail {
		return nil, false
	}

	if c.data == nil || time.Now().After(c.expiry) {
		return nil, false
	}

	return c.data, true
}

// SetTree stores the tree in cache
func (c *MockCache) SetTree(tree []*models.Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SetTreeCalls++

	if !c.ShouldFail {
		c.data = tree
		c.expiry = time.Now().Add(c.ttl)
	}
}

// InvalidateCache removes the tree from cache
func (c *MockCache) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.InvalidateCalls++

	if !c.ShouldFail {
		c.data = nil
		c.expiry = time.Time{}
	}
}

// SetCacheTTL sets the cache time-to-live duration
func (c *MockCache) SetCacheTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SetTTLCalls++

	if !c.ShouldFail {
		c.ttl = ttl
		if c.data != nil {
			c.expiry = time.Now().Add(ttl)
		}
	}
}

// Reset resets all counters and state
func (c *MockCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetTreeCalls = 0
	c.SetTreeCalls = 0
	c.InvalidateCalls = 0
	c.SetTTLCalls = 0
	c.InitCalls = 0
	c.ShouldFail = false
	c.data = nil
	c.expiry = time.Time{}
}

// GetCallCounts returns the number of times each method was called
func (c *MockCache) GetCallCounts() (getTree, setTree, invalidate, setTTL, init int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetTreeCalls, c.SetTreeCalls, c.InvalidateCalls, c.SetTTLCalls, c.InitCalls
}

// SetShouldFail makes the mock cache fail all operations
func (c *MockCache) SetShouldFail(shouldFail bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ShouldFail = shouldFail
}

// ErrCacheInitialization is returned when the mock cache is configured to fail
var ErrCacheInitialization = errors.New("mock cache initialization failed")
