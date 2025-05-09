package main

import (
	"log"

	"theary_test/cache"
	"theary_test/handlers"
	"theary_test/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize repository
	repo := repository.NewSQLiteRepository()
	if err := repo.Initialize(nil); err != nil {
		log.Fatal("Failed to initialize repository:", err)
	}
	defer repo.Cleanup(nil)

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
