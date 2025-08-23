package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// EximParser parses Exim log entries into structured data
type EximParser struct {
	// Compiled regular expressions for different log patterns
	mainLogPatterns   []*LogPattern
	rejectLogPatterns []*LogPattern
	panicLogPatterns  []*LogPattern
}

// LogPattern represents a compiled regex pattern with its handler
type LogPattern struct {
	Regex   *regexp.Regexp
	Handler func(matches []string, timestamp time.Time, rawLine string) *database.LogEntry
}

// NewEximParser creates a new Exim log parser
func NewEximParser() *EximParser {
	parser := &EximParser{}
	parser.initializePatterns()
	return parser
}

// initializePatterns compiles all the regex patterns for different log types
func (p *EximParser) initializePatterns() {
	// Main log patterns
	p.mainLogPatterns = []*LogPattern{
		// Message arrival
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) ([A-Za-z0-9-]+) <= ([^\s]+) H=([^\s]+) \[([^\]]+)\].*?S=(\d+)`),
			Handler: p.handleMessageArrival,
		},
		// Message delivery
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) ([A-Za-z0-9-]+) => ([^\s]+) R=([^\s]+) T=([^\s]+) H=([^\s]+) \[([^\]]+)\]`),
			Handler: p.handleMessageDelivery,
		},
		// Message deferral
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) ([A-Za-z0-9-]+) == ([^\s]+) R=([^\s]+) T=([^\s]+) defer \(([^)]+)\): (.+)`),
			Handler: p.handleMessageDefer,
		},
		// Message bounce
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) ([A-Za-z0-9-]+) \*\* ([^\s]+) R=([^\s]+) T=([^\s]+): (.+)`),
			Handler: p.handleMessageBounce,
		},
		// Message completion
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) ([A-Za-z0-9-]+) Completed`),
			Handler: p.handleMessageCompleted,
		},
	}

	// Reject log patterns
	p.rejectLogPatterns = []*LogPattern{
		// Connection rejected
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) rejected connection from \[([^\]]+)\]: (.+)`),
			Handler: p.handleConnectionRejected,
		},
		// SMTP rejection
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) H=([^\s]+) \[([^\]]+)\] rejected ([A-Z]+) <([^>]+)>: (.+)`),
			Handler: p.handleSMTPRejected,
		},
	}

	// Panic log patterns
	p.panicLogPatterns = []*LogPattern{
		// General panic/error
		{
			Regex:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) exim: (panic|error): (.+)`),
			Handler: p.handlePanicError,
		},
	}
}

// ParseLogLine parses a single log line and returns a LogEntry
func (p *EximParser) ParseLogLine(line, logType string) (*database.LogEntry, error) {
	if strings.TrimSpace(line) == "" {
		return nil, nil
	}

	var patterns []*LogPattern
	switch logType {
	case database.LogTypeMain:
		patterns = p.mainLogPatterns
	case database.LogTypeReject:
		patterns = p.rejectLogPatterns
	case database.LogTypePanic:
		patterns = p.panicLogPatterns
	default:
		return nil, fmt.Errorf("unknown log type: %s", logType)
	}

	// Try each pattern until one matches
	for _, pattern := range patterns {
		if matches := pattern.Regex.FindStringSubmatch(line); matches != nil {
			// Parse timestamp from the first capture group
			timestamp, err := p.parseTimestamp(matches[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}

			// Call the pattern's handler
			entry := pattern.Handler(matches, timestamp, line)
			if entry != nil {
				entry.LogType = logType
				entry.RawLine = line
				entry.CreatedAt = time.Now()
			}
			return entry, nil
		}
	}

	// If no pattern matched, create a generic log entry
	return p.createGenericLogEntry(line, logType)
}

// parseTimestamp parses Exim timestamp format
func (p *EximParser) parseTimestamp(timestampStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timestampStr)
}

