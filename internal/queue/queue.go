package queue

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/security"
)

// QueueMessage represents a message in the Exim queue
type QueueMessage struct {
	ID          string    `json:"id"`
	Size        int64     `json:"size"`
	Age         string    `json:"age"`
	Sender      string    `json:"sender"`
	Recipients  []string  `json:"recipients"`
	Status      string    `json:"status"` // queued, deferred, frozen
	RetryCount  int       `json:"retry_count"`
	LastAttempt time.Time `json:"last_attempt"`
	NextRetry   time.Time `json:"next_retry"`
}

// QueueStatus represents the overall queue status
type QueueStatus struct {
	TotalMessages    int            `json:"total_messages"`
	DeferredMessages int            `json:"deferred_messages"`
	FrozenMessages   int            `json:"frozen_messages"`
	OldestMessageAge time.Duration  `json:"oldest_message_age"`
	Messages         []QueueMessage `json:"messages"`
}

// Interface defines the queue management operations
type Interface interface {
	ListQueue() (*QueueStatus, error)
	InspectMessage(messageID string) (*MessageDetails, error)
	CreateSnapshot() (*database.QueueSnapshot, error)
}

// Manager implements the queue interface
type Manager struct {
	eximPath        string
	db              *database.DB
	securityService *security.Service
}

// MessageEnvelope represents envelope information for a message
type MessageEnvelope struct {
	Sender      string    `json:"sender"`
	Recipients  []string  `json:"recipients"`
	ReceivedAt  time.Time `json:"received_at"`
	Size        int64     `json:"size"`
}

// MessageDetails represents detailed information about a message
type MessageDetails struct {
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	RetryCount       int               `json:"retry_count"`
	LastAttempt      time.Time         `json:"last_attempt,omitempty"`
	NextRetry        time.Time         `json:"next_retry,omitempty"`
	Envelope         MessageEnvelope   `json:"envelope"`
	Headers          map[string]string `json:"headers"`
	ContentPreview   string            `json:"content_preview,omitempty"`
	SMTPLogs         []SMTPLogEntry    `json:"smtp_logs"`
	DeliveryAttempts []DeliveryAttempt `json:"delivery_attempts"`
}

// SMTPLogEntry represents an SMTP log entry
type SMTPLogEntry struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
	Host      string `json:"host,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Message   string `json:"message"`
}

// DeliveryAttempt represents a delivery attempt
type DeliveryAttempt struct {
	ID           int       `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Recipient    string    `json:"recipient"`
	Host         string    `json:"host,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Status       string    `json:"status"`
	SMTPCode     string    `json:"smtp_code,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// NewManager creates a new queue manager
func NewManager(eximPath string, db *database.DB) *Manager {
	if eximPath == "" {
		eximPath = "/usr/sbin/exim4" // Default path for Ubuntu/Debian
	}
	return &Manager{
		eximPath:        eximPath,
		db:              db,
		securityService: security.NewService(),
	}
}

// createCommand creates an exec.Cmd for the Exim binary, handling Windows batch files
func (m *Manager) createCommand(args ...string) *exec.Cmd {
	// Handle Windows batch files
	if strings.HasSuffix(m.eximPath, ".cmd") || strings.HasSuffix(m.eximPath, ".bat") {
		cmdArgs := append([]string{"/c", m.eximPath}, args...)
		return exec.Command("cmd", cmdArgs...)
	}
	return exec.Command(m.eximPath, args...)
}

// ListQueue retrieves the current queue status using exim -bp
func (m *Manager) ListQueue() (*QueueStatus, error) {
	cmd := m.createCommand("-bp")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute exim -bp: %w", err)
	}

	status, err := m.parseQueueOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse queue output: %w", err)
	}

	return status, nil
}

