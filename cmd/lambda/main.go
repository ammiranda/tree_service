package main

import (
	"github.com/ammiranda/tree_service/internal/lambda"
	"github.com/ammiranda/tree_service/repository"

	awslambda "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// Initialize repository
	repo := repository.NewMockRepository()
	if err := repo.Initialize(nil); err != nil {
		panic(err)
	}

	// Create handler with repository
	handler := lambda.NewHandler(repo)

	// Start Lambda
	awslambda.Start(handler.Handle)
}
