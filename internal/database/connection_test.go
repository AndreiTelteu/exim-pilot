package database

import (
	"os"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	// Use a temporary database file
	dbPath := "test_connection.db"
	defer os.Remove(dbPath)

	config := &Config{
		Path:            dbPath,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	}

	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test that we can ping the database
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Verify the path is set correctly
	if db.Path() != dbPath {
		t.Errorf("Expected path %s, got %s", dbPath, db.Path())
	}
}

func TestConnectWithDefaultConfig(t *testing.T) {
	// Use a temporary database file
	dbPath := "test_default.db"
	defer os.Remove(dbPath)

	config := DefaultConfig()
	config.Path = dbPath

	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect with default config: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

func TestConnectWithNilConfig(t *testing.T) {
	// This should use default config
	db, err := Connect(nil)
	if err != nil {
		t.Fatalf("Failed to connect with nil config: %v", err)
	}
	defer db.Close()
	defer os.Remove("exim-pilot.db") // Clean up default db file

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

func TestBeginTx(t *testing.T) {
	dbPath := "test_tx.db"
	defer os.Remove(dbPath)

	config := &Config{Path: dbPath}
	db, err := Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
}
