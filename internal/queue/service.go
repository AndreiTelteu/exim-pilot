package queue

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// Service provides queue management functionality
type Service struct {
	manager *Manager
	db      *database.DB
}

// NewService creates a new queue service
func NewService(eximPath string, db *database.DB) *Service {
	return &Service{
		manager: NewManager(eximPath, db),
		db:      db,
	}
}

// GetQueueStatus retrieves current queue status
func (s *Service) GetQueueStatus() (*QueueStatus, error) {
	return s.manager.ListQueue()
}

// GetMessageDetails retrieves detailed information about a message
func (s *Service) GetMessageDetails(messageID string) (*MessageDetails, error) {
	return s.manager.InspectMessage(messageID)
}

// CreateQueueSnapshot creates and stores a queue snapshot
func (s *Service) CreateQueueSnapshot() (*database.QueueSnapshot, error) {
	return s.manager.CreateSnapshot()
}

// StartPeriodicSnapshots starts a background goroutine that creates queue snapshots periodically
func (s *Service) StartPeriodicSnapshots(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Create initial snapshot
	if _, err := s.CreateQueueSnapshot(); err != nil {
		log.Printf("Failed to create initial queue snapshot: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping periodic queue snapshots")
			return
		case <-ticker.C:
			if _, err := s.CreateQueueSnapshot(); err != nil {
				log.Printf("Failed to create queue snapshot: %v", err)
			}
		}
	}
}

// GetQueueHealth returns queue health metrics
func (s *Service) GetQueueHealth() (*QueueHealth, error) {
	status, err := s.manager.ListQueue()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	health := &QueueHealth{
		TotalMessages:    status.TotalMessages,
		DeferredMessages: status.DeferredMessages,
		FrozenMessages:   status.FrozenMessages,
		OldestMessageAge: status.OldestMessageAge,
		Timestamp:        time.Now(),
	}

	// Calculate growth trend by comparing with recent snapshots
	repo := database.NewQueueSnapshotRepository(s.db)
	recent, err := repo.List(5, 0, nil, nil) // Get last 5 snapshots
	if err == nil && len(recent) > 1 {
		// Calculate average growth over recent snapshots
		var totalGrowth int
		for i := 0; i < len(recent)-1; i++ {
			growth := recent[i].TotalMessages - recent[i+1].TotalMessages
			totalGrowth += growth
		}
		health.GrowthTrend = totalGrowth / (len(recent) - 1)
	}

	return health, nil
}

// QueueHealth represents queue health metrics
type QueueHealth struct {
	TotalMessages    int           `json:"total_messages"`
	DeferredMessages int           `json:"deferred_messages"`
	FrozenMessages   int           `json:"frozen_messages"`
	OldestMessageAge time.Duration `json:"oldest_message_age"`
	GrowthTrend      int           `json:"growth_trend"` // Messages per snapshot interval
	Timestamp        time.Time     `json:"timestamp"`
}

// SearchQueueMessages searches queue messages based on criteria
func (s *Service) SearchQueueMessages(criteria *SearchCriteria) ([]QueueMessage, error) {
	status, err := s.manager.ListQueue()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	var filtered []QueueMessage
	for _, msg := range status.Messages {
		if s.matchesCriteria(&msg, criteria) {
			filtered = append(filtered, msg)
		}
	}

	return filtered, nil
}

// SearchCriteria defines search parameters for queue messages
type SearchCriteria struct {
	Sender     string `json:"sender"`
	Recipient  string `json:"recipient"`
	MessageID  string `json:"message_id"`
	Status     string `json:"status"`
	Subject    string `json:"subject"`
	MinAge     string `json:"min_age"`
	MaxAge     string `json:"max_age"`
	MinSize    int64  `json:"min_size"`
	MaxSize    int64  `json:"max_size"`
	MinRetries int    `json:"min_retries"`
	MaxRetries int    `json:"max_retries"`
}

