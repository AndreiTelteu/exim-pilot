package logprocessor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// StreamingProcessor handles efficient log processing with streaming
type StreamingProcessor struct {
	repository *database.Repository
	parser     *LogParser
	config     StreamingConfig
	mu         sync.RWMutex
	stats      ProcessingStats
}

// StreamingConfig holds configuration for streaming log processing
type StreamingConfig struct {
	BatchSize           int           `json:"batch_size"`
	FlushInterval       time.Duration `json:"flush_interval"`
	MaxMemoryUsage      int64         `json:"max_memory_usage"` // bytes
	BufferSize          int           `json:"buffer_size"`      // lines
	ConcurrentWorkers   int           `json:"concurrent_workers"`
	ProcessingTimeout   time.Duration `json:"processing_timeout"`
	EnableCompression   bool          `json:"enable_compression"`
	EnableDeduplication bool          `json:"enable_deduplication"`
}

// DefaultStreamingConfig returns default streaming configuration
func DefaultStreamingConfig() StreamingConfig {
	return StreamingConfig{
		BatchSize:           1000,
		FlushInterval:       5 * time.Second,
		MaxMemoryUsage:      100 * 1024 * 1024, // 100MB
		BufferSize:          10000,
		ConcurrentWorkers:   4,
		ProcessingTimeout:   30 * time.Second,
		EnableCompression:   true,
		EnableDeduplication: true,
	}
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor(repository *database.Repository, config StreamingConfig) *StreamingProcessor {
	return &StreamingProcessor{
		repository: repository,
		parser:     NewLogParser(),
		config:     config,
		stats:      ProcessingStats{StartTime: time.Now()},
	}
}

// ProcessLogFileStreaming processes a log file using streaming approach
func (sp *StreamingProcessor) ProcessLogFileStreaming(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}
	defer file.Close()

	return sp.ProcessReaderStreaming(ctx, file, filePath)
}

