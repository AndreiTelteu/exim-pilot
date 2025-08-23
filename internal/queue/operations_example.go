package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// ExampleQueueOperations demonstrates how to use the queue operations
func ExampleQueueOperations() {
	// This is an example of how the queue operations would be used
	// In a real application, this would be called from API handlers

	// Initialize database connection (example)
	dbConfig := database.DefaultConfig()
	dbConfig.Path = "exim-pilot.db"
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create queue service
	service := NewService("/usr/sbin/exim4", db)

	// Example user context
	userID := "admin"
	ipAddress := "192.168.1.100"

	// Example 1: Get queue status
	fmt.Println("=== Queue Status ===")
	status, err := service.GetQueueStatus()
	if err != nil {
		log.Printf("Failed to get queue status: %v", err)
		return
	}

	fmt.Printf("Total messages: %d\n", status.TotalMessages)
	fmt.Printf("Deferred messages: %d\n", status.DeferredMessages)
	fmt.Printf("Frozen messages: %d\n", status.FrozenMessages)

	// Example 2: Single message operations
	if len(status.Messages) > 0 {
		messageID := status.Messages[0].ID
		fmt.Printf("\n=== Operating on message: %s ===\n", messageID)

		// Inspect message details
		details, err := service.GetMessageDetails(messageID)
		if err != nil {
			log.Printf("Failed to get message details: %v", err)
		} else {
			fmt.Printf("Message size: %d bytes\n", details.Size)
			fmt.Printf("Sender: %s\n", details.Sender)
			fmt.Printf("Recipients: %v\n", details.Recipients)
		}

		// Freeze the message
		result, err := service.FreezeMessage(messageID, userID, ipAddress)
		if err != nil {
			log.Printf("Failed to freeze message: %v", err)
		} else {
			fmt.Printf("Freeze result: %s (success: %t)\n", result.Message, result.Success)
		}

		// Thaw the message
		result, err = service.ThawMessage(messageID, userID, ipAddress)
		if err != nil {
			log.Printf("Failed to thaw message: %v", err)
		} else {
			fmt.Printf("Thaw result: %s (success: %t)\n", result.Message, result.Success)
		}

		// Force delivery
		result, err = service.DeliverNow(messageID, userID, ipAddress)
		if err != nil {
			log.Printf("Failed to deliver message: %v", err)
		} else {
			fmt.Printf("Delivery result: %s (success: %t)\n", result.Message, result.Success)
		}
	}

	// Example 3: Bulk operations
	if len(status.Messages) >= 2 {
		messageIDs := []string{status.Messages[0].ID, status.Messages[1].ID}
		fmt.Printf("\n=== Bulk Operations on messages: %v ===\n", messageIDs)

		// Bulk freeze
		bulkResult, err := service.BulkFreeze(messageIDs, userID, ipAddress)
		if err != nil {
			log.Printf("Failed to bulk freeze: %v", err)
		} else {
			fmt.Printf("Bulk freeze: %d successful, %d failed\n",
				bulkResult.SuccessfulCount, bulkResult.FailedCount)
		}

		// Bulk thaw
		bulkResult, err = service.BulkThaw(messageIDs, userID, ipAddress)
		if err != nil {
			log.Printf("Failed to bulk thaw: %v", err)
		} else {
			fmt.Printf("Bulk thaw: %d successful, %d failed\n",
				bulkResult.SuccessfulCount, bulkResult.FailedCount)
		}
	}

	// Example 4: Search functionality
	fmt.Println("\n=== Search Operations ===")
	searchCriteria := &SearchCriteria{
		Status:  "deferred",
		MinSize: 1000,
	}

	searchResults, err := service.SearchQueueMessages(searchCriteria)
	if err != nil {
		log.Printf("Failed to search messages: %v", err)
	} else {
		fmt.Printf("Found %d deferred messages over 1KB\n", len(searchResults))
	}

	// Example 5: Queue health monitoring
	fmt.Println("\n=== Queue Health ===")
	health, err := service.GetQueueHealth()
	if err != nil {
		log.Printf("Failed to get queue health: %v", err)
	} else {
		fmt.Printf("Queue health - Total: %d, Deferred: %d, Frozen: %d\n",
			health.TotalMessages, health.DeferredMessages, health.FrozenMessages)
		fmt.Printf("Oldest message age: %v\n", health.OldestMessageAge)
		fmt.Printf("Growth trend: %d messages per interval\n", health.GrowthTrend)
	}

	// Example 6: Operation history
	if len(status.Messages) > 0 {
		messageID := status.Messages[0].ID
		fmt.Printf("\n=== Operation History for %s ===\n", messageID)

		history, err := service.GetOperationHistory(messageID)
		if err != nil {
			log.Printf("Failed to get operation history: %v", err)
		} else {
			fmt.Printf("Found %d operations in history\n", len(history))
			for _, op := range history {
				fmt.Printf("- %s: %s at %v\n", op.Action,
					getStringValue(op.Details), op.Timestamp)
			}
		}
	}

	// Example 7: Recent operations
	fmt.Println("\n=== Recent Operations ===")
	recentOps, err := service.GetRecentOperations(10)
	if err != nil {
		log.Printf("Failed to get recent operations: %v", err)
	} else {
		fmt.Printf("Found %d recent operations\n", len(recentOps))
		for _, op := range recentOps {
			fmt.Printf("- %s by %s at %v\n", op.Action,
				getStringValue(op.UserID), op.Timestamp)
		}
	}

	// Example 8: Queue statistics
	fmt.Println("\n=== Queue Statistics ===")
	stats, err := service.GetQueueStatistics()
	if err != nil {
		log.Printf("Failed to get queue statistics: %v", err)
	} else {
		fmt.Printf("Total size: %d bytes (avg: %d bytes per message)\n",
			stats.TotalSize, stats.AverageSize)
		fmt.Printf("Status breakdown: %v\n", stats.StatusBreakdown)
		fmt.Printf("Size distribution: %v\n", stats.SizeDistribution)
	}

	// Example 9: Periodic snapshots (would run in background)
	fmt.Println("\n=== Starting Periodic Snapshots ===")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go service.StartPeriodicSnapshots(ctx, 1*time.Minute)
	fmt.Println("Periodic snapshots started (would run in background)")
}

