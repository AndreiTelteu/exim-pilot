package performance

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
)

// BenchmarkDatabaseOperations benchmarks database operations with sample data
func BenchmarkDatabaseOperations(b *testing.B) {
	db, cleanup := setupTestDatabase(b)
	defer cleanup()

	repository := database.NewRepository(db)
	ctx := context.Background()

	// Generate sample data
	sampleEntries := generateSampleLogEntries(1000)

	b.ResetTimer()

	b.Run("InsertLogEntries", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entry := sampleEntries[i%len(sampleEntries)]
			if err := repository.CreateLogEntry(ctx, entry); err != nil {
				b.Fatalf("Failed to insert log entry: %v", err)
			}
		}
	})

	b.Run("QueryLogEntries", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			criteria := logprocessor.SearchCriteria{
				Limit:  100,
				Offset: 0,
			}
			_, err := repository.SearchLogEntries(ctx, criteria)
			if err != nil {
				b.Fatalf("Failed to query log entries: %v", err)
			}
		}
	})

	b.Run("QueryLogEntriesByMessageID", func(b *testing.B) {
		messageID := "test-message-1"
		for i := 0; i < b.N; i++ {
			criteria := logprocessor.SearchCriteria{
				MessageID: messageID,
				Limit:     100,
			}
			_, err := repository.SearchLogEntries(ctx, criteria)
			if err != nil {
				b.Fatalf("Failed to query log entries by message ID: %v", err)
			}
		}
	})

	b.Run("QueryLogEntriesByTimeRange", func(b *testing.B) {
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()

		for i := 0; i < b.N; i++ {
			criteria := logprocessor.SearchCriteria{
				StartTime: &startTime,
				EndTime:   &endTime,
				Limit:     100,
			}
			_, err := repository.SearchLogEntries(ctx, criteria)
			if err != nil {
				b.Fatalf("Failed to query log entries by time range: %v", err)
			}
		}
	})
}

// BenchmarkLogProcessing benchmarks log processing operations
func BenchmarkLogProcessing(b *testing.B) {
	db, cleanup := setupTestDatabase(b)
	defer cleanup()

	repository := database.NewRepository(db)
	service := logprocessor.NewService(repository, logprocessor.DefaultServiceConfig())

	// Generate sample log lines
	sampleLogLines := generateSampleLogLines(1000)

	b.ResetTimer()

	b.Run("ParseLogLines", func(b *testing.B) {
		parser := logprocessor.NewLogParser()

		for i := 0; i < b.N; i++ {
			line := sampleLogLines[i%len(sampleLogLines)]
			_, err := parser.ParseLogLine(line)
			if err != nil {
				// Parsing errors are expected for some lines
				continue
			}
		}
	})

	b.Run("ProcessLogEntries", func(b *testing.B) {
		ctx := context.Background()
		entries := generateSampleLogEntries(100)

		for i := 0; i < b.N; i++ {
			batch := entries[i%10 : (i%10)+10] // Process in batches of 10
			if err := service.ProcessLogEntries(ctx, batch); err != nil {
				b.Fatalf("Failed to process log entries: %v", err)
			}
		}
	})
}

// BenchmarkStreamingProcessor benchmarks streaming log processing
func BenchmarkStreamingProcessor(b *testing.B) {
	db, cleanup := setupTestDatabase(b)
	defer cleanup()

	repository := database.NewRepository(db)
	config := logprocessor.DefaultStreamingConfig()
	config.BatchSize = 100
	config.ConcurrentWorkers = 2

	processor := logprocessor.NewStreamingProcessor(repository, config)

	b.ResetTimer()

	b.Run("StreamingProcessing", func(b *testing.B) {
		ctx := context.Background()

		for i := 0; i < b.N; i++ {
			// Create a temporary file with sample log data
			tempFile := createTempLogFile(b, 1000)
			defer os.Remove(tempFile)

			if err := processor.ProcessLogFileStreaming(ctx, tempFile); err != nil {
				b.Fatalf("Failed to process log file: %v", err)
			}
		}
	})
}

