package repository

import (
	"context"
	"sort"
	"sync"
)

// MockRepository implements Repository interface for testing
type MockRepository struct {
	nodes map[int64]*Node
	mu    sync.RWMutex
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		nodes: make(map[int64]*Node),
	}
}

// Initialize performs any necessary setup
func (m *MockRepository) Initialize(ctx context.Context) error {
	return nil
}

// Cleanup performs any necessary cleanup
func (m *MockRepository) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes = make(map[int64]*Node)
	return nil
}

// CreateNode creates a new node
func (m *MockRepository) CreateNode(ctx context.Context, label string, parentID *int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a new ID
	id := int64(len(m.nodes) + 1)

	// Create the node
	node := &Node{
		ID:       id,
		Label:    label,
		ParentID: parentID,
	}

	// Store the node
	m.nodes[id] = node

	return id, nil
}

// GetNode retrieves a node by ID
func (m *MockRepository) GetNode(ctx context.Context, id int64) (*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[id]
	if !ok {
		return nil, ErrNodeNotFound
	}

	return node, nil
}

// GetAllNodes retrieves all nodes with pagination
func (m *MockRepository) GetAllNodes(ctx context.Context, page, pageSize int) ([]*Node, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// First, identify and sort root nodes
	var rootNodes []*Node
	for _, node := range m.nodes {
		if node.ParentID == nil {
			rootNodes = append(rootNodes, node)
		}
	}
	sort.Slice(rootNodes, func(i, j int) bool {
		return rootNodes[i].ID < rootNodes[j].ID
	})

	// If no pagination is needed (pageSize >= total nodes), return all nodes
	if pageSize >= len(m.nodes) {
		result := make([]*Node, 0, len(m.nodes))
		for _, node := range m.nodes {
			nodeCopy := &Node{
				ID:       node.ID,
				Label:    node.Label,
				ParentID: node.ParentID,
			}
			result = append(result, nodeCopy)
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
		return result, int64(len(m.nodes)), nil
	}

	// Calculate pagination for root nodes
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if end > len(rootNodes) {
		end = len(rootNodes)
	}

	// Get the paginated root nodes
	var paginatedRoots []*Node
	if offset < len(rootNodes) {
		paginatedRoots = rootNodes[offset:end]
	}

	// Build result set: first add roots, then all their children
	result := make([]*Node, 0)

	// Add root nodes first
	for _, root := range paginatedRoots {
		// Create a copy of the root node without children
		rootCopy := &Node{
			ID:       root.ID,
			Label:    root.Label,
			ParentID: root.ParentID,
		}
		result = append(result, rootCopy)

		// Find all children of this root node
		for _, node := range m.nodes {
			if node.ParentID != nil && *node.ParentID == root.ID {
				// Create a copy of the child node without children
				childCopy := &Node{
					ID:       node.ID,
					Label:    node.Label,
					ParentID: node.ParentID,
				}
				result = append(result, childCopy)
			}
		}
	}

	// If we have no results but there are nodes in the repository,
	// it means we need to include nodes whose parents are not in the current page
	if len(result) == 0 && len(m.nodes) > 0 {
		// Find all nodes that should be in this page
		for _, node := range m.nodes {
			// Skip nodes that are already included
			alreadyIncluded := false
			for _, includedNode := range result {
				if includedNode.ID == node.ID {
					alreadyIncluded = true
					break
				}
			}
			if !alreadyIncluded {
				// If this node has a parent, make sure the parent is included
				if node.ParentID != nil {
					parent, exists := m.nodes[*node.ParentID]
					if exists {
						// Add parent first
						parentCopy := &Node{
							ID:       parent.ID,
							Label:    parent.Label,
							ParentID: parent.ParentID,
						}
						result = append(result, parentCopy)
					}
				}
				// Add the node
				nodeCopy := &Node{
					ID:       node.ID,
					Label:    node.Label,
					ParentID: node.ParentID,
				}
				result = append(result, nodeCopy)
			}
		}
		// Sort by ID
		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
	}

	// If we have a root node with children, make sure we include all children
	if len(result) > 0 && result[0].ParentID == nil {
		// Find all children of the root node
		for _, node := range m.nodes {
			if node.ParentID != nil && *node.ParentID == result[0].ID {
				// Skip nodes that are already included
				alreadyIncluded := false
				for _, includedNode := range result {
					if includedNode.ID == node.ID {
						alreadyIncluded = true
						break
					}
				}
				if !alreadyIncluded {
					nodeCopy := &Node{
						ID:       node.ID,
						Label:    node.Label,
						ParentID: node.ParentID,
					}
					result = append(result, nodeCopy)
				}
			}
		}
		// Sort by ID
		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
	}

	return result, int64(len(m.nodes)), nil
}

// UpdateNode updates a node
func (m *MockRepository) UpdateNode(ctx context.Context, id int64, label string, parentID *int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	if !ok {
		return ErrNodeNotFound
	}

	node.Label = label
	node.ParentID = parentID

	return nil
}

// DeleteNode deletes a node and its children
func (m *MockRepository) DeleteNode(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First, find and delete all child nodes
	toDelete := []int64{id}
	deleted := make(map[int64]bool)

	for len(toDelete) > 0 {
		currentID := toDelete[0]
		toDelete = toDelete[1:]

		if deleted[currentID] {
			continue
		}

		// Find all children of the current node
		for nodeID, node := range m.nodes {
			if node.ParentID != nil && *node.ParentID == currentID {
				toDelete = append(toDelete, nodeID)
			}
		}

		// Delete the current node
		delete(m.nodes, currentID)
		deleted[currentID] = true
	}

	return nil
}
