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

// PaginatedTreeResponse represents a paginated tree response
type PaginatedTreeResponse struct {
	Data       []*models.Node `json:"data"`
	Pagination struct {
		Page       int   `json:"page"`
		PageSize   int   `json:"pageSize"`
		Total      int64 `json:"total"`
		TotalPages int64 `json:"totalPages"`
		HasNext    bool  `json:"hasNext"`
		HasPrev    bool  `json:"hasPrev"`
	} `json:"pagination"`
}

// CacheProvider defines the interface for cache implementations.
// It provides methods for caching and retrieving tree structures.
type CacheProvider interface {
	// GetPaginatedTree retrieves the paginated tree from cache if available.
	// Parameters:
	//   - page: The page number
	//   - pageSize: The size of each page
	// Returns:
	//   - The paginated tree response
	//   - A boolean indicating whether the response was found in cache
	GetPaginatedTree(page, pageSize int) (*PaginatedTreeResponse, bool)

	// SetPaginatedTree stores the paginated tree in cache.
	// Parameters:
	//   - page: The page number
	//   - pageSize: The size of each page
	//   - response: The paginated tree response to cache
	SetPaginatedTree(page, pageSize int, response *PaginatedTreeResponse)

	// InvalidateCache removes all cached data.
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

// GetPaginatedTree retrieves the paginated tree from cache if available
func GetPaginatedTree(page, pageSize int) (*PaginatedTreeResponse, bool) {
	mu.RLock()
	defer mu.RUnlock()
	return provider.GetPaginatedTree(page, pageSize)
}

// SetPaginatedTree stores the paginated tree in cache
func SetPaginatedTree(page, pageSize int, response *PaginatedTreeResponse) {
	mu.Lock()
	defer mu.Unlock()
	provider.SetPaginatedTree(page, pageSize, response)
}

// InvalidateCache removes all cached data
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
