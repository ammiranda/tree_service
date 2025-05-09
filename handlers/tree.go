package handlers

import (
	"errors"
	"net/http"

	"theary_test/cache"
	"theary_test/models"
	"theary_test/repository"

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

// GetTree returns all trees in the database
func (h *TreeHandler) GetTree(c *gin.Context) {
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
	nodes, err := h.repo.GetAllNodes(ctx)
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

	c.JSON(http.StatusOK, rootNodes)
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
