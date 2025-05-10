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
	defer func() {
		if err := repo.Cleanup(context.Background()); err != nil {
			t.Errorf("Failed to cleanup repository: %v", err)
		}
	}()

	// Create initial root node
	id, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

	// Create cache provider
	cacheProvider := cache.NewMemoryCache()
	err = cacheProvider.Initialize()
	assert.NoError(t, err)

	// Test caching
	response, found := cacheProvider.GetPaginatedTree(1, 10)
	assert.False(t, found)
	assert.Nil(t, response)

	// Create test data
	nodes := []*models.Node{
		{
			ID:       id,
			Label:    "root",
			Children: make([]*models.Node, 0),
		},
	}

	// Create paginated response
	testResponse := &cache.PaginatedTreeResponse{
		Data: nodes,
	}
	testResponse.Pagination.Page = 1
	testResponse.Pagination.PageSize = 10
	testResponse.Pagination.Total = 1
	testResponse.Pagination.TotalPages = 1
	testResponse.Pagination.HasNext = false
	testResponse.Pagination.HasPrev = false

	// Set response in cache
	cacheProvider.SetPaginatedTree(1, 10, testResponse)

	// Get response from cache
	response, found = cacheProvider.GetPaginatedTree(1, 10)
	assert.True(t, found)
	assert.NotNil(t, response)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "root", response.Data[0].Label)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.PageSize)
	assert.Equal(t, int64(1), response.Pagination.Total)
	assert.Equal(t, int64(1), response.Pagination.TotalPages)
	assert.False(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)

	// Test different page size
	response, found = cacheProvider.GetPaginatedTree(1, 20)
	assert.False(t, found)
	assert.Nil(t, response)

	// Test cache invalidation
	cacheProvider.InvalidateCache()
	response, found = cacheProvider.GetPaginatedTree(1, 10)
	assert.False(t, found)
	assert.Nil(t, response)
}

func TestCacheTTL(t *testing.T) {
	// Create cache provider
	cacheProvider := cache.NewMemoryCache()
	err := cacheProvider.Initialize()
	assert.NoError(t, err)

	// Set TTL to 1 second
	cacheProvider.SetCacheTTL(time.Second)

	// Create test data
	nodes := []*models.Node{
		{
			ID:       1,
			Label:    "root",
			Children: make([]*models.Node, 0),
		},
	}

	// Create paginated response
	testResponse := &cache.PaginatedTreeResponse{
		Data: nodes,
	}
	testResponse.Pagination.Page = 1
	testResponse.Pagination.PageSize = 10
	testResponse.Pagination.Total = 1
	testResponse.Pagination.TotalPages = 1
	testResponse.Pagination.HasNext = false
	testResponse.Pagination.HasPrev = false

	// Set response in cache
	cacheProvider.SetPaginatedTree(1, 10, testResponse)

	// Get response from cache immediately
	response, found := cacheProvider.GetPaginatedTree(1, 10)
	assert.True(t, found)
	assert.NotNil(t, response)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Response should be gone from cache
	response, found = cacheProvider.GetPaginatedTree(1, 10)
	assert.False(t, found)
	assert.Nil(t, response)
}

func TestMultiplePages(t *testing.T) {
	// Create cache provider
	cacheProvider := cache.NewMemoryCache()
	err := cacheProvider.Initialize()
	assert.NoError(t, err)

	// Create test data for page 1
	page1Response := &cache.PaginatedTreeResponse{
		Data: []*models.Node{
			{
				ID:       1,
				Label:    "node1",
				Children: make([]*models.Node, 0),
			},
			{
				ID:       2,
				Label:    "node2",
				Children: make([]*models.Node, 0),
			},
		},
	}
	page1Response.Pagination.Page = 1
	page1Response.Pagination.PageSize = 2
	page1Response.Pagination.Total = 4
	page1Response.Pagination.TotalPages = 2
	page1Response.Pagination.HasNext = true
	page1Response.Pagination.HasPrev = false

	// Create test data for page 2
	page2Response := &cache.PaginatedTreeResponse{
		Data: []*models.Node{
			{
				ID:       3,
				Label:    "node3",
				Children: make([]*models.Node, 0),
			},
			{
				ID:       4,
				Label:    "node4",
				Children: make([]*models.Node, 0),
			},
		},
	}
	page2Response.Pagination.Page = 2
	page2Response.Pagination.PageSize = 2
	page2Response.Pagination.Total = 4
	page2Response.Pagination.TotalPages = 2
	page2Response.Pagination.HasNext = false
	page2Response.Pagination.HasPrev = true

	// Set both pages in cache
	cacheProvider.SetPaginatedTree(1, 2, page1Response)
	cacheProvider.SetPaginatedTree(2, 2, page2Response)

	// Test retrieving page 1
	response, found := cacheProvider.GetPaginatedTree(1, 2)
	assert.True(t, found)
	assert.NotNil(t, response)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, "node1", response.Data[0].Label)
	assert.Equal(t, "node2", response.Data[1].Label)
	assert.True(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)

	// Test retrieving page 2
	response, found = cacheProvider.GetPaginatedTree(2, 2)
	assert.True(t, found)
	assert.NotNil(t, response)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, "node3", response.Data[0].Label)
	assert.Equal(t, "node4", response.Data[1].Label)
	assert.False(t, response.Pagination.HasNext)
	assert.True(t, response.Pagination.HasPrev)

	// Test cache invalidation affects all pages
	cacheProvider.InvalidateCache()
	response, found = cacheProvider.GetPaginatedTree(1, 2)
	assert.False(t, found)
	assert.Nil(t, response)
	response, found = cacheProvider.GetPaginatedTree(2, 2)
	assert.False(t, found)
	assert.Nil(t, response)
}