// parseQueueOutput parses the output from exim -bp command
func (m *Manager) parseQueueOutput(output string) (*QueueStatus, error) {
	status := &QueueStatus{
		Messages: make([]QueueMessage, 0),
	}

	if strings.TrimSpace(output) == "" || strings.Contains(output, "The queue is empty") {
		return status, nil
	}

	lines := strings.Split(output, "\n")
	var currentMessage *QueueMessage

	// Regex patterns for parsing queue output
	messageLineRegex := regexp.MustCompile(`^(\s*)(\d+[a-zA-Z]\s+)?(\d+[KMGT]?)\s+(\S+)\s+<(.*)>$`)
	recipientLineRegex := regexp.MustCompile(`^\s+(.+)$`)
	frozenRegex := regexp.MustCompile(`\*\*\* frozen \*\*\*`)

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

		if line == "" {
			continue
		}

		// Check if this is a message header line
		if matches := messageLineRegex.FindStringSubmatch(line); matches != nil {
			// Save previous message if exists
			if currentMessage != nil {
				status.Messages = append(status.Messages, *currentMessage)
			}

			// Parse message details
			sizeStr := matches[3]
			ageStr := matches[2]
			sender := matches[5]

			size, err := m.parseSize(sizeStr)
			if err != nil {
				continue // Skip malformed entries
			}

			messageID := m.extractMessageID(line)
			if messageID == "" {
				continue // Skip if we can't extract message ID
			}

			currentMessage = &QueueMessage{
				ID:         messageID,
				Size:       size,
				Age:        strings.TrimSpace(ageStr),
				Sender:     sender,
				Recipients: make([]string, 0),
				Status:     "queued",
			}

			// Check if message is frozen
			if frozenRegex.MatchString(line) {
				currentMessage.Status = "frozen"
				status.FrozenMessages++
			}

		} else if currentMessage != nil && recipientLineRegex.MatchString(line) {
			// This is a recipient line
			recipient := strings.TrimSpace(line)
			if recipient != "" {
				currentMessage.Recipients = append(currentMessage.Recipients, recipient)
			}
		}
	}

	// Add the last message
	if currentMessage != nil {
		status.Messages = append(status.Messages, *currentMessage)
	}

	// Calculate statistics
	status.TotalMessages = len(status.Messages)
	for _, msg := range status.Messages {
		if msg.Status == "deferred" {
			status.DeferredMessages++
		}
	}

	// Calculate oldest message age
	if len(status.Messages) > 0 {
		oldestAge := m.parseAge(status.Messages[0].Age)
		for _, msg := range status.Messages {
			age := m.parseAge(msg.Age)
			if age > oldestAge {
				oldestAge = age
			}
		}
		status.OldestMessageAge = oldestAge
	}

	return status, nil
}

// extractMessageID extracts message ID from queue line
func (m *Manager) extractMessageID(line string) string {
	// Message ID is typically at the beginning of the line after whitespace
	// Format: 1a2b3c-000001-AB
	messageIDRegex := regexp.MustCompile(`([0-9a-zA-Z]{6}-[0-9a-zA-Z]{6}-[0-9a-zA-Z]{2})`)
	matches := messageIDRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseSize converts size string to bytes
func (m *Manager) parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, nil
	}

	// Handle size suffixes (K, M, G, T)
	multiplier := int64(1)
	if len(sizeStr) > 0 {
		lastChar := sizeStr[len(sizeStr)-1]
		switch lastChar {
		case 'K':
			multiplier = 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		case 'M':
			multiplier = 1024 * 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		case 'G':
			multiplier = 1024 * 1024 * 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		case 'T':
			multiplier = 1024 * 1024 * 1024 * 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		}
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return size * multiplier, nil
}

