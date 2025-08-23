package database

import (
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db *DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// InitMigrationTable creates the migrations table if it doesn't exist
func (m *MigrationManager) InitMigrationTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns a list of applied migration versions
func (m *MigrationManager) GetAppliedMigrations() ([]int, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// ApplyMigration applies a single migration
func (m *MigrationManager) ApplyMigration(migration Migration) error {
	tx, err := m.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute the migration
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
	}

	// Record the migration
	insertQuery := `
	INSERT INTO schema_migrations (version, description, applied_at) 
	VALUES (?, ?, ?)`

	if _, err := tx.Exec(insertQuery, migration.Version, migration.Description, time.Now()); err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	return nil
}

// RollbackMigration rolls back a single migration
func (m *MigrationManager) RollbackMigration(migration Migration) error {
	tx, err := m.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute the rollback
	if _, err := tx.Exec(migration.Down); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
	}

	// Remove the migration record
	deleteQuery := "DELETE FROM schema_migrations WHERE version = ?"
	if _, err := tx.Exec(deleteQuery, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %w", migration.Version, err)
	}

	return nil
}

// Migrate applies all pending migrations
func (m *MigrationManager) Migrate(migrations []Migration) error {
	if err := m.InitMigrationTable(); err != nil {
		return err
	}

	appliedVersions, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Create a map for quick lookup
	applied := make(map[int]bool)
	for _, version := range appliedVersions {
		applied[version] = true
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Apply pending migrations
	for _, migration := range migrations {
		if !applied[migration.Version] {
			fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Description)
			if err := m.ApplyMigration(migration); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func (m *MigrationManager) GetMigrationStatus(migrations []Migration) (map[int]bool, error) {
	appliedVersions, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	applied := make(map[int]bool)
	for _, version := range appliedVersions {
		applied[version] = true
	}

	status := make(map[int]bool)
	for _, migration := range migrations {
		status[migration.Version] = applied[migration.Version]
	}

	return status, nil
}
