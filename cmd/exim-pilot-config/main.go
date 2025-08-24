package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/andreitelteu/exim-pilot/internal/config"
	"github.com/andreitelteu/exim-pilot/internal/database"
)

const (
	defaultConfigPath = "/opt/exim-pilot/config/config.yaml"
)

func main() {
	var (
		configPath = flag.String("config", defaultConfigPath, "Path to configuration file")
		validate   = flag.Bool("validate", false, "Validate configuration file")
		generate   = flag.Bool("generate", false, "Generate default configuration file")
		migrate    = flag.String("migrate", "", "Run database migrations (up|down|status)")
		version    = flag.Int("version", 0, "Target migration version (use with migrate)")
		helpFlag   = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	// Handle generate command
	if *generate {
		if err := generateConfig(*configPath); err != nil {
			log.Fatalf("Failed to generate configuration: %v", err)
		}
		return
	}

	// Handle validate command
	if *validate {
		if err := validateConfig(*configPath); err != nil {
			log.Fatalf("Configuration validation failed: %v", err)
		}
		fmt.Println("Configuration is valid")
		return
	}

	// Handle migrate command
	if *migrate != "" {
		if err := handleMigration(*configPath, *migrate, *version); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		return
	}

	// If no specific command, show help
	showHelp()
}

func showHelp() {
	fmt.Println("Exim Control Panel Configuration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  exim-pilot-config [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config string")
	fmt.Printf("        Path to configuration file (default %q)\n", defaultConfigPath)
	fmt.Println("  -validate")
	fmt.Println("        Validate configuration file")
	fmt.Println("  -generate")
	fmt.Println("        Generate default configuration file")
	fmt.Println("  -migrate string")
	fmt.Println("        Run database migrations (up|down|status)")
	fmt.Println("  -version int")
	fmt.Println("        Target migration version (use with -migrate)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate default configuration")
	fmt.Println("  exim-pilot-config -generate -config /opt/exim-pilot/config/config.yaml")
	fmt.Println()
	fmt.Println("  # Validate configuration")
	fmt.Println("  exim-pilot-config -validate -config /opt/exim-pilot/config/config.yaml")
	fmt.Println()
	fmt.Println("  # Run database migrations")
	fmt.Println("  exim-pilot-config -migrate up -config /opt/exim-pilot/config/config.yaml")
	fmt.Println()
	fmt.Println("  # Check migration status")
	fmt.Println("  exim-pilot-config -migrate status -config /opt/exim-pilot/config/config.yaml")
	fmt.Println()
	fmt.Println("  # Migrate to specific version")
	fmt.Println("  exim-pilot-config -migrate up -version 2 -config /opt/exim-pilot/config/config.yaml")
}

func generateConfig(configPath string) error {
	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists: %s\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Configuration generation cancelled")
			return nil
		}
	}

	// Generate default configuration
	cfg := config.DefaultConfig()

	// Save to file
	if err := cfg.SaveToFile(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Default configuration generated: %s\n", configPath)
	fmt.Println()
	fmt.Println("IMPORTANT:")
	fmt.Println("- Review and customize the configuration for your environment")
	fmt.Println("- Change the default admin password")
	fmt.Println("- Configure TLS certificates for production use")
	fmt.Println("- Adjust file paths and permissions as needed")

	return nil
}

