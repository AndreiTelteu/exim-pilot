package database

import (
	"context"
	"fmt"
	"log"
	"time"
)

// OptimizationService handles database performance optimizations
type OptimizationService struct {
	db *DB
}

// NewOptimizationService creates a new optimization service
func NewOptimizationService(db *DB) *OptimizationService {
	return &OptimizationService{db: db}
}

// OptimizeDatabase performs various database optimizations
func (s *OptimizationService) OptimizeDatabase(ctx context.Context) error {
	log.Println("Starting database optimization...")

	// Run ANALYZE to update table statistics
	if err := s.analyzeDatabase(ctx); err != nil {
		log.Printf("Failed to analyze database: %v", err)
	}

	// Run VACUUM to reclaim space and defragment
	if err := s.vacuumDatabase(ctx); err != nil {
		log.Printf("Failed to vacuum database: %v", err)
	}

	// Optimize SQLite settings
	if err := s.optimizeSQLiteSettings(ctx); err != nil {
		log.Printf("Failed to optimize SQLite settings: %v", err)
	}

	log.Println("Database optimization completed")
	return nil
}

// analyzeDatabase updates table statistics for better query planning
func (s *OptimizationService) analyzeDatabase(ctx context.Context) error {
	tables := []string{
		"messages", "recipients", "delivery_attempts",
		"log_entries", "audit_log", "queue_snapshots",
		"message_notes", "message_tags", "users", "sessions",
	}

	for _, table := range tables {
		query := fmt.Sprintf("ANALYZE %s", table)
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to analyze table %s: %w", table, err)
		}
	}

	return nil
}

// vacuumDatabase reclaims space and defragments the database
func (s *OptimizationService) vacuumDatabase(ctx context.Context) error {
	// Use incremental vacuum to avoid locking the database for too long
	if _, err := s.db.ExecContext(ctx, "PRAGMA incremental_vacuum"); err != nil {
		return fmt.Errorf("failed to run incremental vacuum: %w", err)
	}

	return nil
}

// optimizeSQLiteSettings configures SQLite for optimal performance
func (s *OptimizationService) optimizeSQLiteSettings(ctx context.Context) error {
	settings := map[string]string{
		// Use WAL mode for better concurrency
		"journal_mode": "WAL",
		// Enable foreign key constraints
		"foreign_keys": "ON",
		// Optimize cache size (in pages, -2000 = 2MB)
		"cache_size": "-2000",
		// Optimize page size for better I/O
		"page_size": "4096",
		// Use memory for temporary tables
		"temp_store": "MEMORY",
		// Optimize synchronous mode for WAL
		"synchronous": "NORMAL",
		// Set busy timeout for better concurrency
		"busy_timeout": "5000",
		// Optimize mmap size (64MB)
		"mmap_size": "67108864",
	}

	for pragma, value := range settings {
		query := fmt.Sprintf("PRAGMA %s = %s", pragma, value)
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			log.Printf("Failed to set PRAGMA %s: %v", pragma, err)
			// Continue with other settings even if one fails
		}
	}

	return nil
}

// GetDatabaseStats returns database statistics for monitoring
func (s *OptimizationService) GetDatabaseStats(ctx context.Context) (*DatabaseStats, error) {
	stats := &DatabaseStats{
		Timestamp: time.Now(),
	}

	// Get database file size
	var pageCount, pageSize int64
	if err := s.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount); err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize); err != nil {
		return nil, fmt.Errorf("failed to get page size: %w", err)
	}
	stats.DatabaseSize = pageCount * pageSize

	// Get table statistics
	stats.TableStats = make(map[string]TableStats)
	tables := []string{
		"messages", "recipients", "delivery_attempts",
		"log_entries", "audit_log", "queue_snapshots",
	}

	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			log.Printf("Failed to get count for table %s: %v", table, err)
			continue
		}
		stats.TableStats[table] = TableStats{
			RowCount: count,
		}
	}

	// Get index usage statistics (if available)
	stats.IndexStats = s.getIndexStats(ctx)

	return stats, nil
}

