package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"theary_test/cache"
	"theary_test/database"
	"theary_test/models"

	"github.com/stretchr/testify/assert"
)

func TestDynamoDBCache(t *testing.T) {
	// Create DynamoDB cache provider with mock client
	mockClient := cache.NewMockDynamoDBClient()
	dynamoCache := cache.NewDynamoDBCacheWithClient(mockClient)
	assert.NoError(t, dynamoCache.Initialize())

	testCacheProvider(t, dynamoCache)
}

func TestMemoryCache(t *testing.T) {
	// Create in-memory cache provider
	memoryCache := cache.NewMemoryCache()
	assert.NoError(t, memoryCache.Initialize())

	testCacheProvider(t, memoryCache)
}

func TestMockCache(t *testing.T) {
	// Create mock cache provider
	mockCache := cache.NewMockCache()
	assert.NoError(t, mockCache.Initialize())

	// Test basic functionality
	testCacheProvider(t, mockCache)

	// Test call counts
	getTree, setTree, invalidate, setTTL, init := mockCache.GetCallCounts()
	assert.Greater(t, getTree, 0, "GetTree should have been called")
	assert.Greater(t, setTree, 0, "SetTree should have been called")
	assert.Greater(t, invalidate, 0, "InvalidateCache should have been called")
	assert.Greater(t, setTTL, 0, "SetCacheTTL should have been called")
	assert.Equal(t, 1, init, "Initialize should have been called once")

	// Test failure mode
	mockCache.Reset()
	mockCache.SetShouldFail(true)
	assert.Error(t, mockCache.Initialize(), "Initialize should fail when ShouldFail is true")
	tree, found := mockCache.GetTree()
	assert.Nil(t, tree, "GetTree should return nil when ShouldFail is true")
	assert.False(t, found, "GetTree should return false when ShouldFail is true")

	// Test reset functionality
	mockCache.Reset()
	getTree, setTree, invalidate, setTTL, init = mockCache.GetCallCounts()
	assert.Equal(t, 0, getTree, "GetTree calls should be reset")
	assert.Equal(t, 0, setTree, "SetTree calls should be reset")
	assert.Equal(t, 0, invalidate, "InvalidateCache calls should be reset")
	assert.Equal(t, 0, setTTL, "SetCacheTTL calls should be reset")
	assert.Equal(t, 0, init, "Initialize calls should be reset")
	assert.False(t, mockCache.ShouldFail, "ShouldFail should be reset")
}

func testCacheProvider(t *testing.T, provider cache.CacheProvider) {
	// Test setting and getting tree
	tree := []*models.Node{
		{
			ID:       1,
			Label:    "Root",
			Children: []*models.Node{},
		},
	}

	// Test SetTree and GetTree
	provider.SetTree(tree)
	cachedTree, found := provider.GetTree()
	assert.True(t, found)
	assert.Equal(t, tree, cachedTree)

	// Test cache invalidation
	provider.InvalidateCache()
	_, found = provider.GetTree()
	assert.False(t, found)

	// Test cache expiration
	provider.SetCacheTTL(1 * time.Second)
	provider.SetTree(tree)
	time.Sleep(2 * time.Second)
	_, found = provider.GetTree()
	assert.False(t, found)
}

func TestCachingInTreeAPI(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Initialize with mock cache for testing
	mockCache := cache.NewMockCache()
	assert.NoError(t, cache.SetProvider(mockCache))
	cache.SetCacheTTL(100 * time.Millisecond)

	// Create initial root node
	db := database.GetDB()
	result, err := db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", "root", nil)
	assert.NoError(t, err)
	rootID, err := result.LastInsertId()
	assert.NoError(t, err)

	r := setupTestRouter()

	// First request should miss cache
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/tree", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should hit cache
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/tree", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())

	// Create a new node should invalidate cache
	payload := models.CreateNodeRequest{
		Label:    "child",
		ParentID: rootID,
	}
	jsonPayload, _ := json.Marshal(payload)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/tree", bytes.NewBuffer(jsonPayload))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusCreated, w3.Code)

	// Next GET request should miss cache and return updated tree
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/api/tree", nil)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
	assert.NotEqual(t, w1.Body.String(), w4.Body.String())

	// Verify cache operations
	getTree, setTree, invalidate, _, _ := mockCache.GetCallCounts()
	assert.Greater(t, getTree, 0, "GetTree should have been called")
	assert.Greater(t, setTree, 0, "SetTree should have been called")
	assert.Greater(t, invalidate, 0, "InvalidateCache should have been called")
}