// Helper function to safely get string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return "unknown"
	}
	return *ptr
}

// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	// Example of proper error handling for queue operations

	service := &Service{} // Uninitialized service for demonstration

	// Validate message ID before operations
	messageID := "invalid-id"
	if err := service.ValidateMessageID(messageID); err != nil {
		fmt.Printf("Invalid message ID: %v\n", err)
		return
	}

	// Handle operation failures gracefully
	result, err := service.DeliverNow(messageID, "user", "127.0.0.1")
	if err != nil {
		fmt.Printf("Operation failed: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("Operation unsuccessful: %s\n", result.Error)
		// Could implement retry logic here
		return
	}

	fmt.Printf("Operation successful: %s\n", result.Message)
}

// ExampleBulkOperationHandling demonstrates bulk operation patterns
func ExampleBulkOperationHandling() {
	service := &Service{} // Example service

	messageIDs := []string{"msg1", "msg2", "msg3"}
	userID := "admin"
	ipAddress := "192.168.1.1"

	// Perform bulk operation
	result, err := service.BulkDelete(messageIDs, userID, ipAddress)
	if err != nil {
		fmt.Printf("Bulk operation failed: %v\n", err)
		return
	}

	// Analyze results
	fmt.Printf("Bulk delete completed: %d/%d successful\n",
		result.SuccessfulCount, result.TotalMessages)

	// Handle partial failures
	if result.FailedCount > 0 {
		fmt.Println("Failed operations:")
		for _, opResult := range result.Results {
			if !opResult.Success {
				fmt.Printf("- %s: %s\n", opResult.MessageID, opResult.Error)
			}
		}
	}

	// Log successful operations
	if result.SuccessfulCount > 0 {
		fmt.Printf("Successfully processed %d messages\n", result.SuccessfulCount)
	}
}