// BenchmarkDatabaseOptimization benchmarks database optimization operations
func BenchmarkDatabaseOptimization(b *testing.B) {
	db, cleanup := setupTestDatabase(b)
	defer cleanup()

	// Insert sample data first
	repository := database.NewRepository(db)
	ctx := context.Background()

	entries := generateSampleLogEntries(10000)
	for _, entry := range entries {
		repository.CreateLogEntry(ctx, entry)
	}

	optimizationService := database.NewOptimizationService(db)

	b.ResetTimer()

	b.Run("DatabaseOptimization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := optimizationService.OptimizeDatabase(ctx); err != nil {
				b.Fatalf("Failed to optimize database: %v", err)
			}
		}
	})

	b.Run("GetDatabaseStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := optimizationService.GetDatabaseStats(ctx)
			if err != nil {
				b.Fatalf("Failed to get database stats: %v", err)
			}
		}
	})
}

// BenchmarkRetentionService benchmarks data retention operations
func BenchmarkRetentionService(b *testing.B) {
	db, cleanup := setupTestDatabase(b)
	defer cleanup()

	// Insert sample data with old timestamps
	repository := database.NewRepository(db)
	ctx := context.Background()

	// Insert old entries that should be cleaned up
	oldEntries := generateOldLogEntries(5000, 100) // 100 days old
	for _, entry := range oldEntries {
		repository.CreateLogEntry(ctx, entry)
	}

	config := database.DefaultRetentionConfig()
	config.LogEntriesRetentionDays = 90 // 90 days retention
	retentionService := database.NewRetentionService(db, config)

	b.ResetTimer()

	b.Run("CleanupExpiredData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := retentionService.CleanupExpiredData(ctx)
			if err != nil {
				b.Fatalf("Failed to cleanup expired data: %v", err)
			}
		}
	})

	b.Run("GetRetentionStatus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := retentionService.GetRetentionStatus(ctx)
			if err != nil {
				b.Fatalf("Failed to get retention status: %v", err)
			}
		}
	})
}

// TestLargeDatasetPerformance tests performance with large datasets
func TestLargeDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset performance test in short mode")
	}

	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repository := database.NewRepository(db)
	ctx := context.Background()

	// Test with different dataset sizes
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Dataset_%d", size), func(t *testing.T) {
			// Clear database
			db.Exec("DELETE FROM log_entries")

			// Insert test data
			entries := generateSampleLogEntries(size)
			start := time.Now()

			for _, entry := range entries {
				if err := repository.CreateLogEntry(ctx, entry); err != nil {
					t.Fatalf("Failed to insert entry: %v", err)
				}
			}

			insertDuration := time.Since(start)
			t.Logf("Inserted %d entries in %v (%.2f entries/sec)",
				size, insertDuration, float64(size)/insertDuration.Seconds())

			// Test query performance
			start = time.Now()
			criteria := logprocessor.SearchCriteria{
				Limit: 1000,
			}
			result, err := repository.SearchLogEntries(ctx, criteria)
			if err != nil {
				t.Fatalf("Failed to query entries: %v", err)
			}

			queryDuration := time.Since(start)
			t.Logf("Queried %d entries from %d total in %v",
				len(result.Entries), size, queryDuration)

			// Test filtered query performance
			start = time.Now()
			criteria.LogTypes = []string{"main"}
			result, err = repository.SearchLogEntries(ctx, criteria)
			if err != nil {
				t.Fatalf("Failed to query filtered entries: %v", err)
			}

			filteredQueryDuration := time.Since(start)
			t.Logf("Filtered query returned %d entries in %v",
				len(result.Entries), filteredQueryDuration)

			// Performance thresholds (adjust based on requirements)
			maxInsertTime := time.Duration(size) * time.Microsecond * 100 // 100μs per entry
			if insertDuration > maxInsertTime {
				t.Errorf("Insert performance too slow: %v > %v", insertDuration, maxInsertTime)
			}

			maxQueryTime := 100 * time.Millisecond
			if queryDuration > maxQueryTime {
				t.Errorf("Query performance too slow: %v > %v", queryDuration, maxQueryTime)
			}
		})
	}
}

