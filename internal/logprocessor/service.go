package logprocessor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// LogEntryCallback is called when a new log entry is processed
type LogEntryCallback func(entry *database.LogEntry)

// Service provides comprehensive log processing functionality
type Service struct {
	repository        *database.Repository
	aggregator        *LogAggregator
	backgroundService *BackgroundService
	searchService     *SearchService
	config            ServiceConfig
	logEntryCallback  LogEntryCallback
	mu                sync.RWMutex
}

// ServiceConfig holds configuration for the log processing service
type ServiceConfig struct {
	// Background processing
	BackgroundConfig BackgroundConfig `json:"background_config"`

	// Search settings
	DefaultSearchLimit int `json:"default_search_limit"`
	MaxSearchLimit     int `json:"max_search_limit"`

	// Performance settings
	BatchSize         int           `json:"batch_size"`
	ProcessingTimeout time.Duration `json:"processing_timeout"`

	// Feature flags
	EnableCorrelation bool `json:"enable_correlation"`
	EnableCleanup     bool `json:"enable_cleanup"`
	EnableMetrics     bool `json:"enable_metrics"`
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		BackgroundConfig:   DefaultBackgroundConfig(),
		DefaultSearchLimit: 100,
		MaxSearchLimit:     1000,
		BatchSize:          500,
		ProcessingTimeout:  30 * time.Second,
		EnableCorrelation:  true,
		EnableCleanup:      true,
		EnableMetrics:      true,
	}
}

// NewService creates a new log processing service
func NewService(repository *database.Repository, config ServiceConfig) *Service {
	service := &Service{
		repository:        repository,
		aggregator:        NewLogAggregator(repository),
		backgroundService: NewBackgroundService(repository, config.BackgroundConfig),
		searchService:     NewSearchService(repository),
		config:            config,
	}

	return service
}

// SetLogEntryCallback sets the callback function for new log entries
func (s *Service) SetLogEntryCallback(callback LogEntryCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logEntryCallback = callback
}

// Start starts the log processing service
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Starting log processing service")

	// Start background service if enabled
	if s.config.EnableCorrelation || s.config.EnableCleanup || s.config.EnableMetrics {
		if err := s.backgroundService.Start(); err != nil {
			return fmt.Errorf("failed to start background service: %w", err)
		}
	}

	log.Println("Log processing service started successfully")
	return nil
}

// Stop stops the log processing service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Stopping log processing service")

	// Stop background service
	if err := s.backgroundService.Stop(); err != nil {
		log.Printf("Error stopping background service: %v", err)
	}

	log.Println("Log processing service stopped")
	return nil
}

// ProcessLogEntry processes a single log entry with correlation
func (s *Service) ProcessLogEntry(ctx context.Context, entry *database.LogEntry) error {
	// Store the log entry
	if err := s.repository.CreateLogEntry(ctx, entry); err != nil {
		return fmt.Errorf("failed to store log entry: %w", err)
	}

	// Call the callback if set (for WebSocket broadcasting)
	s.mu.RLock()
	callback := s.logEntryCallback
	s.mu.RUnlock()

	if callback != nil {
		go callback(entry)
	}

	// If correlation is enabled and entry has a message ID, trigger correlation
	if s.config.EnableCorrelation && entry.MessageID != nil && *entry.MessageID != "" {
		go s.correlateMessageAsync(*entry.MessageID)
	}

	return nil
}

// ProcessLogEntries processes multiple log entries in batch
func (s *Service) ProcessLogEntries(ctx context.Context, entries []*database.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Process in batches
	batchSize := s.config.BatchSize
	messageIDs := make(map[string]bool)

	for i := 0; i < len(entries); i += batchSize {
		end := i + batchSize
		if end > len(entries) {
			end = len(entries)
		}

		batch := entries[i:end]

		// Store batch
		for _, entry := range batch {
			if err := s.repository.CreateLogEntry(ctx, entry); err != nil {
				log.Printf("Failed to store log entry: %v", err)
				continue
			}

			// Collect message IDs for correlation
			if s.config.EnableCorrelation && entry.MessageID != nil && *entry.MessageID != "" {
				messageIDs[*entry.MessageID] = true
			}
		}
	}

	// Trigger correlation for unique message IDs
	if s.config.EnableCorrelation {
		go s.correlateMessagesAsync(messageIDs)
	}

	log.Printf("Processed %d log entries in %d batches", len(entries), (len(entries)+batchSize-1)/batchSize)
	return nil
}

// correlateMessageAsync performs message correlation asynchronously
func (s *Service) correlateMessageAsync(messageID string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ProcessingTimeout)
	defer cancel()

	if _, err := s.aggregator.AggregateMessageData(ctx, messageID); err != nil {
		log.Printf("Failed to correlate message %s: %v", messageID, err)
	}
}

// correlateMessagesAsync performs correlation for multiple messages asynchronously
func (s *Service) correlateMessagesAsync(messageIDs map[string]bool) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ProcessingTimeout*5)
	defer cancel()

	for messageID := range messageIDs {
		if _, err := s.aggregator.AggregateMessageData(ctx, messageID); err != nil {
			log.Printf("Failed to correlate message %s: %v", messageID, err)
		}
	}
}

// SearchLogs performs advanced log search
func (s *Service) SearchLogs(ctx context.Context, criteria SearchCriteria) (*SearchResult, error) {
	// Validate and adjust search criteria
	if criteria.Limit <= 0 {
		criteria.Limit = s.config.DefaultSearchLimit
	}

	if criteria.Limit > s.config.MaxSearchLimit {
		criteria.Limit = s.config.MaxSearchLimit
	}

	return s.searchService.Search(ctx, criteria)
}

