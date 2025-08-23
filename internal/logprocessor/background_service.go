package logprocessor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// BackgroundService handles background log processing tasks
type BackgroundService struct {
	repository *database.Repository
	aggregator *LogAggregator
	config     BackgroundConfig
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mu         sync.RWMutex
}

// BackgroundConfig holds configuration for background processing
type BackgroundConfig struct {
	// Correlation settings
	CorrelationInterval   time.Duration `json:"correlation_interval"`
	CorrelationBatchHours int           `json:"correlation_batch_hours"`

	// Retention settings
	LogRetentionDays      int `json:"log_retention_days"`
	AuditRetentionDays    int `json:"audit_retention_days"`
	SnapshotRetentionDays int `json:"snapshot_retention_days"`

	// Cleanup settings
	CleanupInterval  time.Duration `json:"cleanup_interval"`
	CleanupBatchSize int           `json:"cleanup_batch_size"`

	// Performance settings
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`
}

// DefaultBackgroundConfig returns default configuration
func DefaultBackgroundConfig() BackgroundConfig {
	return BackgroundConfig{
		CorrelationInterval:   30 * time.Minute,
		CorrelationBatchHours: 24,
		LogRetentionDays:      90,
		AuditRetentionDays:    365,
		SnapshotRetentionDays: 30,
		CleanupInterval:       6 * time.Hour,
		CleanupBatchSize:      1000,
		MaxConcurrentTasks:    3,
	}
}

// NewBackgroundService creates a new background service
func NewBackgroundService(repository *database.Repository, config BackgroundConfig) *BackgroundService {
	ctx, cancel := context.WithCancel(context.Background())

	return &BackgroundService{
		repository: repository,
		aggregator: NewLogAggregator(repository),
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins background processing
func (s *BackgroundService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("background service is already running")
	}

	s.running = true

	// Start correlation worker
	s.wg.Add(1)
	go s.correlationWorker()

	// Start cleanup worker
	s.wg.Add(1)
	go s.cleanupWorker()

	// Start metrics worker
	s.wg.Add(1)
	go s.metricsWorker()

	log.Println("Background service started")
	return nil
}

// Stop stops background processing
func (s *BackgroundService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.cancel()
	s.wg.Wait()
	s.running = false

	log.Println("Background service stopped")
	return nil
}

// IsRunning returns whether the service is running
func (s *BackgroundService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// correlationWorker runs periodic log correlation
func (s *BackgroundService) correlationWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CorrelationInterval)
	defer ticker.Stop()

	log.Printf("Correlation worker started (interval: %v)", s.config.CorrelationInterval)

	// Run initial correlation for recent data
	s.runCorrelation()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runCorrelation()
		}
	}
}

// runCorrelation performs log correlation for recent entries
func (s *BackgroundService) runCorrelation() {
	startTime := time.Now().Add(-time.Duration(s.config.CorrelationBatchHours) * time.Hour)
	endTime := time.Now()

	log.Printf("Starting log correlation for period %s to %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	start := time.Now()
	if err := s.aggregator.CorrelateLogEntries(s.ctx, startTime, endTime); err != nil {
		log.Printf("Log correlation failed: %v", err)
		return
	}

	duration := time.Since(start)
	log.Printf("Log correlation completed in %v", duration)
}

// cleanupWorker runs periodic data cleanup
func (s *BackgroundService) cleanupWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	log.Printf("Cleanup worker started (interval: %v)", s.config.CleanupInterval)

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runCleanup()
		}
	}
}

// runCleanup performs data retention cleanup
func (s *BackgroundService) runCleanup() {
	log.Println("Starting data retention cleanup")

	start := time.Now()
	totalCleaned := 0

	// Clean up old log entries
	if s.config.LogRetentionDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -s.config.LogRetentionDays)
		cleaned, err := s.cleanupLogEntries(cutoff)
		if err != nil {
			log.Printf("Failed to cleanup log entries: %v", err)
		} else {
			totalCleaned += cleaned
			log.Printf("Cleaned up %d old log entries (older than %s)", cleaned, cutoff.Format("2006-01-02"))
		}
	}

	// Clean up old audit logs
	if s.config.AuditRetentionDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -s.config.AuditRetentionDays)
		cleaned, err := s.cleanupAuditLogs(cutoff)
		if err != nil {
			log.Printf("Failed to cleanup audit logs: %v", err)
		} else {
			totalCleaned += cleaned
			log.Printf("Cleaned up %d old audit logs (older than %s)", cleaned, cutoff.Format("2006-01-02"))
		}
	}

	// Clean up old queue snapshots
	if s.config.SnapshotRetentionDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -s.config.SnapshotRetentionDays)
		cleaned, err := s.cleanupQueueSnapshots(cutoff)
		if err != nil {
			log.Printf("Failed to cleanup queue snapshots: %v", err)
		} else {
			totalCleaned += cleaned
			log.Printf("Cleaned up %d old queue snapshots (older than %s)", cleaned, cutoff.Format("2006-01-02"))
		}
	}

	// Clean up orphaned messages and related data
	orphansCleaned, err := s.cleanupOrphanedData()
	if err != nil {
		log.Printf("Failed to cleanup orphaned data: %v", err)
	} else {
		totalCleaned += orphansCleaned
		if orphansCleaned > 0 {
			log.Printf("Cleaned up %d orphaned records", orphansCleaned)
		}
	}

	duration := time.Since(start)
	log.Printf("Data cleanup completed in %v (total records cleaned: %d)", duration, totalCleaned)
}

// cleanupLogEntries removes old log entries
func (s *BackgroundService) cleanupLogEntries(cutoff time.Time) (int, error) {
	query := "DELETE FROM log_entries WHERE timestamp < ?"

	result, err := s.repository.GetDB().Exec(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old log entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return int(rowsAffected), nil
}

// cleanupAuditLogs removes old audit logs
func (s *BackgroundService) cleanupAuditLogs(cutoff time.Time) (int, error) {
	query := "DELETE FROM audit_log WHERE timestamp < ?"

	result, err := s.repository.GetDB().Exec(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return int(rowsAffected), nil
}

// cleanupQueueSnapshots removes old queue snapshots
func (s *BackgroundService) cleanupQueueSnapshots(cutoff time.Time) (int, error) {
	snapshotRepo := database.NewQueueSnapshotRepository(s.repository.GetDB())
	rowsAffected, err := snapshotRepo.DeleteOlderThan(cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old queue snapshots: %w", err)
	}

	return int(rowsAffected), nil
}

// cleanupOrphanedData removes orphaned records
func (s *BackgroundService) cleanupOrphanedData() (int, error) {
	totalCleaned := 0

	// Clean up recipients without messages
	query := `DELETE FROM recipients WHERE message_id NOT IN (SELECT id FROM messages)`
	result, err := s.repository.GetDB().Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete orphaned recipients: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows for recipients: %w", err)
	}
	totalCleaned += int(rowsAffected)

	// Clean up delivery attempts without messages
	query = `DELETE FROM delivery_attempts WHERE message_id NOT IN (SELECT id FROM messages)`
	result, err = s.repository.GetDB().Exec(query)
	if err != nil {
		return totalCleaned, fmt.Errorf("failed to delete orphaned delivery attempts: %w", err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return totalCleaned, fmt.Errorf("failed to get affected rows for delivery attempts: %w", err)
	}
	totalCleaned += int(rowsAffected)

	// Clean up messages without any log entries (very old or corrupted data)
	cutoff := time.Now().AddDate(0, 0, -7) // Only clean messages older than 7 days without log entries
	query = `DELETE FROM messages WHERE id NOT IN (
		SELECT DISTINCT message_id FROM log_entries WHERE message_id IS NOT NULL
	) AND created_at < ?`
	result, err = s.repository.GetDB().Exec(query, cutoff)
	if err != nil {
		return totalCleaned, fmt.Errorf("failed to delete orphaned messages: %w", err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return totalCleaned, fmt.Errorf("failed to get affected rows for messages: %w", err)
	}
	totalCleaned += int(rowsAffected)

	return totalCleaned, nil
}

// metricsWorker collects and stores system metrics
func (s *BackgroundService) metricsWorker() {
	defer s.wg.Done()

	// Run metrics collection every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Metrics worker started")

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// collectMetrics collects and stores system metrics
func (s *BackgroundService) collectMetrics() {
	// Collect database size metrics
	if err := s.collectDatabaseMetrics(); err != nil {
		log.Printf("Failed to collect database metrics: %v", err)
	}

	// Collect processing metrics
	if err := s.collectProcessingMetrics(); err != nil {
		log.Printf("Failed to collect processing metrics: %v", err)
	}
}

// collectDatabaseMetrics collects database size and performance metrics
func (s *BackgroundService) collectDatabaseMetrics() error {
	// Get table sizes
	tables := []string{"messages", "recipients", "delivery_attempts", "log_entries", "audit_log", "queue_snapshots"}

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := s.repository.GetDB().QueryRow(query).Scan(&count); err != nil {
			log.Printf("Failed to get count for table %s: %v", table, err)
			continue
		}

		// Log metrics (in production, you might want to store these in a metrics table)
		log.Printf("Table %s: %d records", table, count)
	}

	return nil
}

// collectProcessingMetrics collects processing performance metrics
func (s *BackgroundService) collectProcessingMetrics() error {
	// Get recent processing statistics
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)

	// Count log entries processed in the last hour
	var recentLogCount int
	query := "SELECT COUNT(*) FROM log_entries WHERE created_at >= ?"
	if err := s.repository.GetDB().QueryRow(query, hourAgo).Scan(&recentLogCount); err != nil {
		return fmt.Errorf("failed to get recent log count: %w", err)
	}

	// Count messages processed in the last hour
	var recentMessageCount int
	query = "SELECT COUNT(*) FROM messages WHERE created_at >= ?"
	if err := s.repository.GetDB().QueryRow(query, hourAgo).Scan(&recentMessageCount); err != nil {
		return fmt.Errorf("failed to get recent message count: %w", err)
	}

	log.Printf("Processing metrics - Last hour: %d log entries, %d messages", recentLogCount, recentMessageCount)

	return nil
}

// ProcessHistoricalLogs processes historical log files in the background
func (s *BackgroundService) ProcessHistoricalLogs(logPaths []string) error {
	log.Println("Starting background historical log processing")

	// Process each log file in a separate goroutine with concurrency control
	semaphore := make(chan struct{}, s.config.MaxConcurrentTasks)
	var wg sync.WaitGroup

	for _, logPath := range logPaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := s.processHistoricalLogFile(path); err != nil {
				log.Printf("Failed to process historical log file %s: %v", path, err)
			}
		}(logPath)
	}

	wg.Wait()
	log.Println("Background historical log processing completed")

	return nil
}

// processHistoricalLogFile processes a single historical log file
func (s *BackgroundService) processHistoricalLogFile(logPath string) error {
	// This would integrate with the existing log monitor's historical processing
	// For now, we'll just log that we would process it
	log.Printf("Processing historical log file: %s", logPath)

	// In a real implementation, this would:
	// 1. Read the log file
	// 2. Parse entries using the parser
	// 3. Store entries in batches
	// 4. Update correlation data

	return nil
}

// GetStatus returns the current status of the background service
func (s *BackgroundService) GetStatus() BackgroundStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return BackgroundStatus{
		Running:            s.running,
		LastCorrelationRun: time.Now(), // In production, track this properly
		LastCleanupRun:     time.Now(), // In production, track this properly
		NextCorrelationRun: time.Now().Add(s.config.CorrelationInterval),
		NextCleanupRun:     time.Now().Add(s.config.CleanupInterval),
		Config:             s.config,
	}
}

// BackgroundStatus represents the current status of background processing
type BackgroundStatus struct {
	Running            bool             `json:"running"`
	LastCorrelationRun time.Time        `json:"last_correlation_run"`
	LastCleanupRun     time.Time        `json:"last_cleanup_run"`
	NextCorrelationRun time.Time        `json:"next_correlation_run"`
	NextCleanupRun     time.Time        `json:"next_cleanup_run"`
	Config             BackgroundConfig `json:"config"`
}

// UpdateConfig updates the background service configuration
func (s *BackgroundService) UpdateConfig(config BackgroundConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	log.Println("Background service configuration updated")

	return nil
}
