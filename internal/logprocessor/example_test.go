package logprocessor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// ExampleService demonstrates how to use the log processing service
func ExampleService() {
	// This example shows how to set up and use the log processing service

	// 1. Create database connection (in real code, you'd initialize this properly)
	var db *database.DB // This would be initialized with actual database
	repository := database.NewRepository(db)

	// 2. Create service with custom configuration
	config := ServiceConfig{
		BackgroundConfig: BackgroundConfig{
			CorrelationInterval: 15 * time.Minute, // Run correlation every 15 minutes
			LogRetentionDays:    60,               // Keep logs for 60 days
			AuditRetentionDays:  180,              // Keep audit logs for 180 days
			CleanupInterval:     4 * time.Hour,    // Run cleanup every 4 hours
		},
		DefaultSearchLimit: 50,
		MaxSearchLimit:     500,
		BatchSize:          200,
		EnableCorrelation:  true,
		EnableCleanup:      true,
		EnableMetrics:      true,
	}

	service := NewService(repository, config)

	// 3. Start the service
	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// 4. Process individual log entries
	ctx := context.Background()

	logEntry := &database.LogEntry{
		Timestamp:  time.Now(),
		MessageID:  stringPtr("1rABC-123456-78"),
		LogType:    database.LogTypeMain,
		Event:      database.EventArrival,
		Sender:     stringPtr("user@example.com"),
		Recipients: []string{"recipient@example.com"},
		Size:       int64Ptr(1024),
		Status:     stringPtr("received"),
		RawLine:    "2024-01-15 10:30:00 1rABC-123456-78 <= user@example.com H=mail.example.com [192.168.1.100] S=1024",
	}

	if err := service.ProcessLogEntry(ctx, logEntry); err != nil {
		log.Printf("Failed to process log entry: %v", err)
	}

	// 5. Search logs with various criteria
	searchCriteria := SearchCriteria{
		StartTime: timePtr(time.Now().Add(-24 * time.Hour)),
		EndTime:   timePtr(time.Now()),
		Sender:    "user@example.com",
		LogTypes:  []string{database.LogTypeMain},
		Events:    []string{database.EventArrival, database.EventDelivery},
		Limit:     20,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	results, err := service.SearchLogs(ctx, searchCriteria)
	if err != nil {
		log.Printf("Search failed: %v", err)
	} else {
		fmt.Printf("Found %d log entries (total: %d)\n", len(results.Entries), results.TotalCount)

		// Print aggregations
		if results.Aggregations != nil {
			fmt.Printf("Event counts: %+v\n", results.Aggregations.EventCounts)
			fmt.Printf("Top senders: %+v\n", results.Aggregations.TopSenders)
		}
	}

	// 6. Get message correlation data
	messageID := "1rABC-123456-78"
	correlation, err := service.GetMessageCorrelation(ctx, messageID)
	if err != nil {
		log.Printf("Failed to get correlation: %v", err)
	} else {
		fmt.Printf("Message %s: %d timeline events, status: %s\n",
			correlation.MessageID, len(correlation.Timeline), correlation.Summary.FinalStatus)
	}

	// 7. Get log statistics
	stats, err := service.GetLogStatistics(ctx, time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		log.Printf("Failed to get statistics: %v", err)
	} else {
		fmt.Printf("Statistics: %d total entries\n", stats.TotalEntries)
		fmt.Printf("By log type: %+v\n", stats.ByLogType)
		fmt.Printf("By event: %+v\n", stats.ByEvent)
	}

	// 8. Manually trigger correlation for recent data
	if err := service.TriggerCorrelation(ctx, time.Now().Add(-1*time.Hour), time.Now()); err != nil {
		log.Printf("Failed to trigger correlation: %v", err)
	}

	// 9. Get service status
	status := service.GetServiceStatus()
	fmt.Printf("Service running: %t\n", status.Running)
	fmt.Printf("Background service running: %t\n", status.BackgroundStatus.Running)

	// 10. Get retention information
	retentionInfo, err := service.GetRetentionInfo(ctx)
	if err != nil {
		log.Printf("Failed to get retention info: %v", err)
	} else {
		fmt.Printf("Retention config: %+v\n", retentionInfo.Config)
		fmt.Printf("Current counts: %+v\n", retentionInfo.CurrentCounts)
	}
}

// ExampleSearchCriteria demonstrates various search scenarios
func ExampleSearchCriteria() {
	// Search for messages from a specific sender in the last 24 hours
	criteria1 := SearchCriteria{
		StartTime: timePtr(time.Now().Add(-24 * time.Hour)),
		EndTime:   timePtr(time.Now()),
		Sender:    "important@example.com",
		Limit:     50,
	}

	// Search for delivery failures with error keywords
	criteria2 := SearchCriteria{
		Events:    []string{database.EventDefer, database.EventBounce},
		Keywords:  []string{"timeout", "connection refused"},
		Limit:     100,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	// Search for large messages
	criteria3 := SearchCriteria{
		MinSize:  int64Ptr(1024 * 1024), // 1MB
		LogTypes: []string{database.LogTypeMain},
		Events:   []string{database.EventArrival},
		Limit:    25,
	}

	// Search for messages to specific recipients
	criteria4 := SearchCriteria{
		Recipients: []string{"user1@domain.com", "user2@domain.com"},
		StartTime:  timePtr(time.Now().Add(-7 * 24 * time.Hour)), // Last week
		Limit:      200,
	}

	// Search for panic/error log entries
	criteria5 := SearchCriteria{
		LogTypes:  []string{database.LogTypePanic},
		Events:    []string{database.EventPanic},
		Limit:     10,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	fmt.Printf("Example search criteria created: %d scenarios\n", 5)
	_ = criteria1
	_ = criteria2
	_ = criteria3
	_ = criteria4
	_ = criteria5
}

// ExampleMessageCorrelation demonstrates message correlation features
func ExampleMessageCorrelation() {
	// Create a sample message correlation
	correlation := &MessageCorrelation{
		MessageID: "1rABC-123456-78",
		Message: &database.Message{
			ID:        "1rABC-123456-78",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Sender:    "sender@example.com",
			Status:    database.StatusDelivered,
		},
		Recipients: []database.Recipient{
			{
				MessageID: "1rABC-123456-78",
				Recipient: "user1@domain.com",
				Status:    database.RecipientStatusDelivered,
			},
			{
				MessageID: "1rABC-123456-78",
				Recipient: "user2@domain.com",
				Status:    database.RecipientStatusDeferred,
			},
		},
		Timeline: []TimelineEvent{
			{
				Timestamp:   time.Now().Add(-2 * time.Hour),
				Event:       database.EventArrival,
				Description: "Message received from sender@example.com",
				Status:      "received",
				Details:     "Size: 2048 bytes",
			},
			{
				Timestamp:   time.Now().Add(-90 * time.Minute),
				Event:       database.EventDelivery,
				Description: "Delivered to user1@domain.com",
				Status:      "delivered",
				Details:     "Host: mail.domain.com",
			},
			{
				Timestamp:   time.Now().Add(-85 * time.Minute),
				Event:       database.EventDefer,
				Description: "Deferred for user2@domain.com",
				Status:      "deferred",
				Details:     "Temporary failure: Connection timeout",
			},
		},
		Summary: MessageSummary{
			FirstSeen:       time.Now().Add(-2 * time.Hour),
			LastActivity:    time.Now().Add(-85 * time.Minute),
			TotalRecipients: 2,
			DeliveredCount:  1,
			DeferredCount:   1,
			AttemptCount:    3,
			FinalStatus:     "partially_delivered",
			Duration:        "35m0s",
		},
	}

	fmt.Printf("Message %s correlation:\n", correlation.MessageID)
	fmt.Printf("- Recipients: %d total, %d delivered, %d deferred\n",
		correlation.Summary.TotalRecipients,
		correlation.Summary.DeliveredCount,
		correlation.Summary.DeferredCount)
	fmt.Printf("- Timeline events: %d\n", len(correlation.Timeline))
	fmt.Printf("- Final status: %s\n", correlation.Summary.FinalStatus)
	fmt.Printf("- Duration: %s\n", correlation.Summary.Duration)
}

// ExampleBackgroundConfig demonstrates background processing configuration
func ExampleBackgroundConfig() {
	// Create background service configuration
	config := BackgroundConfig{
		// Run correlation every 30 minutes
		CorrelationInterval:   30 * time.Minute,
		CorrelationBatchHours: 24, // Process last 24 hours of data

		// Retention policies
		LogRetentionDays:      90,  // Keep logs for 90 days
		AuditRetentionDays:    365, // Keep audit logs for 1 year
		SnapshotRetentionDays: 30,  // Keep queue snapshots for 30 days

		// Cleanup runs every 6 hours
		CleanupInterval:  6 * time.Hour,
		CleanupBatchSize: 1000, // Delete up to 1000 records per batch

		// Performance settings
		MaxConcurrentTasks: 3, // Run up to 3 background tasks concurrently
	}

	fmt.Printf("Background processing configuration:\n")
	fmt.Printf("- Correlation interval: %v\n", config.CorrelationInterval)
	fmt.Printf("- Log retention: %d days\n", config.LogRetentionDays)
	fmt.Printf("- Audit retention: %d days\n", config.AuditRetentionDays)
	fmt.Printf("- Cleanup interval: %v\n", config.CleanupInterval)
	fmt.Printf("- Max concurrent tasks: %d\n", config.MaxConcurrentTasks)
}

// Helper functions for examples
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
