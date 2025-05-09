package tests

import (
	"context"
	"testing"

	"github.com/ammiranda/tree_service/repository"

	"github.com/stretchr/testify/assert"
)

func TestMockRepository(t *testing.T) {
	// Create mock repository
	repo := repository.NewMockRepository()
	err := repo.Initialize(context.Background())
	assert.NoError(t, err)
	defer repo.Cleanup(context.Background())

	// Test creating a node
	id, err := repo.CreateNode(context.Background(), "test", nil)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Test getting the node
	node, err := repo.GetNode(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, "test", node.Label)
	assert.Nil(t, node.ParentID)

	// Test getting all nodes
	nodes, err := repo.GetAllNodes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, id, nodes[0].ID)
	assert.Equal(t, "test", nodes[0].Label)

	// Test updating the node
	err = repo.UpdateNode(context.Background(), id, "updated", nil)
	assert.NoError(t, err)

	// Verify the update
	node, err = repo.GetNode(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, "updated", node.Label)

	// Test deleting the node
	err = repo.DeleteNode(context.Background(), id)
	assert.NoError(t, err)

	// Verify the deletion
	_, err = repo.GetNode(context.Background(), id)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNodeNotFound, err)
}
