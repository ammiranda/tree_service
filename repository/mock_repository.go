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

	// Get total count
	total := int64(len(m.nodes))

	// Calculate offset
	offset := int64((page - 1) * pageSize)

	// Get all nodes
	allNodes := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		allNodes = append(allNodes, node)
	}

	// Sort nodes by ID
	sort.Slice(allNodes, func(i, j int) bool {
		return allNodes[i].ID < allNodes[j].ID
	})

	// Apply pagination
	start := offset
	end := offset + int64(pageSize)
	if start >= total {
		return []*Node{}, total, nil
	}
	if end > total {
		end = total
	}

	return allNodes[start:end], total, nil
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
