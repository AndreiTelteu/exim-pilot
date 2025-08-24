package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// RetentionService handles data retention policies and cleanup
type RetentionService struct {
	db     *DB
	config RetentionConfig
}

// RetentionConfig defines retention policies for different data types
type RetentionConfig struct {
	LogEntriesRetentionDays       int  `json:"log_entries_retention_days"`
	AuditLogRetentionDays         int  `json:"audit_log_retention_days"`
	QueueSnapshotsRetentionDays   int  `json:"queue_snapshots_retention_days"`
	DeliveryAttemptsRetentionDays int  `json:"delivery_attempts_retention_days"`
	SessionsRetentionDays         int  `json:"sessions_retention_days"`
	EnableAutoCleanup             bool `json:"enable_auto_cleanup"`
	CleanupBatchSize              int  `json:"cleanup_batch_size"`
	CleanupIntervalHours          int  `json:"cleanup_interval_hours"`
}

// DefaultRetentionConfig returns default retention configuration
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		LogEntriesRetentionDays:       90,  // 3 months
		AuditLogRetentionDays:         365, // 1 year
		QueueSnapshotsRetentionDays:   30,  // 1 month
		DeliveryAttemptsRetentionDays: 180, // 6 months
		SessionsRetentionDays:         7,   // 1 week
		EnableAutoCleanup:             true,
		CleanupBatchSize:              1000,
		CleanupIntervalHours:          24, // Daily cleanup
	}
}

// NewRetentionService creates a new retention service
func NewRetentionService(db *DB, config RetentionConfig) *RetentionService {
	return &RetentionService{
		db:     db,
		config: config,
	}
}

// CleanupExpiredData removes expired data based on retention policies
func (rs *RetentionService) CleanupExpiredData(ctx context.Context) (*CleanupResult, error) {
	log.Println("Starting data retention cleanup...")

	result := &CleanupResult{
		StartTime: time.Now(),
		Tables:    make(map[string]TableCleanupResult),
	}

	// Define cleanup operations
	cleanupOps := []struct {
		table         string
		retentionDays int
		timestampCol  string
		description   string
	}{
		{"log_entries", rs.config.LogEntriesRetentionDays, "timestamp", "Log entries"},
		{"audit_log", rs.config.AuditLogRetentionDays, "timestamp", "Audit log entries"},
		{"queue_snapshots", rs.config.QueueSnapshotsRetentionDays, "timestamp", "Queue snapshots"},
		{"delivery_attempts", rs.config.DeliveryAttemptsRetentionDays, "timestamp", "Delivery attempts"},
		{"sessions", rs.config.SessionsRetentionDays, "created_at", "User sessions"},
	}

	// Perform cleanup for each table
	for _, op := range cleanupOps {
		tableResult, err := rs.cleanupTable(ctx, op.table, op.timestampCol, op.retentionDays, op.description)
		if err != nil {
			log.Printf("Failed to cleanup %s: %v", op.description, err)
			tableResult.Error = err.Error()
		}
		result.Tables[op.table] = tableResult
		result.TotalRowsDeleted += tableResult.RowsDeleted
	}

	// Clean up expired sessions (additional logic for sessions)
	if err := rs.cleanupExpiredSessions(ctx); err != nil {
		log.Printf("Failed to cleanup expired sessions: %v", err)
	}

	// Run database optimization after cleanup
	if err := rs.optimizeAfterCleanup(ctx); err != nil {
		log.Printf("Failed to optimize database after cleanup: %v", err)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Data retention cleanup completed: deleted %d rows in %v",
		result.TotalRowsDeleted, result.Duration)

	return result, nil
}

