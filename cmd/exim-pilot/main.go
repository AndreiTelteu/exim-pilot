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
	"github.com/andreitelteu/exim-pilot/internal/auth"
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

	// Initialize default admin user if no users exist
	if err := initializeDefaultUser(db); err != nil {
		log.Printf("Warning: Failed to initialize default user: %v", err)
	}

	// Initialize API server
	apiConfig := api.NewConfig()
	apiConfig.LoadFromEnv()
	server := api.NewServer(apiConfig, queueService, logService, repository, db)

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

// initializeDefaultUser creates a default admin user if no users exist
func initializeDefaultUser(db *database.DB) error {
	authService := auth.NewService(db)
	userRepo := database.NewUserRepository(db)

	// Check if any users exist
	_, err := userRepo.GetByUsername("admin")
	if err == nil {
		// User already exists
		return nil
	}

	// Create default admin user
	defaultPassword := os.Getenv("ADMIN_PASSWORD")
	if defaultPassword == "" {
		defaultPassword = "admin123" // Default password - should be changed
		log.Println("Warning: Using default password 'admin123' for admin user. Please change it after first login.")
	}

	_, err = authService.CreateUser("admin", defaultPassword, "admin@localhost", "Administrator")
	if err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	log.Println("Created default admin user with username 'admin'")
	return nil
}
