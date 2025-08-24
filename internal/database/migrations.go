package database

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version   int       `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
	Success   bool      `json:"success"`
}

// GetMigrations returns all available migrations in order
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Initial schema creation",
			Up: `
-- Create schema_migrations table to track migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT TRUE
);

-- Messages table for tracking all messages
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    sender TEXT NOT NULL,
    size INTEGER,
    status TEXT NOT NULL, -- received, delivered, deferred, bounced
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Recipients table for message recipients
CREATE TABLE IF NOT EXISTS recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    recipient TEXT NOT NULL,
    status TEXT NOT NULL, -- delivered, deferred, bounced
    delivered_at DATETIME,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);

-- Delivery attempts table
CREATE TABLE IF NOT EXISTS delivery_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    recipient TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    host TEXT,
    ip_address TEXT,
    status TEXT NOT NULL, -- success, defer, bounce
    smtp_code TEXT,
    error_message TEXT,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);

-- Log entries table for searchable log history
CREATE TABLE IF NOT EXISTS log_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    message_id TEXT,
    log_type TEXT NOT NULL, -- main, reject, panic
    event TEXT NOT NULL,
    host TEXT,
    sender TEXT,
    recipients TEXT, -- JSON array
    size INTEGER,
    status TEXT,
    error_code TEXT,
    error_text TEXT,
    raw_line TEXT NOT NULL
);

-- Audit log for administrative actions
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    action TEXT NOT NULL,
    message_id TEXT,
    user_id TEXT,
    details TEXT, -- JSON
    ip_address TEXT
);

-- Queue snapshots for historical tracking
CREATE TABLE IF NOT EXISTS queue_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    total_messages INTEGER,
    deferred_messages INTEGER,
    frozen_messages INTEGER,
    oldest_message_age INTEGER -- seconds
);

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT,
    full_name TEXT,
    role TEXT DEFAULT 'user',
    active BOOLEAN DEFAULT TRUE,
    last_login DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for session management
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`,
			Down: `
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS queue_snapshots;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS log_entries;
DROP TABLE IF EXISTS delivery_attempts;
DROP TABLE IF EXISTS recipients;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS schema_migrations;
`,
		},
		{
			Version:     2,
			Description: "Add indexes for performance",
			Up: `
-- Indexes for messages table
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);

-- Indexes for recipients table
CREATE INDEX IF NOT EXISTS idx_recipients_message_id ON recipients(message_id);
CREATE INDEX IF NOT EXISTS idx_recipients_status ON recipients(status);
CREATE INDEX IF NOT EXISTS idx_recipients_recipient ON recipients(recipient);

-- Indexes for delivery_attempts table
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_message_id ON delivery_attempts(message_id);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_timestamp ON delivery_attempts(timestamp);
CREATE INDEX IF NOT EXISTS idx_delivery_attempts_status ON delivery_attempts(status);

-- Indexes for log_entries table
CREATE INDEX IF NOT EXISTS idx_log_entries_timestamp ON log_entries(timestamp);
CREATE INDEX IF NOT EXISTS idx_log_entries_message_id ON log_entries(message_id);
CREATE INDEX IF NOT EXISTS idx_log_entries_event ON log_entries(event);
CREATE INDEX IF NOT EXISTS idx_log_entries_log_type ON log_entries(log_type);
CREATE INDEX IF NOT EXISTS idx_log_entries_sender ON log_entries(sender);

-- Indexes for audit_log table
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

-- Indexes for queue_snapshots table
CREATE INDEX IF NOT EXISTS idx_queue_snapshots_timestamp ON queue_snapshots(timestamp);

-- Indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);

-- Indexes for sessions table
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
`,
			Down: `
-- Drop all indexes
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_queue_snapshots_timestamp;
DROP INDEX IF EXISTS idx_audit_log_action;
DROP INDEX IF EXISTS idx_audit_log_user_id;
DROP INDEX IF EXISTS idx_audit_log_timestamp;
DROP INDEX IF EXISTS idx_log_entries_sender;
DROP INDEX IF EXISTS idx_log_entries_log_type;
DROP INDEX IF EXISTS idx_log_entries_event;
DROP INDEX IF EXISTS idx_log_entries_message_id;
DROP INDEX IF EXISTS idx_log_entries_timestamp;
DROP INDEX IF EXISTS idx_delivery_attempts_status;
DROP INDEX IF EXISTS idx_delivery_attempts_timestamp;
DROP INDEX IF EXISTS idx_delivery_attempts_message_id;
DROP INDEX IF EXISTS idx_recipients_recipient;
DROP INDEX IF EXISTS idx_recipients_status;
DROP INDEX IF EXISTS idx_recipients_message_id;
DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_messages_status;
DROP INDEX IF EXISTS idx_messages_timestamp;
`,
		},
		{
			Version:     3,
			Description: "Add message notes and tags for troubleshooting",
			Up: `
