package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ammiranda/tree_service/cache"
	"github.com/ammiranda/tree_service/models"
	"github.com/ammiranda/tree_service/repository"

	"github.com/gin-gonic/gin"
)

var (
	ErrTreeNotFound = errors.New("tree not found")
)

// TreeHandler handles tree-related HTTP requests
type TreeHandler struct {
	repo repository.Repository
}

// NewTreeHandler creates a new TreeHandler instance
func NewTreeHandler(repo repository.Repository) *TreeHandler {
	return &TreeHandler{
		repo: repo,
	}
}

// BuildTreeFromNodes builds the tree structure from a list of nodes
func BuildTreeFromNodes(nodes []*repository.Node) ([]*models.Node, error) {
	if len(nodes) == 0 {
		return nil, ErrTreeNotFound
	}

	// Create a map to store all nodes
	nodeMap := make(map[int64]*models.Node)
	var rootNodes []*models.Node

	// First pass: create all nodes
	for _, node := range nodes {
		modelNode := models.NewNode(node.Label)
		modelNode.ID = node.ID
		nodeMap[node.ID] = modelNode

		// If it's a root node (no parent), add it to rootNodes
		if node.ParentID == nil {
			rootNodes = append(rootNodes, modelNode)
		}
	}

	// Check if we found any root nodes
	if len(rootNodes) == 0 {
		return nil, ErrTreeNotFound
	}

	// Second pass: build the tree structure
	for _, node := range nodes {
		if node.ParentID != nil {
			if parent, exists := nodeMap[*node.ParentID]; exists {
				if child, exists := nodeMap[node.ID]; exists {
					parent.AddChild(child)
				}
			}
		}
	}

	return rootNodes, nil
}

// GetTree returns all trees in the database with pagination
func (h *TreeHandler) GetTree(c *gin.Context) {
	// Get pagination parameters
	page := 1
	pageSize := 10

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse pageSize parameter
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Try to get from cache first
	if cachedTree, found := cache.GetTree(); found {
		if len(cachedTree) == 0 {
			c.JSON(http.StatusNotFound, map[string]string{"error": "tree not found"})
			return
		}
		c.JSON(http.StatusOK, cachedTree)
		return
	}

	// If not in cache, get from repository
	ctx := c.Request.Context()
	nodes, total, err := h.repo.GetAllNodes(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Build tree structure
	rootNodes, err := BuildTreeFromNodes(nodes)
	if err != nil {
		if errors.Is(err, ErrTreeNotFound) {
			c.JSON(http.StatusNotFound, map[string]string{"error": "tree not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Store in cache
	cache.SetTree(rootNodes)

	// Calculate pagination metadata
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"data": rootNodes,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
	})
}

// CreateNode creates a new node in the tree
func (h *TreeHandler) CreateNode(c *gin.Context) {
	var req models.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var parentID *int64
	if req.ParentID > 0 {
		parentID = &req.ParentID
		// Check if parent exists
		_, err := h.repo.GetNode(ctx, *parentID)
		if err != nil {
			if errors.Is(err, repository.ErrNodeNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "parent node not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Create node using repository
	id, err := h.repo.CreateNode(ctx, req.Label, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache since we modified the tree
	cache.InvalidateCache()

	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"label":    req.Label,
		"parentId": parentID,
	})
}

// UpdateNode updates an existing node in the tree
func (h *TreeHandler) UpdateNode(c *gin.Context) {
	var req models.UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get node ID from path
	nodeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node ID"})
		return
	}

	// Update node using repository
	err = h.repo.UpdateNode(c.Request.Context(), nodeID, req.Label, req.ParentID)
	if err != nil {
		if errors.Is(err, repository.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache since we modified the tree
	cache.InvalidateCache()

	c.JSON(http.StatusOK, gin.H{
		"id":       nodeID,
		"label":    req.Label,
		"parentId": req.ParentID,
	})
}
