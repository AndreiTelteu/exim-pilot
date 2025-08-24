package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andreitelteu/exim-pilot/internal/auth"
	"github.com/andreitelteu/exim-pilot/internal/database"
)

func main() {
	fmt.Println("Resetting Exim Control Panel database...")

	// Remove existing database files
	dbFiles := []string{
		"data/exim-pilot.db",
		"data/exim-pilot.db-wal",
		"data/exim-pilot.db-shm",
	}

	for _, file := range dbFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Could not remove %s: %v", file, err)
		} else {
			fmt.Printf("Removed %s\n", file)
		}
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize new database
	dbConfig := database.DefaultConfig()
	dbConfig.Path = "data/exim-pilot.db"

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.MigrateUp(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	fmt.Println("✅ Database migrations completed")

	// Create default admin user
	authService := auth.NewService(db)
	userRepo := database.NewUserRepository(db)

	// Check if admin user already exists
	_, err = userRepo.GetByUsername("admin")
	if err == nil {
		fmt.Println("Admin user already exists")
		return
	}

	// Create default admin user
	defaultPassword := os.Getenv("ADMIN_PASSWORD")
	if defaultPassword == "" {
		defaultPassword = "admin123"
		fmt.Println("Using default password 'admin123' for admin user")
	}

	user, err := authService.CreateUser("admin", defaultPassword, "admin@localhost", "Administrator")
	if err != nil {
		log.Fatalf("Failed to create default admin user: %v", err)
	}

	fmt.Printf("✅ Created default admin user (ID: %d)\n", user.ID)
	fmt.Println("✅ Database reset completed successfully")
	fmt.Println("")
	fmt.Println("You can now start the server with: go run cmd/exim-pilot/main.go")
}
