package logprocessor

import (
	"testing"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

func TestNewService(t *testing.T) {
	// Create a mock repository (in a real test, you'd use a test database)
	var repository *database.Repository // This would be initialized with a test DB

	config := DefaultServiceConfig()
	service := NewService(repository, config)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.config.DefaultSearchLimit != config.DefaultSearchLimit {
		t.Errorf("Expected default search limit %d, got %d",
			config.DefaultSearchLimit, service.config.DefaultSearchLimit)
	}
}

func TestServiceConfig(t *testing.T) {
	config := DefaultServiceConfig()

	// Test default values
	if config.DefaultSearchLimit != 100 {
		t.Errorf("Expected default search limit 100, got %d", config.DefaultSearchLimit)
	}

	if config.MaxSearchLimit != 1000 {
		t.Errorf("Expected max search limit 1000, got %d", config.MaxSearchLimit)
	}

	if config.BatchSize != 500 {
		t.Errorf("Expected batch size 500, got %d", config.BatchSize)
	}

	if !config.EnableCorrelation {
		t.Error("Expected correlation to be enabled by default")
	}

	if !config.EnableCleanup {
		t.Error("Expected cleanup to be enabled by default")
	}

	if !config.EnableMetrics {
		t.Error("Expected metrics to be enabled by default")
	}
}

func TestBackgroundConfig(t *testing.T) {
	config := DefaultBackgroundConfig()

	// Test default values
	if config.CorrelationInterval != 30*time.Minute {
		t.Errorf("Expected correlation interval 30m, got %v", config.CorrelationInterval)
	}

	if config.LogRetentionDays != 90 {
		t.Errorf("Expected log retention 90 days, got %d", config.LogRetentionDays)
	}

	if config.AuditRetentionDays != 365 {
		t.Errorf("Expected audit retention 365 days, got %d", config.AuditRetentionDays)
	}

	if config.CleanupInterval != 6*time.Hour {
		t.Errorf("Expected cleanup interval 6h, got %v", config.CleanupInterval)
	}
}

func TestSearchCriteria(t *testing.T) {
	criteria := SearchCriteria{
		MessageID: "test-message-id",
		Sender:    "test@example.com",
		LogTypes:  []string{"main", "reject"},
		Events:    []string{"arrival", "delivery"},
		Limit:     50,
		Offset:    0,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	if criteria.MessageID != "test-message-id" {
		t.Errorf("Expected message ID 'test-message-id', got '%s'", criteria.MessageID)
	}

	if len(criteria.LogTypes) != 2 {
		t.Errorf("Expected 2 log types, got %d", len(criteria.LogTypes))
	}

	if criteria.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", criteria.Limit)
	}
}

func TestMessageCorrelation(t *testing.T) {
	correlation := &MessageCorrelation{
		MessageID: "test-message-id",
		Timeline:  make([]TimelineEvent, 0),
		Summary: MessageSummary{
			FirstSeen:       time.Now().Add(-1 * time.Hour),
			LastActivity:    time.Now(),
			TotalRecipients: 2,
			DeliveredCount:  1,
			DeferredCount:   1,
			FinalStatus:     "deferred",
		},
	}

	if correlation.MessageID != "test-message-id" {
		t.Errorf("Expected message ID 'test-message-id', got '%s'", correlation.MessageID)
	}

	if correlation.Summary.TotalRecipients != 2 {
		t.Errorf("Expected 2 total recipients, got %d", correlation.Summary.TotalRecipients)
	}

	if correlation.Summary.FinalStatus != "deferred" {
		t.Errorf("Expected final status 'deferred', got '%s'", correlation.Summary.FinalStatus)
	}
}

func TestTimelineEvent(t *testing.T) {
	event := TimelineEvent{
		Timestamp:   time.Now(),
		Event:       "arrival",
		Description: "Message received from test@example.com",
		Status:      "received",
		Details:     "Size: 1024 bytes",
	}

	if event.Event != "arrival" {
		t.Errorf("Expected event 'arrival', got '%s'", event.Event)
	}

	if event.Status != "received" {
		t.Errorf("Expected status 'received', got '%s'", event.Status)
	}
}

func TestSearchAggregations(t *testing.T) {
	agg := &SearchAggregations{
		EventCounts:   map[string]int{"arrival": 10, "delivery": 8, "defer": 2},
		LogTypeCounts: map[string]int{"main": 18, "reject": 2},
		StatusCounts:  map[string]int{"delivered": 8, "deferred": 2},
		TopSenders: []SenderCount{
			{Sender: "user1@example.com", Count: 5},
			{Sender: "user2@example.com", Count: 3},
		},
		TopHosts: []HostCount{
			{Host: "mail.example.com", Count: 8},
			{Host: "backup.example.com", Count: 2},
		},
	}

	if agg.EventCounts["arrival"] != 10 {
		t.Errorf("Expected 10 arrival events, got %d", agg.EventCounts["arrival"])
	}

	if len(agg.TopSenders) != 2 {
		t.Errorf("Expected 2 top senders, got %d", len(agg.TopSenders))
	}

	if agg.TopSenders[0].Count != 5 {
		t.Errorf("Expected top sender count 5, got %d", agg.TopSenders[0].Count)
	}
}

func TestLogStatistics(t *testing.T) {
	stats := &LogStatistics{
		Period: Period{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
		TotalEntries: 100,
		ByLogType:    map[string]int{"main": 80, "reject": 15, "panic": 5},
		ByEvent:      map[string]int{"arrival": 30, "delivery": 25, "defer": 20, "bounce": 5},
	}

	if stats.TotalEntries != 100 {
		t.Errorf("Expected 100 total entries, got %d", stats.TotalEntries)
	}

	if stats.ByLogType["main"] != 80 {
		t.Errorf("Expected 80 main log entries, got %d", stats.ByLogType["main"])
	}

	if stats.ByEvent["arrival"] != 30 {
		t.Errorf("Expected 30 arrival events, got %d", stats.ByEvent["arrival"])
	}
}

func TestRetentionInfo(t *testing.T) {
	info := &RetentionInfo{
		Config: RetentionConfig{
			LogRetentionDays:      90,
			AuditRetentionDays:    365,
			SnapshotRetentionDays: 30,
		},
		CurrentCounts: map[string]int{
			"log_entries":     10000,
			"audit_log":       500,
			"queue_snapshots": 720,
		},
	}

	if info.Config.LogRetentionDays != 90 {
		t.Errorf("Expected log retention 90 days, got %d", info.Config.LogRetentionDays)
	}

	if info.CurrentCounts["log_entries"] != 10000 {
		t.Errorf("Expected 10000 log entries, got %d", info.CurrentCounts["log_entries"])
	}
}

// Benchmark tests for performance
func BenchmarkSearchCriteria(b *testing.B) {
	for i := 0; i < b.N; i++ {
		criteria := SearchCriteria{
			MessageID: "test-message-id",
			Sender:    "test@example.com",
			LogTypes:  []string{"main", "reject"},
			Events:    []string{"arrival", "delivery"},
			Keywords:  []string{"error", "timeout"},
			Limit:     100,
		}
		_ = criteria
	}
}

func BenchmarkMessageCorrelation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		correlation := &MessageCorrelation{
			MessageID: "test-message-id",
			Timeline:  make([]TimelineEvent, 10),
			Summary: MessageSummary{
				TotalRecipients: 5,
				DeliveredCount:  3,
				DeferredCount:   2,
				FinalStatus:     "delivered",
			},
		}
		_ = correlation
	}
}

// Helper functions for testing
func createTestLogEntry(messageID, event, logType string) *database.LogEntry {
	now := time.Now()
	return &database.LogEntry{
		Timestamp: now,
		MessageID: &messageID,
		LogType:   logType,
		Event:     event,
		CreatedAt: now,
	}
}

func createTestMessage(id, sender, status string) *database.Message {
	now := time.Now()
	return &database.Message{
		ID:        id,
		Timestamp: now,
		Sender:    sender,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func createTestRecipient(messageID, recipient, status string) *database.Recipient {
	now := time.Now()
	return &database.Recipient{
		MessageID: messageID,
		Recipient: recipient,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