// cleanupTable removes expired data from a specific table
func (rs *RetentionService) cleanupTable(ctx context.Context, tableName, timestampCol string, retentionDays int, description string) (TableCleanupResult, error) {
	result := TableCleanupResult{
		TableName: tableName,
		StartTime: time.Now(),
	}

	if retentionDays <= 0 {
		log.Printf("Skipping cleanup for %s: retention disabled", description)
		return result, nil
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Count rows to be deleted first
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s < ?", tableName, timestampCol)
	var rowsToDelete int64
	if err := rs.db.QueryRowContext(ctx, countQuery, cutoffDate).Scan(&rowsToDelete); err != nil {
		return result, fmt.Errorf("failed to count rows for deletion: %w", err)
	}

	if rowsToDelete == 0 {
		log.Printf("No expired data found in %s", description)
		result.EndTime = time.Now()
		return result, nil
	}

	log.Printf("Found %d expired rows in %s (older than %s)",
		rowsToDelete, description, cutoffDate.Format("2006-01-02"))

	// Delete in batches to avoid long-running transactions
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE %s < ? 
		AND rowid IN (
			SELECT rowid FROM %s 
			WHERE %s < ? 
			LIMIT ?
		)`, tableName, timestampCol, tableName, timestampCol)

	var totalDeleted int64
	batchSize := int64(rs.config.CleanupBatchSize)

	for totalDeleted < rowsToDelete {
		// Use a transaction for each batch
		tx, err := rs.db.BeginTx()
		if err != nil {
			return result, fmt.Errorf("failed to begin transaction: %w", err)
		}

		batchResult, err := tx.ExecContext(ctx, deleteQuery, cutoffDate, cutoffDate, batchSize)
		if err != nil {
			tx.Rollback()
			return result, fmt.Errorf("failed to delete batch: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return result, fmt.Errorf("failed to commit batch deletion: %w", err)
		}

		batchDeleted, _ := batchResult.RowsAffected()
		totalDeleted += batchDeleted

		if batchDeleted == 0 {
			// No more rows to delete
			break
		}

		// Small delay between batches to avoid overwhelming the database
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}

	result.RowsDeleted = totalDeleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Deleted %d rows from %s in %v", totalDeleted, description, result.Duration)
	return result, nil
}

// cleanupExpiredSessions removes expired user sessions
func (rs *RetentionService) cleanupExpiredSessions(ctx context.Context) error {
	query := "DELETE FROM sessions WHERE expires_at < ?"
	result, err := rs.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected > 0 {
		log.Printf("Deleted %d expired sessions", rowsAffected)
	}

	return nil
}

// optimizeAfterCleanup runs database optimization after cleanup
func (rs *RetentionService) optimizeAfterCleanup(ctx context.Context) error {
	// Run incremental vacuum to reclaim space
	if _, err := rs.db.ExecContext(ctx, "PRAGMA incremental_vacuum"); err != nil {
		return fmt.Errorf("failed to run incremental vacuum: %w", err)
	}

	// Update table statistics
	if _, err := rs.db.ExecContext(ctx, "ANALYZE"); err != nil {
		return fmt.Errorf("failed to analyze database: %w", err)
	}

	return nil
}

// GetRetentionStatus returns current retention status and statistics
func (rs *RetentionService) GetRetentionStatus(ctx context.Context) (*RetentionStatus, error) {
	status := &RetentionStatus{
		Config:     rs.config,
		Timestamp:  time.Now(),
		TableStats: make(map[string]RetentionTableStats),
	}

	tables := []struct {
		name          string
		timestampCol  string
		retentionDays int
	}{
		{"log_entries", "timestamp", rs.config.LogEntriesRetentionDays},
		{"audit_log", "timestamp", rs.config.AuditLogRetentionDays},
		{"queue_snapshots", "timestamp", rs.config.QueueSnapshotsRetentionDays},
		{"delivery_attempts", "timestamp", rs.config.DeliveryAttemptsRetentionDays},
		{"sessions", "created_at", rs.config.SessionsRetentionDays},
	}

	for _, table := range tables {
		stats, err := rs.getTableRetentionStats(ctx, table.name, table.timestampCol, table.retentionDays)
		if err != nil {
			log.Printf("Failed to get retention stats for %s: %v", table.name, err)
			continue
		}
		status.TableStats[table.name] = stats
	}

	return status, nil
}

// getTableRetentionStats gets retention statistics for a specific table
func (rs *RetentionService) getTableRetentionStats(ctx context.Context, tableName, timestampCol string, retentionDays int) (RetentionTableStats, error) {
	stats := RetentionTableStats{
		TableName:     tableName,
		RetentionDays: retentionDays,
	}

	// Get total row count
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err := rs.db.QueryRowContext(ctx, totalQuery).Scan(&stats.TotalRows); err != nil {
		return stats, fmt.Errorf("failed to get total rows: %w", err)
	}

	if retentionDays > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

		// Get expired row count
		expiredQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s < ?", tableName, timestampCol)
		if err := rs.db.QueryRowContext(ctx, expiredQuery, cutoffDate).Scan(&stats.ExpiredRows); err != nil {
			return stats, fmt.Errorf("failed to get expired rows: %w", err)
		}

		// Get oldest record timestamp
		oldestQuery := fmt.Sprintf("SELECT MIN(%s) FROM %s", timestampCol, tableName)
		var oldestTime sql.NullTime
		if err := rs.db.QueryRowContext(ctx, oldestQuery).Scan(&oldestTime); err == nil && oldestTime.Valid {
			stats.OldestRecord = &oldestTime.Time
		}

		// Get newest record timestamp
		newestQuery := fmt.Sprintf("SELECT MAX(%s) FROM %s", timestampCol, tableName)
		var newestTime sql.NullTime
		if err := rs.db.QueryRowContext(ctx, newestQuery).Scan(&newestTime); err == nil && newestTime.Valid {
			stats.NewestRecord = &newestTime.Time
		}
	}

	return stats, nil
}

// ScheduleAutoCleanup starts automatic cleanup based on configuration
func (rs *RetentionService) ScheduleAutoCleanup(ctx context.Context) {
	if !rs.config.EnableAutoCleanup {
		log.Println("Auto cleanup is disabled")
		return
	}

	log.Printf("Starting auto cleanup scheduler (interval: %d hours)", rs.config.CleanupIntervalHours)

	ticker := time.NewTicker(time.Duration(rs.config.CleanupIntervalHours) * time.Hour)
	defer ticker.Stop()

	// Run initial cleanup
	go func() {
		if _, err := rs.CleanupExpiredData(ctx); err != nil {
			log.Printf("Initial cleanup failed: %v", err)
		}
	}()

	// Schedule periodic cleanup
	for {
		select {
		case <-ctx.Done():
			log.Println("Auto cleanup scheduler stopped")
			return
		case <-ticker.C:
			go func() {
				if _, err := rs.CleanupExpiredData(ctx); err != nil {
					log.Printf("Scheduled cleanup failed: %v", err)
				}
			}()
		}
	}
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	StartTime        time.Time                     `json:"start_time"`
	EndTime          time.Time                     `json:"end_time"`
	Duration         time.Duration                 `json:"duration"`
	TotalRowsDeleted int64                         `json:"total_rows_deleted"`
	Tables           map[string]TableCleanupResult `json:"tables"`
}

// TableCleanupResult represents cleanup result for a single table
type TableCleanupResult struct {
	TableName   string        `json:"table_name"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	RowsDeleted int64         `json:"rows_deleted"`
	Error       string        `json:"error,omitempty"`
}

// RetentionStatus represents current retention status
type RetentionStatus struct {
	Config     RetentionConfig                `json:"config"`
	Timestamp  time.Time                      `json:"timestamp"`
	TableStats map[string]RetentionTableStats `json:"table_stats"`
}

// RetentionTableStats represents retention statistics for a table
type RetentionTableStats struct {
	TableName     string     `json:"table_name"`
	RetentionDays int        `json:"retention_days"`
	TotalRows     int64      `json:"total_rows"`
	ExpiredRows   int64      `json:"expired_rows"`
	OldestRecord  *time.Time `json:"oldest_record,omitempty"`
	NewestRecord  *time.Time `json:"newest_record,omitempty"`
}
