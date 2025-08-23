package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/api"
	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/queue"
)

func main() {
	fmt.Println("Exim Control Panel starting...")

	// Initialize database
	dbConfig := database.DefaultConfig()
	dbConfig.Path = "data/exim-pilot.db"

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.MigrateUp(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize repository
	repository := database.NewRepository(db)

	// Initialize queue service
	eximPath := os.Getenv("EXIM_PATH")
	if eximPath == "" {
		eximPath = "/usr/sbin/exim4" // Default for Ubuntu/Debian
	}
	queueService := queue.NewService(eximPath, db)

	// Initialize log processing service
	logConfig := logprocessor.DefaultServiceConfig()
	logService := logprocessor.NewService(repository, logConfig)

	// Start log service
	if err := logService.Start(); err != nil {
		log.Fatalf("Failed to start log service: %v", err)
	}
	defer logService.Stop()

	// Initialize API server
	apiConfig := api.NewConfig()
	apiConfig.LoadFromEnv()
	server := api.NewServer(apiConfig, queueService, logService, repository)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