// GetMessageCorrelation gets correlated data for a message
func (s *Service) GetMessageCorrelation(ctx context.Context, messageID string) (*MessageCorrelation, error) {
	return s.aggregator.AggregateMessageData(ctx, messageID)
}

// GetMessageHistory gets the complete history for a message
func (s *Service) GetMessageHistory(ctx context.Context, messageID string) (*MessageCorrelation, error) {
	return s.searchService.SearchMessageHistory(ctx, messageID)
}

// FindSimilarMessages finds messages similar to the given message
func (s *Service) FindSimilarMessages(ctx context.Context, messageID string, limit int) ([]database.LogEntry, error) {
	if limit <= 0 {
		limit = s.config.DefaultSearchLimit
	}

	if limit > s.config.MaxSearchLimit {
		limit = s.config.MaxSearchLimit
	}

	return s.searchService.SearchSimilarMessages(ctx, messageID, limit)
}

// GetLogStatistics gets log statistics for a time period
func (s *Service) GetLogStatistics(ctx context.Context, startTime, endTime time.Time) (*LogStatistics, error) {
	return s.searchService.GetLogStatistics(ctx, startTime, endTime)
}

// TriggerCorrelation manually triggers correlation for a time period
func (s *Service) TriggerCorrelation(ctx context.Context, startTime, endTime time.Time) error {
	log.Printf("Manually triggering correlation for period %s to %s",
		startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	return s.aggregator.CorrelateLogEntries(ctx, startTime, endTime)
}

// TriggerCleanup manually triggers data cleanup
func (s *Service) TriggerCleanup(ctx context.Context) error {
	if !s.config.EnableCleanup {
		return fmt.Errorf("cleanup is disabled")
	}

	log.Println("Manually triggering data cleanup")

	// This would trigger the cleanup process
	// For now, we'll just log that it would run
	log.Println("Cleanup process would run here")

	return nil
}

// GetServiceStatus returns the current service status
func (s *Service) GetServiceStatus() ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ServiceStatus{
		Running:          true, // In production, track this properly
		BackgroundStatus: s.backgroundService.GetStatus(),
		Config:           s.config,
		LastActivity:     time.Now(), // In production, track this properly
	}
}

// UpdateConfig updates the service configuration
func (s *Service) UpdateConfig(config ServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	// Update background service config
	if err := s.backgroundService.UpdateConfig(config.BackgroundConfig); err != nil {
		return fmt.Errorf("failed to update background config: %w", err)
	}

	log.Println("Log processing service configuration updated")
	return nil
}

// ProcessHistoricalLogs processes historical log files
func (s *Service) ProcessHistoricalLogs(ctx context.Context, logPaths []string) error {
	log.Printf("Starting historical log processing for %d files", len(logPaths))

	return s.backgroundService.ProcessHistoricalLogs(logPaths)
}

// GetRetentionInfo returns information about data retention
func (s *Service) GetRetentionInfo(ctx context.Context) (*RetentionInfo, error) {
	info := &RetentionInfo{
		Config: RetentionConfig{
			LogRetentionDays:      s.config.BackgroundConfig.LogRetentionDays,
			AuditRetentionDays:    s.config.BackgroundConfig.AuditRetentionDays,
			SnapshotRetentionDays: s.config.BackgroundConfig.SnapshotRetentionDays,
		},
	}

	// Get current data counts
	tables := []string{"log_entries", "audit_log", "queue_snapshots"}
	info.CurrentCounts = make(map[string]int)

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := s.repository.GetDB().QueryRow(query).Scan(&count); err != nil {
			log.Printf("Failed to get count for table %s: %v", table, err)
			continue
		}
		info.CurrentCounts[table] = count
	}

	// Calculate oldest data
	var oldestLogEntry time.Time
	query := "SELECT MIN(timestamp) FROM log_entries"
	if err := s.repository.GetDB().QueryRow(query).Scan(&oldestLogEntry); err == nil {
		info.OldestLogEntry = &oldestLogEntry
	}

	var oldestAuditEntry time.Time
	query = "SELECT MIN(timestamp) FROM audit_log"
	if err := s.repository.GetDB().QueryRow(query).Scan(&oldestAuditEntry); err == nil {
		info.OldestAuditEntry = &oldestAuditEntry
	}

	return info, nil
}

// ServiceStatus represents the current status of the log processing service
type ServiceStatus struct {
	Running          bool             `json:"running"`
	BackgroundStatus BackgroundStatus `json:"background_status"`
	Config           ServiceConfig    `json:"config"`
	LastActivity     time.Time        `json:"last_activity"`
}

// RetentionInfo provides information about data retention
type RetentionInfo struct {
	Config           RetentionConfig `json:"config"`
	CurrentCounts    map[string]int  `json:"current_counts"`
	OldestLogEntry   *time.Time      `json:"oldest_log_entry,omitempty"`
	OldestAuditEntry *time.Time      `json:"oldest_audit_entry,omitempty"`
}

// RetentionConfig holds retention configuration
type RetentionConfig struct {
	LogRetentionDays      int `json:"log_retention_days"`
	AuditRetentionDays    int `json:"audit_retention_days"`
	SnapshotRetentionDays int `json:"snapshot_retention_days"`
}
