package logmonitor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/parser"
	"github.com/andreitelteu/exim-pilot/internal/security"
	"github.com/fsnotify/fsnotify"
)

// LogMonitor monitors Exim log files for changes and processes new entries
type LogMonitor struct {
	watcher         *fsnotify.Watcher
	parser          *parser.EximParser
	repository      *database.Repository
	logProcessor    LogProcessor
	logPaths        []string
	fileStates      map[string]*FileState
	securityService *security.Service
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	done            chan struct{}
}

// FileState tracks the state of a monitored log file
type FileState struct {
	Path     string
	Size     int64
	ModTime  time.Time
	Position int64
	File     *os.File
}

// Config holds configuration for the log monitor
type Config struct {
	LogPaths   []string
	Repository *database.Repository
}

// NewLogMonitor creates a new log monitor instance
func NewLogMonitor(config Config) (*LogMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	monitor := &LogMonitor{
		watcher:         watcher,
		parser:          parser.NewEximParser(),
		repository:      config.Repository,
		logPaths:        config.LogPaths,
		fileStates:      make(map[string]*FileState),
		securityService: security.NewService(),
		ctx:             ctx,
		cancel:          cancel,
		done:            make(chan struct{}),
	}

	return monitor, nil
}

// Start begins monitoring log files
func (m *LogMonitor) Start() error {
	// Validate log paths
	if err := m.validateLogPaths(); err != nil {
		return fmt.Errorf("log path validation failed: %w", err)
	}

	// Initialize file states and add watchers
	successCount := 0
	for _, logPath := range m.logPaths {
		if err := m.addLogFile(logPath); err != nil {
			log.Printf("Warning: failed to add log file %s: %v", logPath, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("no log files could be monitored")
	}

	// Start the monitoring goroutine
	go m.monitorLoop()

	log.Printf("Log monitor started, watching %d files", len(m.fileStates))
	return nil
}

// validateLogPaths validates that log paths are reasonable
func (m *LogMonitor) validateLogPaths() error {
	if len(m.logPaths) == 0 {
		return fmt.Errorf("no log paths configured")
	}

	for _, logPath := range m.logPaths {
		if logPath == "" {
			return fmt.Errorf("empty log path found")
		}

		// Check if the directory exists
		dir := filepath.Dir(logPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("log directory does not exist: %s", dir)
		}

		// Check if we have read permissions on the directory
		if file, err := os.Open(dir); err != nil {
			return fmt.Errorf("cannot access log directory %s: %w", dir, err)
		} else {
			file.Close()
		}
	}

	return nil
}

// Stop stops the log monitor
func (m *LogMonitor) Stop() error {
	m.cancel()

	// Close all open files
	m.mu.Lock()
	for _, state := range m.fileStates {
		if state.File != nil {
			state.File.Close()
		}
	}
	m.mu.Unlock()

	// Close the watcher
	if err := m.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	// Wait for monitoring loop to finish
	<-m.done

	log.Println("Log monitor stopped")
	return nil
}

// addLogFile adds a log file to the monitor
func (m *LogMonitor) addLogFile(logPath string) error {
	// Validate file access with security service
	if err := m.securityService.ValidateFileAccess(logPath, security.AccessRead); err != nil {
		log.Printf("SECURITY: File access denied for %s: %v", logPath, err)
		return fmt.Errorf("security validation failed for log file %s: %w", logPath, err)
	}

	// Check if file exists
	info, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Log file %s does not exist, will monitor for creation", logPath)
			// Watch the directory for file creation (with security validation)
			dir := filepath.Dir(logPath)
			if err := m.securityService.ValidateFileAccess(dir, security.AccessRead); err != nil {
				log.Printf("SECURITY: Directory access denied for %s: %v", dir, err)
				return fmt.Errorf("security validation failed for log directory %s: %w", dir, err)
			}
			return m.watcher.Add(dir)
		}
		return fmt.Errorf("failed to stat log file %s: %w", logPath, err)
	}

	// Open file for reading
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	// Seek to end of file to only process new entries
	position, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to seek to end of file %s: %w", logPath, err)
	}

	// Create file state
	m.mu.Lock()
	m.fileStates[logPath] = &FileState{
		Path:     logPath,
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		Position: position,
		File:     file,
	}
	m.mu.Unlock()

	// Add to watcher
	if err := m.watcher.Add(logPath); err != nil {
		file.Close()
		return fmt.Errorf("failed to add file to watcher %s: %w", logPath, err)
	}

	log.Printf("Added log file to monitor: %s (size: %d, position: %d)", logPath, info.Size(), position)
	return nil
}

