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
	_, err := repo.CreateNode(context.Background(), "root", nil)
	assert.NoError(t, err)

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

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check data
	data := response["data"].([]interface{})
	assert.Len(t, data, 1)
	assert.Equal(t, "root", data[0].(map[string]interface{})["label"])

	// Check pagination
	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["total"])
	assert.Equal(t, float64(10), pagination["pageSize"])
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
	assert.Equal(t, http.StatusNotFound, w.Code)
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
	router.POST("/node", handler.CreateNode)

	// Create test request
	payload := models.CreateNodeRequest{
		Label:    "child",
		ParentID: rootID,
	}
	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/node", bytes.NewBuffer(jsonPayload))
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
		{
			name:           "Invalid page size",
			query:          "?pageSize=200",
			expectedCount:  10,
			expectedTotal:  15,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test request
			req, _ := http.NewRequest("GET", "/tree"+tc.query, nil)
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check data
				data := response["data"].([]interface{})
				assert.Len(t, data, tc.expectedCount)

				// Check pagination
				pagination := response["pagination"].(map[string]interface{})
				assert.Equal(t, float64(tc.expectedTotal), pagination["total"])
				assert.Equal(t, float64(tc.expectedCount), pagination["pageSize"])
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

	// Create root nodes
	for i, tree := range treeStructures {
		// Create root node
		rootID, err := repo.CreateNode(context.Background(), tree.rootLabel, nil)
		assert.NoError(t, err)
		rootIDs[i] = rootID

		// Create children for this tree
		for _, childLabel := range tree.children {
			payload := models.CreateNodeRequest{
				Label:    childLabel,
				ParentID: rootID,
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

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check data
	data := response["data"].([]interface{})
	assert.Len(t, data, 3) // Should have 3 root nodes

	// Verify each tree's structure
	for i, rootNode := range data {
		nodeMap := rootNode.(map[string]interface{})
		assert.Equal(t, treeStructures[i].rootLabel, nodeMap["label"])

		// Get children count
		children := nodeMap["children"].([]interface{})
		assert.Len(t, children, len(treeStructures[i].children))
	}

	// Test pagination with multiple trees
	testCases := []struct {
		name           string
		query          string
		expectedCount  int
		expectedTotal  int64
		expectedStatus int
	}{
		{
			name:           "First page with 2 items",
			query:          "?pageSize=2",
			expectedCount:  2,
			expectedTotal:  3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Second page with 2 items",
			query:          "?page=2&pageSize=2",
			expectedCount:  1,
			expectedTotal:  3,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/tree"+tc.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			data := response["data"].([]interface{})
			assert.Len(t, data, tc.expectedCount)

			pagination := response["pagination"].(map[string]interface{})
			assert.Equal(t, float64(tc.expectedTotal), pagination["total"])
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