// ProcessReaderStreaming processes log entries from a reader using streaming approach
func (sp *StreamingProcessor) ProcessReaderStreaming(ctx context.Context, reader io.Reader, source string) error {
	sp.mu.Lock()
	sp.stats.FilesProcessed++
	sp.stats.CurrentFile = source
	sp.mu.Unlock()

	// Create buffered scanner for efficient reading
	scanner := bufio.NewScanner(reader)

	// Increase buffer size for large log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	// Create channels for pipeline processing
	linesChan := make(chan string, sp.config.BufferSize)
	entriesChan := make(chan *database.LogEntry, sp.config.BufferSize)
	batchChan := make(chan []*database.LogEntry, 10)

	// Start pipeline workers
	var wg sync.WaitGroup

	// Stage 1: Line reading
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(linesChan)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case linesChan <- scanner.Text():
				sp.incrementLinesRead()
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading from %s: %v", source, err)
		}
	}()

	// Stage 2: Parsing workers
	for i := 0; i < sp.config.ConcurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for line := range linesChan {
				select {
				case <-ctx.Done():
					return
				default:
					if entry := sp.parseLine(line); entry != nil {
						select {
						case entriesChan <- entry:
							sp.incrementLinesParsed()
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}(i)
	}

	// Stage 3: Batching
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(batchChan)

		batch := make([]*database.LogEntry, 0, sp.config.BatchSize)
		ticker := time.NewTicker(sp.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if len(batch) > 0 {
					select {
					case batchChan <- batch:
					case <-time.After(time.Second):
					}
				}
				return
			case entry, ok := <-entriesChan:
				if !ok {
					// Channel closed, flush remaining batch
					if len(batch) > 0 {
						select {
						case batchChan <- batch:
						case <-ctx.Done():
						}
					}
					return
				}

				batch = append(batch, entry)
				if len(batch) >= sp.config.BatchSize {
					select {
					case batchChan <- batch:
						batch = make([]*database.LogEntry, 0, sp.config.BatchSize)
					case <-ctx.Done():
						return
					}
				}
			case <-ticker.C:
				if len(batch) > 0 {
					select {
					case batchChan <- batch:
						batch = make([]*database.LogEntry, 0, sp.config.BatchSize)
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	// Close entriesChan when all parsing workers are done
	go func() {
		wg.Wait()
		close(entriesChan)
	}()

	// Stage 4: Database insertion
	insertWg := sync.WaitGroup{}
	insertWg.Add(1)
	go func() {
		defer insertWg.Done()

		for batch := range batchChan {
			if err := sp.processBatch(ctx, batch); err != nil {
				log.Printf("Failed to process batch: %v", err)
				sp.incrementErrors()
			} else {
				sp.incrementBatchesProcessed()
			}
		}
	}()

	// Wait for all processing to complete
	insertWg.Wait()

	sp.mu.Lock()
	sp.stats.LastProcessedFile = source
	sp.stats.LastProcessedTime = time.Now()
	sp.mu.Unlock()

	return nil
}

// processBatch processes a batch of log entries
func (sp *StreamingProcessor) processBatch(ctx context.Context, entries []*database.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Deduplicate entries if enabled
	if sp.config.EnableDeduplication {
		entries = sp.deduplicateEntries(entries)
	}

	// Use transaction for batch insert
	tx, err := sp.repository.GetDB().BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for efficient batch insert
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO log_entries (
			timestamp, message_id, log_type, event, host, sender, 
			recipients, size, status, error_code, error_text, raw_line
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert entries in batch
	for _, entry := range entries {
		var recipients *string
		if len(entry.Recipients) > 0 {
			recipientsJSON := fmt.Sprintf(`["%s"]`, entry.Recipients[0])
			if len(entry.Recipients) > 1 {
				// Simple JSON array construction for multiple recipients
				recipientsJSON = `["` + entry.Recipients[0]
				for _, r := range entry.Recipients[1:] {
					recipientsJSON += `","` + r
				}
				recipientsJSON += `"]`
			}
			recipients = &recipientsJSON
		}

		_, err := stmt.ExecContext(ctx,
			entry.Timestamp,
			entry.MessageID,
			entry.LogType,
			entry.Event,
			entry.Host,
			entry.Sender,
			recipients,
			entry.Size,
			entry.Status,
			entry.ErrorCode,
			entry.ErrorText,
			entry.RawLine,
		)
		if err != nil {
			log.Printf("Failed to insert log entry: %v", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	sp.incrementEntriesStored(len(entries))
	return nil
}

// parseLine parses a single log line
func (sp *StreamingProcessor) parseLine(line string) *database.LogEntry {
	entry, err := sp.parser.ParseLogLine(line)
	if err != nil {
		// Log parsing errors at debug level to avoid spam
		return nil
	}
	return entry
}

// deduplicateEntries removes duplicate entries based on raw line content
func (sp *StreamingProcessor) deduplicateEntries(entries []*database.LogEntry) []*database.LogEntry {
	seen := make(map[string]bool)
	deduplicated := make([]*database.LogEntry, 0, len(entries))

	for _, entry := range entries {
		if !seen[entry.RawLine] {
			seen[entry.RawLine] = true
			deduplicated = append(deduplicated, entry)
		}
	}

	return deduplicated
}

// GetProcessingStats returns current processing statistics
func (sp *StreamingProcessor) GetProcessingStats() ProcessingStats {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	stats := sp.stats
	stats.Duration = time.Since(stats.StartTime)

	if stats.Duration > 0 {
		stats.LinesPerSecond = float64(stats.LinesRead) / stats.Duration.Seconds()
		stats.EntriesPerSecond = float64(stats.EntriesStored) / stats.Duration.Seconds()
	}

	return stats
}

// Helper methods for thread-safe stats updates
func (sp *StreamingProcessor) incrementLinesRead() {
	sp.mu.Lock()
	sp.stats.LinesRead++
	sp.mu.Unlock()
}

func (sp *StreamingProcessor) incrementLinesParsed() {
	sp.mu.Lock()
	sp.stats.LinesParsed++
	sp.mu.Unlock()
}

func (sp *StreamingProcessor) incrementEntriesStored(count int) {
	sp.mu.Lock()
	sp.stats.EntriesStored += int64(count)
	sp.mu.Unlock()
}

func (sp *StreamingProcessor) incrementBatchesProcessed() {
	sp.mu.Lock()
	sp.stats.BatchesProcessed++
	sp.mu.Unlock()
}

func (sp *StreamingProcessor) incrementErrors() {
	sp.mu.Lock()
	sp.stats.Errors++
	sp.mu.Unlock()
}

// ProcessingStats represents processing statistics
type ProcessingStats struct {
	StartTime         time.Time     `json:"start_time"`
	Duration          time.Duration `json:"duration"`
	FilesProcessed    int64         `json:"files_processed"`
	LinesRead         int64         `json:"lines_read"`
	LinesParsed       int64         `json:"lines_parsed"`
	EntriesStored     int64         `json:"entries_stored"`
	BatchesProcessed  int64         `json:"batches_processed"`
	Errors            int64         `json:"errors"`
	LinesPerSecond    float64       `json:"lines_per_second"`
	EntriesPerSecond  float64       `json:"entries_per_second"`
	CurrentFile       string        `json:"current_file"`
	LastProcessedFile string        `json:"last_processed_file"`
	LastProcessedTime time.Time     `json:"last_processed_time"`
}

// MemoryMonitor monitors memory usage during processing
type MemoryMonitor struct {
	maxMemory int64
	mu        sync.RWMutex
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(maxMemory int64) *MemoryMonitor {
	return &MemoryMonitor{
		maxMemory: maxMemory,
	}
}

// CheckMemoryUsage checks if memory usage is within limits
func (mm *MemoryMonitor) CheckMemoryUsage() bool {
	// In a production system, you would implement actual memory monitoring
	// For now, we'll assume memory usage is within limits
	return true
}
