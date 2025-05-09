package cache

import (
	"os"
	"sync"
	"time"

	"github.com/ammiranda/tree_service/models"
)

var (
	provider CacheProvider
	once     sync.Once
	mu       sync.RWMutex
)

// CacheProvider defines the interface for cache implementations.
// It provides methods for caching and retrieving tree structures.
type CacheProvider interface {
	// GetTree retrieves the tree from cache if available.
	// Returns:
	//   - A slice of nodes representing the tree structure
	//   - A boolean indicating whether the tree was found in cache
	GetTree() ([]*models.Node, bool)

	// SetTree stores the tree in cache.
	// Parameters:
	//   - tree: The tree structure to cache
	SetTree(tree []*models.Node)

	// InvalidateCache removes the tree from cache.
	// This is typically called when the tree structure is modified.
	InvalidateCache()

	// SetCacheTTL sets the cache time-to-live duration.
	// Parameters:
	//   - ttl: The duration after which cached data should expire
	SetCacheTTL(ttl time.Duration)

	// Initialize performs any necessary setup for the cache provider.
	// This may include establishing connections, creating cache instances,
	// or any other initialization required for the cache to function.
	// Returns an error if initialization fails.
	Initialize() error
}

// Initialize sets up the cache provider
func Initialize() error {
	var err error
	once.Do(func() {
		// Use Redis in local development, MemoryCache otherwise
		if os.Getenv("REDIS_HOST") != "" {
			provider = NewRedisCache()
		} else {
			provider = NewMemoryCache()
		}
		err = provider.Initialize()
	})
	return err
}

// GetTree retrieves the tree from cache if available
func GetTree() ([]*models.Node, bool) {
	mu.RLock()
	defer mu.RUnlock()
	return provider.GetTree()
}

// SetTree stores the tree in cache
func SetTree(tree []*models.Node) {
	mu.Lock()
	defer mu.Unlock()
	provider.SetTree(tree)
}

// InvalidateCache removes the tree from cache
func InvalidateCache() {
	mu.Lock()
	defer mu.Unlock()
	provider.InvalidateCache()
}

// SetCacheTTL sets the cache time-to-live duration
func SetCacheTTL(ttl time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	provider.SetCacheTTL(ttl)
}

// SetProvider allows changing the cache provider at runtime
func SetProvider(p CacheProvider) error {
	mu.Lock()
	defer mu.Unlock()
	if err := p.Initialize(); err != nil {
		return err
	}
	provider = p
	return nil
}

// ResetProvider resets the cache provider for testing
func ResetProvider() {
	mu.Lock()
	defer mu.Unlock()
	provider = nil
	once = sync.Once{}
}
