package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"theary_test/cache"
	"theary_test/models"
	"theary_test/repository"

	"github.com/aws/aws-lambda-go/events"
)

// Handler represents the Lambda handler with its dependencies
type Handler struct {
	repo repository.Repository
}

// NewHandler creates a new Handler with the given repository
func NewHandler(repo repository.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// Handle processes API Gateway events
func (h *Handler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Route the request based on HTTP method and path
	switch {
	case request.HTTPMethod == "GET" && request.Path == "/api/tree":
		return h.handleGetTree(ctx, request)
	case request.HTTPMethod == "POST" && request.Path == "/api/tree":
		return h.handleCreateNode(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "Not found"}`,
		}, nil
	}
}

func (h *Handler) handleGetTree(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Try to get from cache first
	if cachedTree, found := cache.GetTree(); found {
		if len(cachedTree) == 0 {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       `{"error": "tree not found"}`,
			}, nil
		}
		body, err := json.Marshal(cachedTree)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       fmt.Sprintf(`{"error": "Failed to marshal response: %v"}`, err),
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(body),
		}, nil
	}

	// If not in cache, build from repository
	nodes, err := h.repo.GetAllNodes(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
		}, nil
	}

	if len(nodes) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "tree not found"}`,
		}, nil
	}

	// Convert repository nodes to model nodes
	modelNodes := make([]*models.Node, len(nodes))
	for i, node := range nodes {
		modelNodes[i] = &models.Node{
			ID:    node.ID,
			Label: node.Label,
		}
	}

	// Build tree structure
	rootNodes := buildTree(modelNodes, nodes)

	// Store in cache
	cache.SetTree(rootNodes)

	body, err := json.Marshal(rootNodes)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "Failed to marshal response: %v"}`, err),
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

func (h *Handler) handleCreateNode(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req models.CreateNodeRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "Invalid request: %v"}`, err),
		}, nil
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
		}, nil
	}

	// Create the node
	var parentID *int64
	if req.ParentID > 0 {
		parentID = &req.ParentID
	}
	id, err := h.repo.CreateNode(ctx, req.Label, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrNodeNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       `{"error": "parent node not found"}`,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
		}, nil
	}

	// Invalidate cache
	cache.InvalidateCache()

	response := map[string]interface{}{
		"id":       id,
		"label":    req.Label,
		"parentId": req.ParentID,
	}
	body, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "Failed to marshal response: %v"}`, err),
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       string(body),
	}, nil
}

// buildTree converts a flat list of nodes into a tree structure
func buildTree(modelNodes []*models.Node, repoNodes []*repository.Node) []*models.Node {
	// Create a map of nodes by ID for quick lookup
	nodeMap := make(map[int64]*models.Node)
	for _, node := range modelNodes {
		nodeMap[node.ID] = node
	}

	// Find root nodes (nodes without parents)
	var rootNodes []*models.Node
	for i, node := range repoNodes {
		if node.ParentID == nil {
			rootNodes = append(rootNodes, modelNodes[i])
		} else {
			// Add this node as a child of its parent
			if parent, ok := nodeMap[*node.ParentID]; ok {
				parent.Children = append(parent.Children, modelNodes[i])
			}
		}
	}

	return rootNodes
}
