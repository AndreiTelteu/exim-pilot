package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/api"
	"github.com/andreitelteu/exim-pilot/internal/auth"
	"github.com/andreitelteu/exim-pilot/internal/config"
	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/queue"
	"github.com/andreitelteu/exim-pilot/web"
)

func main() {
	var (
		configPath  = flag.String("config", getDefaultConfigPath(), "Path to configuration file")
		migrateUp   = flag.Bool("migrate-up", false, "Run database migrations up")
		migrateDown = flag.Bool("migrate-down", false, "Run database migrations down")
		versionFlag = flag.Bool("version", false, "Show version information")
		helpFlag    = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *versionFlag {
		showVersion()
		return
	}

	fmt.Println("Exim Control Panel starting...")

	// Load configuration
	cfg, err := config.LoadFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize embedded assets
	web.InitEmbeddedAssets()

	// Create database config from main config
	dbConfig := &database.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.GetDatabaseConnMaxLifetime(),
	}

	// Create database directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Connect to database
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Handle migration commands
	if *migrateUp {
		if err := database.MigrateUp(db); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		fmt.Println("Database migrations completed successfully")
		return
	}

	if *migrateDown {
		if err := database.MigrateDown(db); err != nil {
			log.Fatalf("Failed to rollback database migration: %v", err)
		}
		fmt.Println("Database migration rollback completed successfully")
		return
	}

	// Run database migrations automatically in normal startup
	if err := database.MigrateUp(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize repository
	repository := database.NewRepository(db)

	// Initialize queue service
	queueService := queue.NewService(cfg.Exim.BinaryPath, db)

	// Initialize log processing service
	logConfig := logprocessor.DefaultServiceConfig()
	logService := logprocessor.NewService(repository, logConfig)

	// Start log service
	if err := logService.Start(); err != nil {
		log.Fatalf("Failed to start log service: %v", err)
	}
	defer logService.Stop()

	// Initialize default admin user if no users exist
	if err := initializeDefaultUser(db, cfg); err != nil {
		log.Printf("Warning: Failed to initialize default user: %v", err)
	}

	// Create API config from main config
	apiConfig := &api.Config{
		Port:           cfg.Server.Port,
		Host:           cfg.Server.Host,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		AllowedOrigins: cfg.Server.AllowedOrigins,
		LogRequests:    cfg.Server.LogRequests,
	}

	// Initialize API server
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

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	if configPath := os.Getenv("EXIM_PILOT_CONFIG"); configPath != "" {
		return configPath
	}
	return "/opt/exim-pilot/config/config.yaml"
}

// showHelp displays help information
func showHelp() {
	fmt.Println("Exim Control Panel")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  exim-pilot [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config string")
	fmt.Println("        Path to configuration file (default: /opt/exim-pilot/config/config.yaml)")
	fmt.Println("  -migrate-up")
	fmt.Println("        Run database migrations up and exit")
	fmt.Println("  -migrate-down")
	fmt.Println("        Run database migration rollback and exit")
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  EXIM_PILOT_CONFIG    Configuration file path")
	fmt.Println("  EXIM_PILOT_*         Configuration overrides")
}

// showVersion displays version information
func showVersion() {
	fmt.Println("Exim Control Panel (Exim-Pilot)")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Build: development")
	fmt.Println("Go version:", runtime.Version())
}

// initializeDefaultUser creates a default admin user if no users exist
func initializeDefaultUser(db *database.DB, cfg *config.Config) error {
	authService := auth.NewService(db)
	userRepo := database.NewUserRepository(db)

	// Check if any users exist
	_, err := userRepo.GetByUsername(cfg.Auth.DefaultUsername)
	if err == nil {
		// User already exists
		return nil
	}

	// Use password from config or environment
	password := cfg.Auth.DefaultPassword
	if envPassword := os.Getenv("EXIM_PILOT_ADMIN_PASSWORD"); envPassword != "" {
		password = envPassword
	}

	if password == "" {
		password = "admin123" // Fallback default
		log.Println("Warning: Using fallback default password 'admin123' for admin user. Please change it after first login.")
	}

	_, err = authService.CreateUser(cfg.Auth.DefaultUsername, password, "admin@localhost", "Administrator")
	if err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	log.Printf("Created default admin user with username '%s'", cfg.Auth.DefaultUsername)
	if password == "admin123" {
		log.Println("SECURITY WARNING: Please change the default password after first login!")
	}

	return nil
}