func validateConfig(configPath string) error {
	// Load configuration
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional checks
	fmt.Println("Configuration validation results:")
	fmt.Println()

	// Check file permissions
	if info, err := os.Stat(configPath); err == nil {
		mode := info.Mode()
		fmt.Printf("✓ Configuration file permissions: %o\n", mode.Perm())
		if mode.Perm() > 0644 {
			fmt.Printf("  WARNING: Configuration file is more permissive than recommended (644)\n")
		}
	}

	// Check database directory
	dbDir := filepath.Dir(cfg.Database.Path)
	if info, err := os.Stat(dbDir); err == nil {
		fmt.Printf("✓ Database directory exists: %s\n", dbDir)
		if !info.IsDir() {
			return fmt.Errorf("database path parent is not a directory: %s", dbDir)
		}
	} else {
		fmt.Printf("! Database directory does not exist: %s (will be created)\n", dbDir)
	}

	// Check log directory
	if cfg.Logging.File != "" {
		logDir := filepath.Dir(cfg.Logging.File)
		if info, err := os.Stat(logDir); err == nil {
			fmt.Printf("✓ Log directory exists: %s\n", logDir)
			if !info.IsDir() {
				return fmt.Errorf("log path parent is not a directory: %s", logDir)
			}
		} else {
			fmt.Printf("! Log directory does not exist: %s (will be created)\n", logDir)
		}
	}

	// Check Exim binary
	if info, err := os.Stat(cfg.Exim.BinaryPath); err == nil {
		fmt.Printf("✓ Exim binary found: %s\n", cfg.Exim.BinaryPath)
		if mode := info.Mode(); mode&0111 == 0 {
			fmt.Printf("  WARNING: Exim binary is not executable\n")
		}
	} else {
		fmt.Printf("! Exim binary not found: %s\n", cfg.Exim.BinaryPath)
	}

	// Check Exim log paths
	for _, logPath := range cfg.Exim.LogPaths {
		if _, err := os.Stat(logPath); err == nil {
			fmt.Printf("✓ Exim log file accessible: %s\n", logPath)
		} else {
			fmt.Printf("! Exim log file not accessible: %s\n", logPath)
		}
	}

	// Check Exim spool directory
	if info, err := os.Stat(cfg.Exim.SpoolDir); err == nil {
		fmt.Printf("✓ Exim spool directory accessible: %s\n", cfg.Exim.SpoolDir)
		if !info.IsDir() {
			return fmt.Errorf("Exim spool path is not a directory: %s", cfg.Exim.SpoolDir)
		}
	} else {
		fmt.Printf("! Exim spool directory not accessible: %s\n", cfg.Exim.SpoolDir)
	}

	// Check TLS configuration
	if cfg.Server.TLSEnabled {
		if _, err := os.Stat(cfg.Server.TLSCertFile); err == nil {
			fmt.Printf("✓ TLS certificate file found: %s\n", cfg.Server.TLSCertFile)
		} else {
			fmt.Printf("! TLS certificate file not found: %s\n", cfg.Server.TLSCertFile)
		}

		if _, err := os.Stat(cfg.Server.TLSKeyFile); err == nil {
			fmt.Printf("✓ TLS key file found: %s\n", cfg.Server.TLSKeyFile)
		} else {
			fmt.Printf("! TLS key file not found: %s\n", cfg.Server.TLSKeyFile)
		}
	}

	fmt.Println()
	fmt.Println("Configuration validation completed")

	return nil
}

func handleMigration(configPath, command string, targetVersion int) error {
	// Load configuration
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database config
	dbConfig := &database.Config{
		Path:            cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.GetDatabaseConnMaxLifetime(),
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connect to database
	db, err := database.Connect(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	switch command {
	case "up":
		if targetVersion > 0 {
			fmt.Printf("Migrating to version %d...\n", targetVersion)
			return database.MigrateToVersion(db, targetVersion)
		} else {
			fmt.Println("Running all pending migrations...")
			return database.MigrateUp(db)
		}

	case "down":
		if targetVersion > 0 {
			fmt.Printf("Rolling back to version %d...\n", targetVersion)
			return database.MigrateToVersion(db, targetVersion)
		} else {
			fmt.Println("Rolling back last migration...")
			return database.MigrateDown(db)
		}

	case "status":
		return showMigrationStatus(db)

	default:
		return fmt.Errorf("unknown migration command: %s (use: up, down, status)", command)
	}
}

func showMigrationStatus(db *database.DB) error {
	// Get migration status
	records, err := database.GetMigrationStatus(db)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Get available migrations
	migrations := database.GetMigrations()

	fmt.Println("Migration Status:")
	fmt.Println("================")
	fmt.Println()

	if len(records) == 0 {
		fmt.Println("No migrations have been applied")
	} else {
		fmt.Printf("Applied migrations (%d):\n", len(records))
		for _, record := range records {
			status := "✓"
			if !record.Success {
				status = "✗"
			}
			fmt.Printf("  %s Version %d - Applied at %s\n",
				status, record.Version, record.AppliedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println()

	// Show pending migrations
	pending, err := database.GetPendingMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		fmt.Println("No pending migrations")
	} else {
		fmt.Printf("Pending migrations (%d):\n", len(pending))
		for _, migration := range pending {
			fmt.Printf("  - Version %d: %s\n", migration.Version, migration.Description)
		}
	}

	fmt.Println()
	fmt.Printf("Total available migrations: %d\n", len(migrations))

	// Validate migrations
	if err := database.ValidateMigrations(); err != nil {
		fmt.Printf("Migration validation error: %v\n", err)
	} else {
		fmt.Println("All migrations are valid")
	}

	return nil
}
