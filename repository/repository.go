package repository

import (
	"context"
	"errors"
)

// Node represents a node in the tree structure
type Node struct {
	ID       int64  // Unique identifier for the node
	Label    string // Display name or content of the node
	ParentID *int64 // Optional reference to the parent node's ID
}

// Repository defines the interface for data access operations.
// It provides methods for managing tree nodes in a persistent storage.
type Repository interface {
	// Initialize performs any necessary setup for the repository.
	// This may include establishing database connections, creating tables,
	// or any other initialization required for the repository to function.
	// Returns an error if initialization fails.
	Initialize(ctx context.Context) error

	// Cleanup performs any necessary cleanup operations for the repository.
	// This may include closing database connections, cleaning up temporary files,
	// or any other cleanup required when the repository is no longer needed.
	// Returns an error if cleanup fails.
	Cleanup(ctx context.Context) error

	// CreateNode creates a new node in the tree structure.
	// Parameters:
	//   - ctx: Context for the operation
	//   - label: The display name or content for the new node
	//   - parentID: Optional reference to the parent node's ID
	// Returns:
	//   - The ID of the newly created node
	//   - An error if the operation fails
	CreateNode(ctx context.Context, label string, parentID *int64) (int64, error)

	// GetNode retrieves a node by its ID.
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the node to retrieve
	// Returns:
	//   - A pointer to the Node if found
	//   - ErrNodeNotFound if no node exists with the given ID
	//   - Other error if the operation fails
	GetNode(ctx context.Context, id int64) (*Node, error)

	// GetAllNodes retrieves all nodes from the repository.
	// Parameters:
	//   - ctx: Context for the operation
	// Returns:
	//   - A slice of all nodes in the repository
	//   - An error if the operation fails
	GetAllNodes(ctx context.Context) ([]*Node, error)

	// UpdateNode updates an existing node's properties.
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the node to update
	//   - label: The new label for the node
	//   - parentID: The new parent ID for the node
	// Returns:
	//   - ErrNodeNotFound if no node exists with the given ID
	//   - Other error if the operation fails
	UpdateNode(ctx context.Context, id int64, label string, parentID *int64) error

	// DeleteNode deletes a node and all its children from the repository.
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the node to delete
	// Returns:
	//   - ErrNodeNotFound if no node exists with the given ID
	//   - Other error if the operation fails
	DeleteNode(ctx context.Context, id int64) error
}

// Common errors
var (
	// ErrNodeNotFound is returned when a requested node does not exist
	ErrNodeNotFound = errors.New("node not found")
	// ErrInvalidInput is returned when the input parameters are invalid
	ErrInvalidInput = errors.New("invalid input")
)
