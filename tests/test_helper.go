package tests

import (
	"context"
	"testing"

	"github.com/ammiranda/tree_service/repository"
)

var testRepo repository.Repository

// setupTestDB initializes the test database
func setupTestDB(t *testing.T) {
	// Create mock repository
	testRepo = repository.NewMockRepository()
	if err := testRepo.Initialize(context.Background()); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T) {
	if testRepo != nil {
		if err := testRepo.Cleanup(context.Background()); err != nil {
			t.Fatalf("Failed to cleanup test repository: %v", err)
		}
		testRepo = nil
	}
}