// getIndexStats returns index usage statistics
func (s *OptimizationService) getIndexStats(ctx context.Context) map[string]IndexStats {
	indexStats := make(map[string]IndexStats)

	// Query SQLite's index list
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, tbl_name 
		FROM sqlite_master 
		WHERE type = 'index' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		log.Printf("Failed to query index list: %v", err)
		return indexStats
	}
	defer rows.Close()

	for rows.Next() {
		var indexName, tableName string
		if err := rows.Scan(&indexName, &tableName); err != nil {
			continue
		}

		indexStats[indexName] = IndexStats{
			TableName: tableName,
			// Note: SQLite doesn't provide detailed index usage stats
			// In a production system, you might track this separately
		}
	}

	return indexStats
}

// CleanupOldData removes old data based on retention policies
func (s *OptimizationService) CleanupOldData(ctx context.Context, retentionDays map[string]int) error {
	log.Println("Starting data cleanup...")

	// Default retention policies if not specified
	if retentionDays == nil {
		retentionDays = map[string]int{
			"log_entries":       90,  // 3 months
			"audit_log":         365, // 1 year
			"queue_snapshots":   30,  // 1 month
			"delivery_attempts": 180, // 6 months
		}
	}

	for table, days := range retentionDays {
		cutoffDate := time.Now().AddDate(0, 0, -days)

		var query string
		switch table {
		case "log_entries":
			query = "DELETE FROM log_entries WHERE timestamp < ?"
		case "audit_log":
			query = "DELETE FROM audit_log WHERE timestamp < ?"
		case "queue_snapshots":
			query = "DELETE FROM queue_snapshots WHERE timestamp < ?"
		case "delivery_attempts":
			query = "DELETE FROM delivery_attempts WHERE timestamp < ?"
		default:
			continue
		}

		result, err := s.db.ExecContext(ctx, query, cutoffDate)
		if err != nil {
			log.Printf("Failed to cleanup table %s: %v", table, err)
			continue
		}

		if rowsAffected, err := result.RowsAffected(); err == nil {
			log.Printf("Cleaned up %d rows from %s (older than %d days)", rowsAffected, table, days)
		}
	}

	// Run vacuum after cleanup to reclaim space
	if err := s.vacuumDatabase(ctx); err != nil {
		log.Printf("Failed to vacuum after cleanup: %v", err)
	}

	log.Println("Data cleanup completed")
	return nil
}

// OptimizeQueries provides query optimization hints
func (s *OptimizationService) OptimizeQueries() *QueryOptimizationHints {
	return &QueryOptimizationHints{
		LogSearchQueries: []string{
			// Use indexes for timestamp-based queries
			"SELECT * FROM log_entries WHERE timestamp BETWEEN ? AND ? ORDER BY timestamp DESC LIMIT ?",
			// Use composite indexes for filtered searches
			"SELECT * FROM log_entries WHERE log_type = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp DESC LIMIT ?",
			// Use message_id index for message correlation
			"SELECT * FROM log_entries WHERE message_id = ? ORDER BY timestamp ASC",
		},
		QueueQueries: []string{
			// Use status index for queue filtering
			"SELECT * FROM messages WHERE status IN (?, ?, ?) ORDER BY timestamp DESC LIMIT ?",
			// Use composite index for sender-based queries
			"SELECT * FROM messages WHERE sender LIKE ? AND status = ? ORDER BY timestamp DESC LIMIT ?",
		},
		ReportingQueries: []string{
			// Use timestamp indexes for reporting
			"SELECT DATE(timestamp) as date, COUNT(*) as count FROM log_entries WHERE timestamp BETWEEN ? AND ? GROUP BY DATE(timestamp)",
			// Use composite indexes for deliverability reports
			"SELECT status, COUNT(*) as count FROM delivery_attempts WHERE timestamp BETWEEN ? AND ? GROUP BY status",
		},
	}
}

// DatabaseStats represents database performance statistics
type DatabaseStats struct {
	Timestamp    time.Time             `json:"timestamp"`
	DatabaseSize int64                 `json:"database_size"`
	TableStats   map[string]TableStats `json:"table_stats"`
	IndexStats   map[string]IndexStats `json:"index_stats"`
}

// TableStats represents statistics for a single table
type TableStats struct {
	RowCount int64 `json:"row_count"`
}

// IndexStats represents statistics for a single index
type IndexStats struct {
	TableName string `json:"table_name"`
	// Note: SQLite doesn't provide detailed usage stats
	// In production, you might track these separately
}

// QueryOptimizationHints provides optimized query patterns
type QueryOptimizationHints struct {
	LogSearchQueries []string `json:"log_search_queries"`
	QueueQueries     []string `json:"queue_queries"`
	ReportingQueries []string `json:"reporting_queries"`
}
