package main

import (
	"context"
	"log"

	"github.com/ammiranda/tree_service/config"
	"github.com/ammiranda/tree_service/internal/lambda"
	"github.com/ammiranda/tree_service/repository"

	awslambda "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// Initialize configuration
	cfgProvider, err := config.NewAWSConfigProvider()
	if err != nil {
		log.Fatalf("Failed to create config provider: %v", err)
	}

	// Initialize repository
	repo, err := repository.NewPostgresRepository(cfgProvider)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	if err := repo.Initialize(context.Background()); err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create handler with repository
	handler := lambda.NewHandler(repo)

	// Start Lambda
	awslambda.Start(handler.Handle)
}
