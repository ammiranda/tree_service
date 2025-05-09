package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ammiranda/tree_service/cache"
	"github.com/ammiranda/tree_service/models"
	"github.com/ammiranda/tree_service/repository"
)

func TestCache(t *testing.T) {
	// Create mock repository
	repo := repository.NewMockRepository()
	err := repo.Initialize(context.Background())
	assert.NoError(t, err)
	defer repo.Cleanup(context.Background())

	// Create initial root node
	id, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

	// Create cache provider
	cacheProvider := cache.NewMemoryCache()
	err = cacheProvider.Initialize()
	assert.NoError(t, err)

	// Test caching
	tree, found := cacheProvider.GetTree()
	assert.False(t, found)
	assert.Nil(t, tree)

	// Set tree in cache
	nodes := []*models.Node{
		{
			ID:       id,
			Label:    "root",
			Children: make([]*models.Node, 0),
		},
	}
	cacheProvider.SetTree(nodes)

	// Get tree from cache
	tree, found = cacheProvider.GetTree()
	assert.True(t, found)
	assert.NotNil(t, tree)
	assert.Len(t, tree, 1)
	assert.Equal(t, "root", tree[0].Label)

	// Test cache invalidation
	cacheProvider.InvalidateCache()
	tree, found = cacheProvider.GetTree()
	assert.False(t, found)
	assert.Nil(t, tree)
}

func TestCacheTTL(t *testing.T) {
	// Create cache provider
	cacheProvider := cache.NewMemoryCache()
	err := cacheProvider.Initialize()
	assert.NoError(t, err)

	// Set TTL to 1 second
	cacheProvider.SetCacheTTL(time.Second)

	// Set tree in cache
	nodes := []*models.Node{
		{
			ID:       1,
			Label:    "root",
			Children: make([]*models.Node, 0),
		},
	}
	cacheProvider.SetTree(nodes)

	// Get tree from cache immediately
	tree, found := cacheProvider.GetTree()
	assert.True(t, found)
	assert.NotNil(t, tree)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Tree should be gone from cache
	tree, found = cacheProvider.GetTree()
	assert.False(t, found)
	assert.Nil(t, tree)
}