// monitorLoop is the main monitoring loop
func (m *LogMonitor) monitorLoop() {
	defer close(m.done)

	for {
		select {
		case <-m.ctx.Done():
			return

		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			m.handleFileEvent(event)

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

// handleFileEvent processes file system events
func (m *LogMonitor) handleFileEvent(event fsnotify.Event) {
	// Check if this is a log file we're monitoring
	isMonitoredFile := false
	for _, logPath := range m.logPaths {
		if event.Name == logPath || strings.HasSuffix(event.Name, filepath.Base(logPath)) {
			isMonitoredFile = true
			break
		}
	}

	if !isMonitoredFile {
		return
	}

	switch {
	case event.Has(fsnotify.Write):
		m.handleFileWrite(event.Name)
	case event.Has(fsnotify.Create):
		m.handleFileCreate(event.Name)
	case event.Has(fsnotify.Remove):
		m.handleFileRemove(event.Name)
	case event.Has(fsnotify.Rename):
		m.handleFileRename(event.Name)
	}
}

// handleFileWrite processes file write events
func (m *LogMonitor) handleFileWrite(filePath string) {
	m.mu.RLock()
	state, exists := m.fileStates[filePath]
	m.mu.RUnlock()

	if !exists {
		// File not being monitored, try to add it
		if err := m.addLogFile(filePath); err != nil {
			log.Printf("Failed to add new log file %s: %v", filePath, err)
		}
		return
	}

	// Check if file was truncated (log rotation)
	info, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to stat file %s: %v", filePath, err)
		return
	}

	if info.Size() < state.Size {
		log.Printf("Log rotation detected for %s, reopening file", filePath)
		m.handleLogRotation(filePath)
		return
	}

	// Process new content
	if err := m.processNewContent(state); err != nil {
		log.Printf("Failed to process new content in %s: %v", filePath, err)
	}
}

// handleFileCreate processes file creation events
func (m *LogMonitor) handleFileCreate(filePath string) {
	// Check if this is a log file we should monitor
	for _, logPath := range m.logPaths {
		if filePath == logPath {
			log.Printf("Log file created: %s", filePath)
			if err := m.addLogFile(filePath); err != nil {
				log.Printf("Failed to add created log file %s: %v", filePath, err)
			}
			break
		}
	}
}

// handleFileRemove processes file removal events
func (m *LogMonitor) handleFileRemove(filePath string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if state, exists := m.fileStates[filePath]; exists {
		log.Printf("Log file removed: %s", filePath)
		if state.File != nil {
			state.File.Close()
		}
		delete(m.fileStates, filePath)
	}
}

// handleFileRename processes file rename events (log rotation)
func (m *LogMonitor) handleFileRename(filePath string) {
	log.Printf("Log file renamed/rotated: %s", filePath)
	m.handleLogRotation(filePath)
}

// handleLogRotation handles log file rotation
func (m *LogMonitor) handleLogRotation(filePath string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.fileStates[filePath]
	if !exists {
		return
	}

	// Close the old file
	if state.File != nil {
		state.File.Close()
	}

	// Check for rotated files (e.g., mainlog.1, mainlog.2.gz)
	m.processRotatedFiles(filePath)

	// Try to open the new file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to reopen rotated log file %s: %v", filePath, err)
		// Don't delete the state, the file might be recreated
		state.File = nil
		return
	}

	// Update state
	info, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to stat reopened log file %s: %v", filePath, err)
		file.Close()
		state.File = nil
		return
	}

	state.File = file
	state.Size = info.Size()
	state.ModTime = info.ModTime()
	state.Position = 0

	log.Printf("Reopened rotated log file: %s (new size: %d)", filePath, info.Size())
}

// processRotatedFiles checks for and processes rotated log files
func (m *LogMonitor) processRotatedFiles(originalPath string) {
	dir := filepath.Dir(originalPath)
	baseName := filepath.Base(originalPath)

	// Look for rotated files like mainlog.1, mainlog.2, etc.
	for i := 1; i <= 5; i++ { // Check up to 5 rotated files
		rotatedPath := filepath.Join(dir, fmt.Sprintf("%s.%d", baseName, i))
		if _, err := os.Stat(rotatedPath); err == nil {
			log.Printf("Found rotated log file: %s", rotatedPath)
			// Process the rotated file if we haven't seen it before
			// This is a simplified approach - in production you might want to track processed files
		}
	}
}