-- Message notes table for operator notes
CREATE TABLE IF NOT EXISTS message_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    note TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Message tags table for categorization
CREATE TABLE IF NOT EXISTS message_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Indexes for new tables
CREATE INDEX IF NOT EXISTS idx_message_notes_message_id ON message_notes(message_id);
CREATE INDEX IF NOT EXISTS idx_message_notes_user_id ON message_notes(user_id);
CREATE INDEX IF NOT EXISTS idx_message_notes_created_at ON message_notes(created_at);

CREATE INDEX IF NOT EXISTS idx_message_tags_message_id ON message_tags(message_id);
CREATE INDEX IF NOT EXISTS idx_message_tags_tag ON message_tags(tag);
CREATE INDEX IF NOT EXISTS idx_message_tags_user_id ON message_tags(user_id);
`,
			Down: `
DROP INDEX IF EXISTS idx_message_tags_user_id;
DROP INDEX IF EXISTS idx_message_tags_tag;
DROP INDEX IF EXISTS idx_message_tags_message_id;
DROP INDEX IF EXISTS idx_message_notes_created_at;
DROP INDEX IF EXISTS idx_message_notes_user_id;
DROP INDEX IF EXISTS idx_message_notes_message_id;
DROP TABLE IF EXISTS message_tags;
DROP TABLE IF EXISTS message_notes;
`,
		},
		{
			Version:     4,
			Description: "Add configuration and system status tables",
			Up: `
-- System configuration table
CREATE TABLE IF NOT EXISTS system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER,
    FOREIGN KEY (updated_by) REFERENCES users(id)
);

-- System status table for health monitoring
CREATE TABLE IF NOT EXISTS system_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    component TEXT NOT NULL, -- log_processor, queue_monitor, etc.
    status TEXT NOT NULL,    -- healthy, warning, error
    message TEXT,
    details TEXT,            -- JSON
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Login attempts table for security monitoring
CREATE TABLE IF NOT EXISTS login_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    user_agent TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for new tables
CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config(key);
CREATE INDEX IF NOT EXISTS idx_system_status_component ON system_status(component);
CREATE INDEX IF NOT EXISTS idx_system_status_timestamp ON system_status(timestamp);
CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_address ON login_attempts(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_attempts_timestamp ON login_attempts(timestamp);
`,
			Down: `
DROP INDEX IF EXISTS idx_login_attempts_timestamp;
DROP INDEX IF EXISTS idx_login_attempts_ip_address;
DROP INDEX IF EXISTS idx_login_attempts_username;
DROP INDEX IF EXISTS idx_system_status_timestamp;
DROP INDEX IF EXISTS idx_system_status_component;
DROP INDEX IF EXISTS idx_system_config_key;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS system_status;
DROP TABLE IF EXISTS system_config;
`,
		},
	}
}

// MigrateUp runs all pending migrations
func MigrateUp(db *DB) error {
	log.Println("Starting database migrations...")

	// Ensure schema_migrations table exists
	if err := createSchemaMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	log.Printf("Current database version: %d", currentVersion)

	// Get all migrations
	migrations := GetMigrations()
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Apply pending migrations
	applied := 0
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Description)

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		applied++
	}

	if applied == 0 {
		log.Println("No migrations to apply - database is up to date")
	} else {
		log.Printf("Successfully applied %d migrations", applied)
	}

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(db *DB) error {
	log.Println("Rolling back last migration...")

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to roll back")
	}

	// Find the migration to roll back
	migrations := GetMigrations()
	var targetMigration *Migration
	for _, migration := range migrations {
		if migration.Version == currentVersion {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found", currentVersion)
	}

	log.Printf("Rolling back migration %d: %s", targetMigration.Version, targetMigration.Description)

	// Start transaction
	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute down migration
	if targetMigration.Down != "" {
		statements := strings.Split(targetMigration.Down, ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if _, err := tx.Exec(stmt); err != nil {
				return fmt.Errorf("failed to execute down migration statement: %w", err)
			}
		}
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = ?", targetMigration.Version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("Successfully rolled back migration %d", targetMigration.Version)
	return nil
}

// MigrateToVersion migrates to a specific version
func MigrateToVersion(db *DB, targetVersion int) error {
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == targetVersion {
		log.Printf("Database is already at version %d", targetVersion)
		return nil
	}

	if targetVersion > currentVersion {
		// Migrate up
		migrations := GetMigrations()
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Version < migrations[j].Version
		})

		for _, migration := range migrations {
			if migration.Version <= currentVersion || migration.Version > targetVersion {
				continue
			}

			log.Printf("Applying migration %d: %s", migration.Version, migration.Description)

			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	} else {
		// Migrate down
		migrations := GetMigrations()
		sort.Slice(migrations, func(i, j int) bool {
			return migrations[i].Version > migrations[j].Version
		})

		for _, migration := range migrations {
			if migration.Version <= targetVersion || migration.Version > currentVersion {
				continue
			}

			log.Printf("Rolling back migration %d: %s", migration.Version, migration.Description)

			// Start transaction
			tx, err := db.BeginTx()
			if err != nil {
				return fmt.Errorf("failed to start transaction: %w", err)
			}

			// Execute down migration
			if migration.Down != "" {
				statements := strings.Split(migration.Down, ";")
				for _, stmt := range statements {
					stmt = strings.TrimSpace(stmt)
					if stmt == "" {
						continue
					}

					if _, err := tx.Exec(stmt); err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to execute down migration statement: %w", err)
					}
				}
			}

			// Remove migration record
			_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = ?", migration.Version)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove migration record: %w", err)
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit rollback: %w", err)
			}
		}
	}

	log.Printf("Successfully migrated to version %d", targetVersion)
	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *DB) ([]MigrationRecord, error) {
	// Ensure schema_migrations table exists
	if err := createSchemaMigrationsTable(db); err != nil {
		return nil, fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	rows, err := db.Query("SELECT version, applied_at, success FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %w", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Version, &record.AppliedAt, &record.Success); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// createSchemaMigrationsTable creates the schema_migrations table if it doesn't exist
func createSchemaMigrationsTable(db *DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		success BOOLEAN DEFAULT TRUE
	)`

	_, err := db.Exec(query)
	return err
}

