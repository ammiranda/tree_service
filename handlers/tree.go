package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ammiranda/tree_service/cache"
	"github.com/ammiranda/tree_service/models"
	"github.com/ammiranda/tree_service/repository"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
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
	}

	// Second pass: connect children to parents and identify root/orphaned nodes
	for _, node := range nodes {
		modelNode := nodeMap[node.ID]

		if node.ParentID == nil {
			// This is a root node
			rootNodes = append(rootNodes, modelNode)
		} else if parent, exists := nodeMap[*node.ParentID]; exists {
			// Parent is in the current page, add as child
			parent.AddChild(modelNode)
		} else {
			// Parent is not in the current page, treat as root
			rootNodes = append(rootNodes, modelNode)
		}
	}

	// If we found no nodes to return, consider it not found
	if len(rootNodes) == 0 {
		return nil, ErrTreeNotFound
	}

	return rootNodes, nil
}

// GetTree returns all trees in the database with pagination
func (h *TreeHandler) GetTree(c *gin.Context) {
	// Get pagination parameters
	page := 1
	pageSize := defaultPageSize

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse pageSize parameter
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
			return
		}
		if ps <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page size must be greater than 0"})
			return
		}
		if ps > maxPageSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("page size cannot exceed %d", maxPageSize)})
			return
		}
		pageSize = ps
	}

	// Try to get from cache first
	if cachedResponse, found := cache.GetPaginatedTree(page, pageSize); found {
		c.JSON(http.StatusOK, cachedResponse)
		return
	}

	// If not in cache, get all nodes from repository
	ctx := c.Request.Context()
	allNodes, total, err := h.repo.GetAllNodes(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Create response
	response := &cache.PaginatedTreeResponse{
		Data: make([]*models.Node, 0),
	}
	response.Pagination.Page = page
	response.Pagination.PageSize = pageSize
	response.Pagination.Total = total
	response.Pagination.TotalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	response.Pagination.HasNext = int64(page) < response.Pagination.TotalPages
	response.Pagination.HasPrev = page > 1

	// If we have nodes, build the tree structure
	if len(allNodes) > 0 {
		rootNodes, err := BuildTreeFromNodes(allNodes)
		if err != nil {
			if errors.Is(err, ErrTreeNotFound) {
				// Return empty response instead of 404
				c.JSON(http.StatusOK, response)
				return
			}
			c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		response.Data = rootNodes
	}

	// Store in cache
	cache.SetPaginatedTree(page, pageSize, response)

	// Return response
	c.JSON(http.StatusOK, response)
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
