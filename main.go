package main

import (
	"context"
	"log"
	"os"

	"github.com/ammiranda/tree_service/cache"
	"github.com/ammiranda/tree_service/config"
	"github.com/ammiranda/tree_service/handlers"
	"github.com/ammiranda/tree_service/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set development environment
	if err := os.Setenv("APP_ENV", "development"); err != nil {
		log.Fatal("Failed to set environment variable:", err)
	}

	// Create context
	ctx := context.Background()

	// Initialize config provider
	cfgProvider := config.NewEnvProvider("")

	// Initialize repository
	repo, err := repository.NewPostgresRepository(cfgProvider)
	if err != nil {
		log.Fatal("Failed to create repository:", err)
	}
	if err := repo.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize repository:", err)
	}
	defer func() {
		if err := repo.Cleanup(ctx); err != nil {
			log.Printf("Warning: Failed to cleanup repository: %v", err)
		}
	}()

	// Initialize cache
	if err := cache.Initialize(); err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}

	// Initialize handlers
	treeHandler := handlers.NewTreeHandler(repo)

	// Initialize router
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		api.GET("/tree", treeHandler.GetTree)
		api.POST("/tree", treeHandler.CreateNode)
	}

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