// processNewContent reads and processes new content from a log file
func (m *LogMonitor) processNewContent(state *FileState) error {
	if state.File == nil {
		return fmt.Errorf("file handle is nil for %s", state.Path)
	}

	// Seek to last known position
	if _, err := state.File.Seek(state.Position, io.SeekStart); err != nil {
		// If seek fails, try to reopen the file
		log.Printf("Seek failed for %s, attempting to reopen: %v", state.Path, err)
		return m.reopenFile(state)
	}

	scanner := bufio.NewScanner(state.File)
	// Increase buffer size for large log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineCount := 0
	errorCount := 0
	logType := m.getLogType(state.Path)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the log line
		logEntry, err := m.parser.ParseLogLine(line, logType)
		if err != nil {
			errorCount++
			if errorCount <= 5 { // Only log first few errors to avoid spam
				log.Printf("Failed to parse log line from %s: %v", state.Path, err)
			}
			continue
		}

		if logEntry != nil {
			// Use log processor if available, otherwise store directly
			if m.logProcessor != nil {
				if err := m.logProcessor.ProcessLogEntry(m.ctx, logEntry); err != nil {
					log.Printf("Failed to process log entry: %v", err)
					continue
				}
			} else {
				// Store in database with retry logic
				if err := m.storeLogEntryWithRetry(logEntry, 3); err != nil {
					log.Printf("Failed to store log entry after retries: %v", err)
					continue
				}
			}
		}

		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Update position
	newPosition, err := state.File.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("failed to get current position: %w", err)
	}

	state.Position = newPosition

	// Update file size and modification time
	if info, err := os.Stat(state.Path); err == nil {
		state.Size = info.Size()
		state.ModTime = info.ModTime()
	}

	if lineCount > 0 {
		log.Printf("Processed %d new log lines from %s (parse errors: %d)", lineCount, state.Path, errorCount)
	}

	return nil
}

// reopenFile attempts to reopen a file after an error
func (m *LogMonitor) reopenFile(state *FileState) error {
	if state.File != nil {
		state.File.Close()
	}

	file, err := os.Open(state.Path)
	if err != nil {
		return fmt.Errorf("failed to reopen file %s: %w", state.Path, err)
	}

	state.File = file
	state.Position = 0 // Start from beginning after reopen

	log.Printf("Successfully reopened file: %s", state.Path)
	return nil
}

// storeLogEntryWithRetry attempts to store a log entry with retry logic
func (m *LogMonitor) storeLogEntryWithRetry(entry *database.LogEntry, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := m.repository.CreateLogEntry(m.ctx, entry); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				// Wait a bit before retrying
				time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// SetLogProcessor sets a log processor for enhanced processing
func (m *LogMonitor) SetLogProcessor(processor LogProcessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logProcessor = processor
}

// LogProcessor interface for enhanced log processing
type LogProcessor interface {
	ProcessLogEntry(ctx context.Context, entry *database.LogEntry) error
	ProcessLogEntries(ctx context.Context, entries []*database.LogEntry) error
}

// getLogType determines the log type based on file path
func (m *LogMonitor) getLogType(filePath string) string {
	fileName := filepath.Base(filePath)

	switch {
	case strings.Contains(fileName, "reject"):
		return database.LogTypeReject
	case strings.Contains(fileName, "panic"):
		return database.LogTypePanic
	default:
		return database.LogTypeMain
	}
}

// ProcessHistoricalLogs processes existing log files from the beginning
func (m *LogMonitor) ProcessHistoricalLogs() error {
	log.Println("Starting historical log processing...")

	for _, logPath := range m.logPaths {
		if err := m.processHistoricalLogFile(logPath); err != nil {
			log.Printf("Failed to process historical log file %s: %v", logPath, err)
			continue
		}
	}

	log.Println("Historical log processing completed")
	return nil
}

// processHistoricalLogFile processes a single log file from the beginning
func (m *LogMonitor) processHistoricalLogFile(logPath string) error {
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Historical log file %s does not exist, skipping", logPath)
			return nil
		}
		return fmt.Errorf("failed to open historical log file %s: %w", logPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineCount := 0
	errorCount := 0
	logType := m.getLogType(logPath)
	batchSize := 100
	var logEntries []*database.LogEntry

	log.Printf("Processing historical log file: %s (type: %s)", logPath, logType)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the log line
		logEntry, err := m.parser.ParseLogLine(line, logType)
		if err != nil {
			errorCount++
			if errorCount%100 == 0 {
				log.Printf("Parse errors in %s: %d (latest: %v)", logPath, errorCount, err)
			}
			continue
		}

		if logEntry != nil {
			logEntries = append(logEntries, logEntry)
		}

		lineCount++

		// Process in batches to improve performance
		if len(logEntries) >= batchSize {
			if err := m.processBatchLogEntries(logEntries); err != nil {
				log.Printf("Failed to process batch of log entries: %v", err)
			}
			logEntries = logEntries[:0] // Reset slice
		}

		// Progress reporting
		if lineCount%1000 == 0 {
			log.Printf("Processed %d lines from %s (errors: %d)", lineCount, logPath, errorCount)
		}
	}

	// Process remaining entries
	if len(logEntries) > 0 {
		if err := m.processBatchLogEntries(logEntries); err != nil {
			log.Printf("Failed to process final batch of log entries: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading historical log file: %w", err)
	}

	log.Printf("Completed processing %d lines from historical log file: %s (errors: %d)", lineCount, logPath, errorCount)
	return nil
}

// processBatchLogEntries processes a batch of log entries efficiently
func (m *LogMonitor) processBatchLogEntries(entries []*database.LogEntry) error {
	for _, entry := range entries {
		if err := m.repository.CreateLogEntry(m.ctx, entry); err != nil {
			return fmt.Errorf("failed to store log entry: %w", err)
		}
	}
	return nil
}
