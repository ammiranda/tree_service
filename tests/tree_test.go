package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ammiranda/tree_service/cache"
	"github.com/ammiranda/tree_service/handlers"
	"github.com/ammiranda/tree_service/models"
	"github.com/ammiranda/tree_service/repository"
)

func setupTest(t *testing.T) (*repository.MockRepository, func()) {
	// Create mock repository
	repo := repository.NewMockRepository()
	err := repo.Initialize(context.Background())
	assert.NoError(t, err)

	// Initialize cache with memory provider
	err = cache.SetProvider(cache.NewMemoryCache())
	assert.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		if err := repo.Cleanup(context.Background()); err != nil {
			t.Errorf("Failed to cleanup repository: %v", err)
		}
		cache.ResetProvider()
	}

	return repo, cleanup
}

func TestGetTree(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create initial root node
	rootID, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

	// Create some child nodes
	for i := 1; i <= 15; i++ {
		_, err := repo.CreateNode(context.Background(), fmt.Sprintf("child_%d", i), &rootID)
		assert.NoError(t, err)
	}

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.GET("/tree", handler.GetTree)

	// Test default pagination (page 1, pageSize 10)
	req, _ := http.NewRequest("GET", "/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response cache.PaginatedTreeResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check pagination metadata
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.PageSize)
	assert.Equal(t, int64(16), response.Pagination.Total) // root + 15 children
	assert.Equal(t, int64(2), response.Pagination.TotalPages)
	assert.True(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)

	// Test second page
	req, _ = http.NewRequest("GET", "/tree?page=2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check pagination metadata for second page
	assert.Equal(t, 2, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.PageSize)
	assert.Equal(t, int64(16), response.Pagination.Total)
	assert.Equal(t, int64(2), response.Pagination.TotalPages)
	assert.False(t, response.Pagination.HasNext)
	assert.True(t, response.Pagination.HasPrev)

	// Test custom page size
	req, _ = http.NewRequest("GET", "/tree?pageSize=5", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check pagination metadata for custom page size
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 5, response.Pagination.PageSize)
	assert.Equal(t, int64(16), response.Pagination.Total)
	assert.Equal(t, int64(4), response.Pagination.TotalPages)
	assert.True(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)

	// Test cache hit
	req, _ = http.NewRequest("GET", "/tree?pageSize=5", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var cachedResponse cache.PaginatedTreeResponse
	err = json.Unmarshal(w.Body.Bytes(), &cachedResponse)
	assert.NoError(t, err)

	// Verify cached response matches original
	assert.Equal(t, response.Pagination, cachedResponse.Pagination)
	assert.Equal(t, len(response.Data), len(cachedResponse.Data))
}

func TestGetTreeEmpty(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.GET("/tree", handler.GetTree)

	// Create test request
	req, _ := http.NewRequest("GET", "/tree", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response structure
	var response cache.PaginatedTreeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response.Data)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.PageSize)
	assert.Equal(t, int64(0), response.Pagination.Total)
	assert.Equal(t, int64(0), response.Pagination.TotalPages)
	assert.False(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)
}

func TestCreateNode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create initial root node
	rootID, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.POST("/tree", handler.CreateNode)
	router.GET("/tree", handler.GetTree)

	// Create test request
	payload := models.CreateNodeRequest{
		Label:    "child",
		ParentID: rootID,
	}
	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/tree", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "child", response["label"])
	assert.Equal(t, float64(rootID), response["parentId"])

	// Verify cache was invalidated by checking if a new GET request hits the repository
	req, _ = http.NewRequest("GET", "/tree", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var treeResponse cache.PaginatedTreeResponse
	err = json.Unmarshal(w.Body.Bytes(), &treeResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(treeResponse.Data)) // root + child but in one tree
}

func TestCreateNodeInvalidInput(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.POST("/tree", handler.CreateNode)

	// Test cases
	testCases := []struct {
		name     string
		payload  models.CreateNodeRequest
		expected int
	}{
		{
			name:     "Empty label",
			payload:  models.CreateNodeRequest{Label: ""},
			expected: http.StatusBadRequest,
		},
		{
			name:     "Label too long",
			payload:  models.CreateNodeRequest{Label: string(make([]byte, 101))},
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid parent ID",
			payload:  models.CreateNodeRequest{Label: "test", ParentID: -1},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tc.payload)
			req, _ := http.NewRequest("POST", "/tree", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

func TestCreateNodeNonExistentParent(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.POST("/tree", handler.CreateNode)

	// Create test request with non-existent parent
	payload := models.CreateNodeRequest{
		Label:    "child",
		ParentID: 999, // Non-existent parent ID
	}
	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/tree", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateNodeDeepNesting(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.POST("/tree", handler.CreateNode)

	// Create root node
	rootID, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

	// Create a chain of nodes
	lastID := rootID
	for i := 0; i < 10; i++ {
		payload := models.CreateNodeRequest{
			Label:    fmt.Sprintf("level_%d", i+1),
			ParentID: lastID,
		}
		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/tree", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		lastID = int64(response["id"].(float64))
	}

	// Verify the tree structure
	nodes, total, err := repo.GetAllNodes(context.Background(), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), total) // Root + 10 levels
	assert.Len(t, nodes, 11)
}

func TestUpdateNode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create initial root node
	nodeID, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.PUT("/node/:id", handler.UpdateNode)

	// Create test request payload
	payload := models.UpdateNodeRequest{
		Label: "updated_root",
	}
	jsonPayload, _ := json.Marshal(payload)

	// Create test request
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/node/%d", nodeID), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the update
	updatedNode, err := repo.GetNode(context.Background(), nodeID)
	assert.NoError(t, err)
	assert.Equal(t, "updated_root", updatedNode.Label)
}

func TestGetTreePagination(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create multiple root nodes
	for i := 0; i < 15; i++ {
		_, err := repo.CreateNode(context.Background(), fmt.Sprintf("root_%d", i+1), nil)
		assert.NoError(t, err)
	}

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.GET("/tree", handler.GetTree)

	// Test cases
	testCases := []struct {
		name           string
		query          string
		expectedCount  int
		expectedTotal  int64
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Default pagination",
			query:          "",
			expectedCount:  10,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Custom page size",
			query:          "?pageSize=5",
			expectedCount:  5,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Maximum page size",
			query:          "?pageSize=100",
			expectedCount:  15,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Exceeds maximum page size",
			query:          "?pageSize=101",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "page size cannot exceed 100",
		},
		{
			name:           "Invalid page size",
			query:          "?pageSize=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid page size",
		},
		{
			name:           "Zero page size",
			query:          "?pageSize=0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "page size must be greater than 0",
		},
		{
			name:           "Negative page size",
			query:          "?pageSize=-1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "page size must be greater than 0",
		},
		{
			name:           "Second page",
			query:          "?page=2&pageSize=5",
			expectedCount:  5,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Last page",
			query:          "?page=3&pageSize=5",
			expectedCount:  5,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid page",
			query:          "?page=0",
			expectedCount:  10,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/tree"+tc.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response cache.PaginatedTreeResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check data
				assert.Len(t, response.Data, tc.expectedCount)

				// Check pagination
				assert.Equal(t, tc.expectedTotal, response.Pagination.Total)
				assert.Equal(t, tc.expectedCount, len(response.Data))
			} else {
				var errorResponse map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedError, errorResponse["error"])
			}
		})
	}
}

