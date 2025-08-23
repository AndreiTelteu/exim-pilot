package database

import (
	"os"
	"testing"
)

func TestInitMigrationTable(t *testing.T) {
	dbPath := "test_migrations.db"
	defer os.Remove(dbPath)

	config := &Config{Path: dbPath}
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	manager := NewMigrationManager(db)

	if err := manager.InitMigrationTable(); err != nil {
		t.Fatalf("Failed to initialize migration table: %v", err)
	}

	// Verify the table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query for migrations table: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 migrations table, got %d", count)
	}
}

func TestApplyMigration(t *testing.T) {
	dbPath := "test_apply_migration.db"
	defer os.Remove(dbPath)

	config := &Config{Path: dbPath}
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	manager := NewMigrationManager(db)

	if err := manager.InitMigrationTable(); err != nil {
		t.Fatalf("Failed to initialize migration table: %v", err)
	}

	migration := Migration{
		Version:     1,
		Description: "Test migration",
		Up:          "CREATE TABLE test_table (id INTEGER PRIMARY KEY);",
		Down:        "DROP TABLE test_table;",
	}

	if err := manager.ApplyMigration(migration); err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Verify the migration was recorded
	appliedVersions, err := manager.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(appliedVersions) != 1 || appliedVersions[0] != 1 {
		t.Errorf("Expected migration version 1 to be applied, got %v", appliedVersions)
	}

	// Verify the table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query for test table: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 test table, got %d", count)
	}
}

func TestMigrate(t *testing.T) {
	dbPath := "test_migrate.db"
	defer os.Remove(dbPath)

	config := &Config{Path: dbPath}
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	manager := NewMigrationManager(db)

	migrations := []Migration{
		{
			Version:     1,
			Description: "First migration",
			Up:          "CREATE TABLE table1 (id INTEGER PRIMARY KEY);",
			Down:        "DROP TABLE table1;",
		},
		{
			Version:     2,
			Description: "Second migration",
			Up:          "CREATE TABLE table2 (id INTEGER PRIMARY KEY);",
			Down:        "DROP TABLE table2;",
		},
	}

	if err := manager.Migrate(migrations); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Verify both migrations were applied
	appliedVersions, err := manager.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(appliedVersions) != 2 {
		t.Errorf("Expected 2 applied migrations, got %d", len(appliedVersions))
	}

	// Verify both tables were created
	for i, tableName := range []string{"table1", "table2"} {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query for %s: %v", tableName, err)
		}

		if count != 1 {
			t.Errorf("Expected 1 %s table, got %d", tableName, count)
		}

		if appliedVersions[i] != i+1 {
			t.Errorf("Expected migration version %d, got %d", i+1, appliedVersions[i])
		}
	}
}

func TestInitializeDatabase(t *testing.T) {
	dbPath := "test_initialize.db"
	defer os.Remove(dbPath)

	config := &Config{Path: dbPath}
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := InitializeDatabase(db); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify all expected tables were created
	expectedTables := []string{
		"messages", "recipients", "delivery_attempts",
		"log_entries", "audit_log", "queue_snapshots",
	}

	for _, tableName := range expectedTables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query for %s: %v", tableName, err)
		}

		if count != 1 {
			t.Errorf("Expected 1 %s table, got %d", tableName, count)
		}
	}

	// Verify indexes were created
	var indexCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&indexCount)
	if err != nil {
		t.Fatalf("Failed to query for indexes: %v", err)
	}

	if indexCount == 0 {
		t.Error("Expected indexes to be created, but found none")
	}
}