// TestConcurrentOperations tests concurrent database operations
func TestConcurrentOperations(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repository := database.NewRepository(db)
	ctx := context.Background()

	// Test concurrent inserts
	t.Run("ConcurrentInserts", func(t *testing.T) {
		numWorkers := 10
		entriesPerWorker := 100

		done := make(chan error, numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				entries := generateSampleLogEntries(entriesPerWorker)
				for j, entry := range entries {
					entry.MessageID = stringPtr(fmt.Sprintf("worker-%d-msg-%d", workerID, j))
					if err := repository.CreateLogEntry(ctx, entry); err != nil {
						done <- fmt.Errorf("worker %d failed: %v", workerID, err)
						return
					}
				}
				done <- nil
			}(i)
		}

		// Wait for all workers to complete
		for i := 0; i < numWorkers; i++ {
			if err := <-done; err != nil {
				t.Fatalf("Concurrent insert failed: %v", err)
			}
		}

		// Verify total count
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM log_entries").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count entries: %v", err)
		}

		expected := numWorkers * entriesPerWorker
		if count != expected {
			t.Errorf("Expected %d entries, got %d", expected, count)
		}
	})

	// Test concurrent reads
	t.Run("ConcurrentReads", func(t *testing.T) {
		numWorkers := 20
		queriesPerWorker := 50

		done := make(chan error, numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				for j := 0; j < queriesPerWorker; j++ {
					criteria := logprocessor.SearchCriteria{
						Limit:  10,
						Offset: j * 10,
					}
					_, err := repository.SearchLogEntries(ctx, criteria)
					if err != nil {
						done <- fmt.Errorf("worker %d query %d failed: %v", workerID, j, err)
						return
					}
				}
				done <- nil
			}(i)
		}

		// Wait for all workers to complete
		for i := 0; i < numWorkers; i++ {
			if err := <-done; err != nil {
				t.Fatalf("Concurrent read failed: %v", err)
			}
		}
	})
}

// Helper functions

func setupTestDatabase(tb testing.TB) (*database.DB, func()) {
	dbPath := fmt.Sprintf("perf_test_%d.db", time.Now().UnixNano())

	config := &database.Config{
		Path:            dbPath,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := database.Connect(config)
	if err != nil {
		tb.Fatalf("Failed to connect to test database: %v", err)
	}

	// Initialize schema
	if _, err := db.Exec(database.Schema); err != nil {
		tb.Fatalf("Failed to initialize schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func generateSampleLogEntries(count int) []*database.LogEntry {
	entries := make([]*database.LogEntry, count)

	logTypes := []string{"main", "reject", "panic"}
	events := []string{"arrival", "delivery", "defer", "bounce"}
	statuses := []string{"received", "delivered", "deferred", "bounced"}

	for i := 0; i < count; i++ {
		entries[i] = &database.LogEntry{
			Timestamp:  time.Now().Add(-time.Duration(rand.Intn(86400)) * time.Second),
			MessageID:  stringPtr(fmt.Sprintf("test-message-%d", i)),
			LogType:    logTypes[rand.Intn(len(logTypes))],
			Event:      events[rand.Intn(len(events))],
			Host:       stringPtr("localhost"),
			Sender:     stringPtr(fmt.Sprintf("sender%d@example.com", rand.Intn(100))),
			Recipients: []string{fmt.Sprintf("recipient%d@example.com", rand.Intn(100))},
			Size:       int64Ptr(int64(rand.Intn(10000) + 1000)),
			Status:     stringPtr(statuses[rand.Intn(len(statuses))]),
			RawLine:    fmt.Sprintf("2024-01-01 12:00:00 sample log line %d", i),
		}
	}

	return entries
}

func generateOldLogEntries(count int, daysOld int) []*database.LogEntry {
	entries := generateSampleLogEntries(count)

	oldTime := time.Now().AddDate(0, 0, -daysOld)

	for i, entry := range entries {
		entry.Timestamp = oldTime.Add(-time.Duration(i) * time.Second)
	}

	return entries
}

func generateSampleLogLines(count int) []string {
	lines := make([]string, count)

	templates := []string{
		"2024-01-01 12:00:00 1a2b3c-4d5e6f-7g8h9i <= sender@example.com H=localhost [127.0.0.1] P=esmtp S=1024",
		"2024-01-01 12:01:00 1a2b3c-4d5e6f-7g8h9i => recipient@example.com R=local_delivery T=local_delivery",
		"2024-01-01 12:02:00 1a2b3c-4d5e6f-7g8h9i == recipient@example.com R=remote_smtp defer (-44): SMTP error",
		"2024-01-01 12:03:00 1a2b3c-4d5e6f-7g8h9i ** recipient@example.com: retry timeout exceeded",
	}

	for i := 0; i < count; i++ {
		template := templates[rand.Intn(len(templates))]
		lines[i] = fmt.Sprintf(template, i)
	}

	return lines
}

func createTempLogFile(tb testing.TB, lineCount int) string {
	file, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		tb.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	lines := generateSampleLogLines(lineCount)
	for _, line := range lines {
		fmt.Fprintln(file, line)
	}

	return file.Name()
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