// parseAge converts age string to duration
func (m *Manager) parseAge(ageStr string) time.Duration {
	ageStr = strings.TrimSpace(ageStr)
	if ageStr == "" {
		return 0
	}

	// Parse age format like "2h", "30m", "1d"
	if duration, err := time.ParseDuration(ageStr); err == nil {
		return duration
	}

	// Handle day format
	if strings.HasSuffix(ageStr, "d") {
		dayStr := strings.TrimSuffix(ageStr, "d")
		if days, err := strconv.Atoi(dayStr); err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	return 0
}

// InspectMessage retrieves detailed information about a specific message
func (m *Manager) InspectMessage(messageID string) (*MessageDetails, error) {
	// Get message headers using exim -Mvh
	headersCmd := m.createCommand("-Mvh", messageID)
	headersOutput, err := headersCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get message headers: %w", err)
	}

	// Get message body using exim -Mvb
	bodyCmd := m.createCommand("-Mvb", messageID)
	bodyOutput, err := bodyCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get message body: %w", err)
	}

	// Get message log using exim -Mvl
	logCmd := m.createCommand("-Mvl", messageID)
	logOutput, err := logCmd.Output()
	if err != nil {
		// Log might not exist for new messages, continue without error
		logOutput = []byte("")
	}

	// Parse headers
	headers := m.parseHeaders(string(headersOutput))

	// Parse log entries
	smtpLog, _ := m.parseMessageLog(string(logOutput))

	// Get queue message info from current queue status
	var queueMsg *QueueMessage
	status, err := m.ListQueue()
	if err == nil {
		for _, msg := range status.Messages {
			if msg.ID == messageID {
				queueMsg = &msg
				break
			}
		}
	}

	// Create envelope from queue message or headers
	envelope := MessageEnvelope{
		ReceivedAt: time.Now(), // Default to now if not available
		Size:       int64(len(bodyOutput)),
	}

	if queueMsg != nil {
		envelope.Sender = queueMsg.Sender
		envelope.Recipients = queueMsg.Recipients
		envelope.Size = queueMsg.Size
	} else {
		// Populate from headers if not found in queue
		if headers["From"] != "" {
			envelope.Sender = headers["From"]
		}
		if headers["To"] != "" {
			envelope.Recipients = []string{headers["To"]}
		}
	}

	// Convert string logs to structured logs (simplified)
	var smtpLogs []SMTPLogEntry
	for _, log := range smtpLog {
		smtpLogs = append(smtpLogs, SMTPLogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Event:     "smtp",
			Message:   log,
		})
	}

	// Create empty delivery attempts (would be populated from database in real implementation)
	var deliveryAttempts []DeliveryAttempt

	details := &MessageDetails{
		ID:               messageID,
		Status:           "queued",
		RetryCount:       0,
		Envelope:         envelope,
		Headers:          headers,
		ContentPreview:   string(bodyOutput),
		SMTPLogs:         smtpLogs,
		DeliveryAttempts: deliveryAttempts,
	}

	// Set status and retry count from queue message if available
	if queueMsg != nil {
		details.Status = queueMsg.Status
		details.RetryCount = queueMsg.RetryCount
		details.LastAttempt = queueMsg.LastAttempt
		details.NextRetry = queueMsg.NextRetry
	}

	return details, nil
}

// parseHeaders parses message headers from exim -Mvh output
func (m *Manager) parseHeaders(headersOutput string) map[string]string {
	headers := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(headersOutput))
	var currentHeader string
	var currentValue strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a new header (starts with non-whitespace)
		if len(line) > 0 && (line[0] != ' ' && line[0] != '\t') {
			// Save previous header if exists
			if currentHeader != "" {
				headers[currentHeader] = strings.TrimSpace(currentValue.String())
			}

			// Parse new header
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentHeader = strings.TrimSpace(parts[0])
				currentValue.Reset()
				currentValue.WriteString(strings.TrimSpace(parts[1]))
			}
		} else if currentHeader != "" {
			// Continuation of previous header
			currentValue.WriteString(" ")
			currentValue.WriteString(strings.TrimSpace(line))
		}
	}

	// Save last header
	if currentHeader != "" {
		headers[currentHeader] = strings.TrimSpace(currentValue.String())
	}

	return headers
}

// parseMessageLog parses message log from exim -Mvl output
func (m *Manager) parseMessageLog(logOutput string) ([]string, []string) {
	var smtpLog []string
	var deliveryLog []string

	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Categorize log entries
		if strings.Contains(line, "SMTP") || strings.Contains(line, "<=") || strings.Contains(line, "=>") {
			smtpLog = append(smtpLog, line)
		} else {
			deliveryLog = append(deliveryLog, line)
		}
	}

	return smtpLog, deliveryLog
}

// CreateSnapshot creates a queue snapshot for historical tracking
func (m *Manager) CreateSnapshot() (*database.QueueSnapshot, error) {
	status, err := m.ListQueue()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue status: %w", err)
	}

	snapshot := &database.QueueSnapshot{
		Timestamp:        time.Now(),
		TotalMessages:    status.TotalMessages,
		DeferredMessages: status.DeferredMessages,
		FrozenMessages:   status.FrozenMessages,
		CreatedAt:        time.Now(),
	}

	// Convert oldest message age to seconds
	if status.OldestMessageAge > 0 {
		ageSeconds := int(status.OldestMessageAge.Seconds())
		snapshot.OldestMessageAge = &ageSeconds
	}

	// Save snapshot to database
	repo := database.NewRepository(m.db)
	err = repo.CreateQueueSnapshot(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to save queue snapshot: %w", err)
	}

	return snapshot, nil
}