// getCurrentVersion returns the current database version
func getCurrentVersion(db *DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE success = TRUE").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}
	return version, nil
}

// applyMigration applies a single migration
func applyMigration(db *DB, migration Migration) error {
	// Start transaction
	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration statements
	statements := strings.Split(migration.Up, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.Exec(stmt); err != nil {
			// Record failed migration
			tx.Exec("INSERT INTO schema_migrations (version, success) VALUES (?, FALSE)", migration.Version)
			return fmt.Errorf("failed to execute migration statement: %w", err)
		}
	}

	// Record successful migration
	_, err = tx.Exec("INSERT INTO schema_migrations (version, success) VALUES (?, TRUE)", migration.Version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// ValidateMigrations validates that all migrations are consistent
func ValidateMigrations() error {
	migrations := GetMigrations()

	// Check for duplicate versions
	versions := make(map[int]bool)
	for _, migration := range migrations {
		if versions[migration.Version] {
			return fmt.Errorf("duplicate migration version: %d", migration.Version)
		}
		versions[migration.Version] = true

		// Validate migration content
		if migration.Up == "" {
			return fmt.Errorf("migration %d has empty up script", migration.Version)
		}

		if migration.Description == "" {
			return fmt.Errorf("migration %d has empty description", migration.Version)
		}
	}

	// Check for version gaps
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	for i, migration := range migrations {
		expectedVersion := i + 1
		if migration.Version != expectedVersion {
			return fmt.Errorf("migration version gap: expected %d, got %d", expectedVersion, migration.Version)
		}
	}

	return nil
}

// CreateMigration creates a new migration file template
func CreateMigration(description string) Migration {
	// Get the next version number
	migrations := GetMigrations()
	nextVersion := 1
	for _, migration := range migrations {
		if migration.Version >= nextVersion {
			nextVersion = migration.Version + 1
		}
	}

	return Migration{
		Version:     nextVersion,
		Description: description,
		Up:          "-- Add your up migration here\n",
		Down:        "-- Add your down migration here\n",
	}
}

// GetPendingMigrations returns migrations that haven't been applied yet
func GetPendingMigrations(db *DB) ([]Migration, error) {
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	migrations := GetMigrations()
	var pending []Migration

	for _, migration := range migrations {
		if migration.Version > currentVersion {
			pending = append(pending, migration)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending, nil
}