// matchesCriteria checks if a message matches the search criteria
func (s *Service) matchesCriteria(msg *QueueMessage, criteria *SearchCriteria) bool {
	if criteria.Sender != "" && !contains(msg.Sender, criteria.Sender) {
		return false
	}

	if criteria.Recipient != "" {
		found := false
		for _, recipient := range msg.Recipients {
			if contains(recipient, criteria.Recipient) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if criteria.MessageID != "" && !contains(msg.ID, criteria.MessageID) {
		return false
	}

	if criteria.Status != "" && msg.Status != criteria.Status {
		return false
	}

	if criteria.MinSize > 0 && msg.Size < criteria.MinSize {
		return false
	}

	if criteria.MaxSize > 0 && msg.Size > criteria.MaxSize {
		return false
	}

	if criteria.MinRetries > 0 && msg.RetryCount < criteria.MinRetries {
		return false
	}

	if criteria.MaxRetries > 0 && msg.RetryCount > criteria.MaxRetries {
		return false
	}

	// Subject filtering - requires getting message details
	if criteria.Subject != "" {
		details, err := s.manager.InspectMessage(msg.ID)
		if err != nil {
			// If we can't get details, skip subject filtering for this message
			log.Printf("Failed to get message details for subject filtering: %v", err)
		} else {
			subject := details.Headers["Subject"]
			if !contains(subject, criteria.Subject) {
				return false
			}
		}
	}

	// Age filtering - parse age strings and compare durations
	if criteria.MinAge != "" || criteria.MaxAge != "" {
		msgAge := s.parseAge(msg.Age)
		
		if criteria.MinAge != "" {
			minAge := s.parseDurationString(criteria.MinAge)
			if minAge > 0 && msgAge < minAge {
				return false
			}
		}
		
		if criteria.MaxAge != "" {
			maxAge := s.parseDurationString(criteria.MaxAge)
			if maxAge > 0 && msgAge > maxAge {
				return false
			}
		}
	}

	return true
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// parseAge converts age string from queue message to duration
func (s *Service) parseAge(ageStr string) time.Duration {
	ageStr = strings.TrimSpace(ageStr)
	if ageStr == "" {
		return 0
	}

	// Try parsing as Go duration first (e.g., "2h30m", "5m", "1h")
	if duration, err := time.ParseDuration(ageStr); err == nil {
		return duration
	}

	// Handle day format (e.g., "1d", "2d")
	if strings.HasSuffix(ageStr, "d") {
		dayStr := strings.TrimSuffix(ageStr, "d")
		if days, err := strconv.Atoi(dayStr); err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	// Handle other common formats
	if strings.HasSuffix(ageStr, "h") {
		hourStr := strings.TrimSuffix(ageStr, "h")
		if hours, err := strconv.Atoi(hourStr); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}

	if strings.HasSuffix(ageStr, "m") {
		minStr := strings.TrimSuffix(ageStr, "m")
		if mins, err := strconv.Atoi(minStr); err == nil {
			return time.Duration(mins) * time.Minute
		}
	}

	return 0
}

// parseDurationString converts user input duration string to time.Duration
func (s *Service) parseDurationString(durationStr string) time.Duration {
	durationStr = strings.TrimSpace(durationStr)
	if durationStr == "" {
		return 0
	}

	// Try parsing as Go duration first (e.g., "2h", "30m", "1h30m")
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration
	}

	// Handle day format
	if strings.HasSuffix(durationStr, "d") {
		dayStr := strings.TrimSuffix(durationStr, "d")
		if days, err := strconv.Atoi(dayStr); err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	// Handle numeric values with units
	if strings.HasSuffix(durationStr, "h") {
		hourStr := strings.TrimSuffix(durationStr, "h")
		if hours, err := strconv.Atoi(hourStr); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}

	if strings.HasSuffix(durationStr, "m") {
		minStr := strings.TrimSuffix(durationStr, "m")
		if mins, err := strconv.Atoi(minStr); err == nil {
			return time.Duration(mins) * time.Minute
		}
	}

	// Try parsing as plain number (assume minutes)
	if mins, err := strconv.Atoi(durationStr); err == nil {
		return time.Duration(mins) * time.Minute
	}

	return 0
}

// Queue Operations - delegate to manager

// DeliverNow forces immediate delivery of a message
func (s *Service) DeliverNow(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	return s.manager.DeliverNow(messageID, userID, ipAddress)
}

// FreezeMessage freezes a message
func (s *Service) FreezeMessage(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	return s.manager.FreezeMessage(messageID, userID, ipAddress)
}

// ThawMessage thaws a frozen message
func (s *Service) ThawMessage(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	return s.manager.ThawMessage(messageID, userID, ipAddress)
}

// DeleteMessage removes a message from the queue
func (s *Service) DeleteMessage(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	return s.manager.DeleteMessage(messageID, userID, ipAddress)
}

// BulkDeliverNow performs deliver now operation on multiple messages
func (s *Service) BulkDeliverNow(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return s.manager.BulkDeliverNow(messageIDs, userID, ipAddress)
}

// BulkFreeze performs freeze operation on multiple messages
func (s *Service) BulkFreeze(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return s.manager.BulkFreeze(messageIDs, userID, ipAddress)
}

// BulkThaw performs thaw operation on multiple messages
func (s *Service) BulkThaw(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return s.manager.BulkThaw(messageIDs, userID, ipAddress)
}

// BulkDelete performs delete operation on multiple messages
func (s *Service) BulkDelete(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return s.manager.BulkDelete(messageIDs, userID, ipAddress)
}

// GetOperationHistory retrieves the operation history for a message
func (s *Service) GetOperationHistory(messageID string) ([]database.AuditLog, error) {
	return s.manager.GetOperationHistory(messageID)
}

// GetRecentOperations retrieves recent queue operations
func (s *Service) GetRecentOperations(limit int) ([]database.AuditLog, error) {
	return s.manager.GetRecentOperations(limit)
}

// ValidateMessageID checks if a message ID is valid
func (s *Service) ValidateMessageID(messageID string) error {
	return s.manager.ValidateMessageID(messageID)
}

// GetQueueStatistics returns detailed queue statistics
func (s *Service) GetQueueStatistics() (*QueueStatistics, error) {
	status, err := s.manager.ListQueue()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	stats := &QueueStatistics{
		TotalMessages:    status.TotalMessages,
		DeferredMessages: status.DeferredMessages,
		FrozenMessages:   status.FrozenMessages,
		QueuedMessages:   0,
		TotalSize:        0,
		AverageSize:      0,
		OldestMessageAge: status.OldestMessageAge,
		StatusBreakdown:  make(map[string]int),
		SizeDistribution: make(map[string]int),
	}

	// Calculate statistics
	for _, msg := range status.Messages {
		stats.TotalSize += msg.Size
		stats.StatusBreakdown[msg.Status]++

		if msg.Status == "queued" {
			stats.QueuedMessages++
		}

		// Size distribution
		if msg.Size < 1024 {
			stats.SizeDistribution["<1KB"]++
		} else if msg.Size < 10*1024 {
			stats.SizeDistribution["1-10KB"]++
		} else if msg.Size < 100*1024 {
			stats.SizeDistribution["10-100KB"]++
		} else if msg.Size < 1024*1024 {
			stats.SizeDistribution["100KB-1MB"]++
		} else {
			stats.SizeDistribution[">1MB"]++
		}
	}

	if stats.TotalMessages > 0 {
		stats.AverageSize = stats.TotalSize / int64(stats.TotalMessages)
	}

	return stats, nil
}

// QueueStatistics represents detailed queue statistics
type QueueStatistics struct {
	TotalMessages    int            `json:"total_messages"`
	QueuedMessages   int            `json:"queued_messages"`
	DeferredMessages int            `json:"deferred_messages"`
	FrozenMessages   int            `json:"frozen_messages"`
	TotalSize        int64          `json:"total_size"`
	AverageSize      int64          `json:"average_size"`
	OldestMessageAge time.Duration  `json:"oldest_message_age"`
	StatusBreakdown  map[string]int `json:"status_breakdown"`
	SizeDistribution map[string]int `json:"size_distribution"`
}