func TestMultipleTrees(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize test dependencies
	repo, cleanup := setupTest(t)
	defer cleanup()

	// Create handler
	handler := handlers.NewTreeHandler(repo)

	// Set up routes
	router.POST("/tree", handler.CreateNode)
	router.GET("/tree", handler.GetTree)
	router.PUT("/tree/:id", handler.UpdateNode)

	// Create three independent trees
	treeStructures := []struct {
		rootLabel string
		children  []string
	}{
		{
			rootLabel: "Tree1",
			children:  []string{"Child1.1", "Child1.2", "Child1.3"},
		},
		{
			rootLabel: "Tree2",
			children:  []string{"Child2.1", "Child2.2"},
		},
		{
			rootLabel: "Tree3",
			children:  []string{"Child3.1", "Child3.2", "Child3.3", "Child3.4"},
		},
	}

	rootIDs := make([]int64, len(treeStructures))

	// Create root nodes first
	for i, tree := range treeStructures {
		// Create root node
		payload := models.CreateNodeRequest{
			Label: tree.rootLabel,
		}
		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/tree", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		rootIDs[i] = int64(response["id"].(float64))

		// Create children for this tree
		for _, childLabel := range tree.children {
			payload := models.CreateNodeRequest{
				Label:    childLabel,
				ParentID: rootIDs[i],
			}
			jsonPayload, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", "/tree", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code)
		}
	}

	// Test getting all trees
	req, _ := http.NewRequest("GET", "/tree", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response cache.PaginatedTreeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check data
	assert.NotNil(t, response.Data, "Response data should not be nil")
	assert.Len(t, response.Data, 3) // Should have 3 root nodes

	// Verify each tree's structure
	for i, rootNode := range response.Data {
		assert.NotNil(t, rootNode, "Root node should not be nil")
		assert.Equal(t, treeStructures[i].rootLabel, rootNode.Label)

		// Get children count
		assert.NotNil(t, rootNode.Children, "Children slice should not be nil")
		assert.Len(t, rootNode.Children, len(treeStructures[i].children))
	}

	// Test pagination with multiple trees
	testCases := []struct {
		name           string
		query          string
		expectedCount  int
		expectedTotal  int64
		expectedStatus int
		expectedLabels []string
	}{
		{
			name:           "First page with 2 items",
			query:          "?pageSize=2",
			expectedCount:  2,  // Tree1 and Tree2
			expectedTotal:  12, // Total number of nodes (3 root nodes + 9 children)
			expectedStatus: http.StatusOK,
			expectedLabels: []string{"Tree1", "Tree2"},
		},
		{
			name:           "Second page with 2 items",
			query:          "?page=2&pageSize=2",
			expectedCount:  1,  // Tree3
			expectedTotal:  12, // Total number of nodes (3 root nodes + 9 children)
			expectedStatus: http.StatusOK,
			expectedLabels: []string{"Tree3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/tree"+tc.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var pageResponse cache.PaginatedTreeResponse
			err := json.Unmarshal(w.Body.Bytes(), &pageResponse)
			assert.NoError(t, err)

			assert.NotNil(t, pageResponse.Data, "Response data should not be nil")
			assert.Len(t, pageResponse.Data, tc.expectedCount)

			// Verify we got the expected root nodes
			for i, expectedLabel := range tc.expectedLabels {
				if i < len(pageResponse.Data) { // Only check if we have enough data
					assert.Equal(t, expectedLabel, pageResponse.Data[i].Label,
						fmt.Sprintf("Expected root node %d to have label %s", i, expectedLabel))
				}
			}

			assert.Equal(t, tc.expectedTotal, pageResponse.Pagination.Total)
			assert.Equal(t, tc.expectedCount, len(pageResponse.Data))
		})
	}

	// Test updating nodes in different trees
	for i, rootID := range rootIDs {
		// Update root node
		updatePayload := models.UpdateNodeRequest{
			Label: fmt.Sprintf("Updated%s", treeStructures[i].rootLabel),
		}
		jsonPayload, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/tree/%d", rootID), bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update
		updatedNode, err := repo.GetNode(context.Background(), rootID)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Updated%s", treeStructures[i].rootLabel), updatedNode.Label)
	}

	// Test deleting one tree
	err = repo.DeleteNode(context.Background(), rootIDs[0])
	assert.NoError(t, err)

	// Verify remaining trees
	nodes, total, err := repo.GetAllNodes(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), total) // 2 root nodes + 6 children (2 from Tree2 + 4 from Tree3)
	assert.Len(t, nodes, 8)          // Verify we have all remaining nodes

	// Verify Tree1 and its children are deleted
	for _, node := range nodes {
		assert.NotEqual(t, "Tree1", node.Label)
		assert.NotEqual(t, "Child1.1", node.Label)
		assert.NotEqual(t, "Child1.2", node.Label)
		assert.NotEqual(t, "Child1.3", node.Label)
	}

	// Verify remaining trees are intact
	remainingTrees := 0
	for _, node := range nodes {
		if node.ParentID == nil {
			remainingTrees++
		}
	}
	assert.Equal(t, 2, remainingTrees) // Should have 2 root nodes (Tree2 and Tree3)
}
