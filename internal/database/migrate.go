package database

// GetAllMigrations returns all available migrations
func GetAllMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Initial schema creation",
			Up:          Schema,
			Down: `
				DROP INDEX IF EXISTS idx_queue_snapshots_timestamp;
				DROP INDEX IF EXISTS idx_audit_log_user_id;
				DROP INDEX IF EXISTS idx_audit_log_message_id;
				DROP INDEX IF EXISTS idx_audit_log_action;
				DROP INDEX IF EXISTS idx_audit_log_timestamp;
				DROP INDEX IF EXISTS idx_log_entries_sender;
				DROP INDEX IF EXISTS idx_log_entries_log_type;
				DROP INDEX IF EXISTS idx_log_entries_event;
				DROP INDEX IF EXISTS idx_log_entries_message_id;
				DROP INDEX IF EXISTS idx_log_entries_timestamp;
				DROP INDEX IF EXISTS idx_delivery_attempts_status;
				DROP INDEX IF EXISTS idx_delivery_attempts_recipient;
				DROP INDEX IF EXISTS idx_delivery_attempts_timestamp;
				DROP INDEX IF EXISTS idx_delivery_attempts_message_id;
				DROP INDEX IF EXISTS idx_recipients_recipient;
				DROP INDEX IF EXISTS idx_recipients_status;
				DROP INDEX IF EXISTS idx_recipients_message_id;
				DROP INDEX IF EXISTS idx_messages_created_at;
				DROP INDEX IF EXISTS idx_messages_sender;
				DROP INDEX IF EXISTS idx_messages_status;
				DROP INDEX IF EXISTS idx_messages_timestamp;
				DROP TABLE IF EXISTS queue_snapshots;
				DROP TABLE IF EXISTS audit_log;
				DROP TABLE IF EXISTS log_entries;
				DROP TABLE IF EXISTS delivery_attempts;
				DROP TABLE IF EXISTS recipients;
				DROP TABLE IF EXISTS messages;
			`,
		},
	}
}

// InitializeDatabase initializes the database with the schema
func InitializeDatabase(db *DB) error {
	manager := NewMigrationManager(db)
	migrations := GetAllMigrations()

	return manager.Migrate(migrations)
}

// MigrateUp applies all pending migrations
func MigrateUp(db *DB) error {
	return InitializeDatabase(db)
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *DB) (map[int]bool, error) {
	manager := NewMigrationManager(db)
	migrations := GetAllMigrations()

	return manager.GetMigrationStatus(migrations)
}