// createGenericLogEntry creates a generic log entry for unparsed lines
func (p *EximParser) createGenericLogEntry(line, logType string) (*database.LogEntry, error) {
	timestampRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)
	matches := timestampRegex.FindStringSubmatch(line)

	var timestamp time.Time
	var err error

	if len(matches) > 1 {
		timestamp, err = p.parseTimestamp(matches[1])
		if err != nil {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	return &database.LogEntry{
		Timestamp: timestamp,
		LogType:   logType,
		Event:     "unknown",
		RawLine:   line,
		CreatedAt: time.Now(),
	}, nil
}

// Handler functions for different log patterns

func (p *EximParser) handleMessageArrival(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	messageID := matches[2]
	sender := matches[3]
	host := matches[4]
	sizeStr := matches[6]

	size, _ := strconv.ParseInt(sizeStr, 10, 64)

	return &database.LogEntry{
		Timestamp: timestamp,
		MessageID: &messageID,
		Event:     database.EventArrival,
		Host:      &host,
		Sender:    &sender,
		Size:      &size,
		Status:    stringPtr("received"),
	}
}

func (p *EximParser) handleMessageDelivery(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	messageID := matches[2]
	recipient := matches[3]
	host := matches[6]

	return &database.LogEntry{
		Timestamp:  timestamp,
		MessageID:  &messageID,
		Event:      database.EventDelivery,
		Host:       &host,
		Recipients: []string{recipient},
		Status:     stringPtr("delivered"),
	}
}

func (p *EximParser) handleMessageDefer(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	messageID := matches[2]
	recipient := matches[3]
	errorCode := matches[6]
	errorText := matches[7]

	return &database.LogEntry{
		Timestamp:  timestamp,
		MessageID:  &messageID,
		Event:      database.EventDefer,
		Recipients: []string{recipient},
		Status:     stringPtr("deferred"),
		ErrorCode:  &errorCode,
		ErrorText:  &errorText,
	}
}

func (p *EximParser) handleMessageBounce(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	messageID := matches[2]
	recipient := matches[3]
	errorText := matches[6]

	return &database.LogEntry{
		Timestamp:  timestamp,
		MessageID:  &messageID,
		Event:      database.EventBounce,
		Recipients: []string{recipient},
		Status:     stringPtr("bounced"),
		ErrorText:  &errorText,
	}
}

func (p *EximParser) handleMessageCompleted(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	messageID := matches[2]

	return &database.LogEntry{
		Timestamp: timestamp,
		MessageID: &messageID,
		Event:     "completed",
		Status:    stringPtr("completed"),
	}
}

func (p *EximParser) handleConnectionRejected(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	ipAddress := matches[2]
	reason := matches[3]

	return &database.LogEntry{
		Timestamp: timestamp,
		Event:     database.EventReject,
		Host:      &ipAddress,
		Status:    stringPtr("rejected"),
		ErrorText: &reason,
	}
}

func (p *EximParser) handleSMTPRejected(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	host := matches[2]
	recipient := matches[5]
	reason := matches[6]

	return &database.LogEntry{
		Timestamp:  timestamp,
		Event:      database.EventReject,
		Host:       &host,
		Recipients: []string{recipient},
		Status:     stringPtr("rejected"),
		ErrorText:  &reason,
	}
}

func (p *EximParser) handlePanicError(matches []string, timestamp time.Time, rawLine string) *database.LogEntry {
	level := matches[2]
	message := matches[3]

	return &database.LogEntry{
		Timestamp: timestamp,
		Event:     database.EventPanic,
		Status:    &level,
		ErrorText: &message,
	}
}

// Helper functions

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ExtractMessageID extracts message ID from a log line if present
func (p *EximParser) ExtractMessageID(line string) string {
	// Exim message IDs have format: 1rABC-123456-78 (6 chars, dash, 6 chars, dash, 2 chars)
	messageIDRegex := regexp.MustCompile(`\b([A-Za-z0-9]{6}-[A-Za-z0-9]{6}-[A-Za-z0-9]{2})\b`)
	matches := messageIDRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
