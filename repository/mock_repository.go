package repository

import (
	"context"
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

// GetAllNodes retrieves all nodes
func (m *MockRepository) GetAllNodes(ctx context.Context) ([]*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}

	return nodes, nil
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

// DeleteNode deletes a node
func (m *MockRepository) DeleteNode(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.nodes[id]; !ok {
		return ErrNodeNotFound
	}

	delete(m.nodes, id)

	return nil
}
