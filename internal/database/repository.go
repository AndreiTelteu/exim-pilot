package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Repository provides common database operations
type Repository struct {
	db           *DB
	logEntryRepo *LogEntryRepository
}

// NewRepository creates a new repository instance
func NewRepository(db *DB) *Repository {
	return &Repository{
		db:           db,
		logEntryRepo: NewLogEntryRepository(db),
	}
}

// GetDB returns the database connection
func (r *Repository) GetDB() *DB {
	return r.db
}

// CreateLogEntry creates a log entry (context-aware wrapper)
func (r *Repository) CreateLogEntry(ctx context.Context, entry *LogEntry) error {
	return r.logEntryRepo.Create(entry)
}

// CreateQueueSnapshot creates a queue snapshot
func (r *Repository) CreateQueueSnapshot(snapshot *QueueSnapshot) error {
	repo := NewQueueSnapshotRepository(r.db)
	return repo.Create(snapshot)
}

// MessageRepository handles message-related database operations
type MessageRepository struct {
	*Repository
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{Repository: NewRepository(db)}
}

// Create inserts a new message
func (r *MessageRepository) Create(msg *Message) error {
	query := `
		INSERT INTO messages (id, timestamp, sender, size, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	msg.CreatedAt = now
	msg.UpdatedAt = now

	_, err := r.db.Exec(query, msg.ID, msg.Timestamp, msg.Sender, msg.Size, msg.Status, msg.CreatedAt, msg.UpdatedAt)
	return err
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(id string) (*Message, error) {
	query := `
		SELECT id, timestamp, sender, size, status, created_at, updated_at
		FROM messages WHERE id = ?`

	msg := &Message{}
	err := r.db.QueryRow(query, id).Scan(
		&msg.ID, &msg.Timestamp, &msg.Sender, &msg.Size, &msg.Status, &msg.CreatedAt, &msg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return msg, err
}

// Update updates an existing message
func (r *MessageRepository) Update(msg *Message) error {
	query := `
		UPDATE messages 
		SET timestamp = ?, sender = ?, size = ?, status = ?, updated_at = ?
		WHERE id = ?`

	msg.UpdatedAt = time.Now()

	result, err := r.db.Exec(query, msg.Timestamp, msg.Sender, msg.Size, msg.Status, msg.UpdatedAt, msg.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message with ID %s not found", msg.ID)
	}

	return nil
}

// Delete removes a message by ID
func (r *MessageRepository) Delete(id string) error {
	query := "DELETE FROM messages WHERE id = ?"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message with ID %s not found", id)
	}

	return nil
}

// List retrieves messages with pagination and filtering
func (r *MessageRepository) List(limit, offset int, status string) ([]Message, error) {
	query := `
		SELECT id, timestamp, sender, size, status, created_at, updated_at
		FROM messages`

	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Timestamp, &msg.Sender, &msg.Size, &msg.Status, &msg.CreatedAt, &msg.UpdatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// Count returns the total number of messages, optionally filtered by status
func (r *MessageRepository) Count(status string) (int, error) {
	query := "SELECT COUNT(*) FROM messages"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// RecipientRepository handles recipient-related database operations
type RecipientRepository struct {
	*Repository
}

// NewRecipientRepository creates a new recipient repository
func NewRecipientRepository(db *DB) *RecipientRepository {
	return &RecipientRepository{Repository: NewRepository(db)}
}

// Create inserts a new recipient
func (r *RecipientRepository) Create(recipient *Recipient) error {
	query := `
		INSERT INTO recipients (message_id, recipient, status, delivered_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	recipient.CreatedAt = now
	recipient.UpdatedAt = now

	result, err := r.db.Exec(query, recipient.MessageID, recipient.Recipient, recipient.Status, recipient.DeliveredAt, recipient.CreatedAt, recipient.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	recipient.ID = id
	return nil
}

// GetByMessageID retrieves all recipients for a message
func (r *RecipientRepository) GetByMessageID(messageID string) ([]Recipient, error) {
	query := `
		SELECT id, message_id, recipient, status, delivered_at, created_at, updated_at
		FROM recipients WHERE message_id = ? ORDER BY id`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []Recipient
	for rows.Next() {
		var recipient Recipient
		err := rows.Scan(&recipient.ID, &recipient.MessageID, &recipient.Recipient, &recipient.Status, &recipient.DeliveredAt, &recipient.CreatedAt, &recipient.UpdatedAt)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}

	return recipients, rows.Err()
}

// Update updates an existing recipient
func (r *RecipientRepository) Update(recipient *Recipient) error {
	query := `
		UPDATE recipients 
		SET recipient = ?, status = ?, delivered_at = ?, updated_at = ?
		WHERE id = ?`

	recipient.UpdatedAt = time.Now()

	result, err := r.db.Exec(query, recipient.Recipient, recipient.Status, recipient.DeliveredAt, recipient.UpdatedAt, recipient.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recipient with ID %d not found", recipient.ID)
	}

	return nil
}

// DeliveryAttemptRepository handles delivery attempt operations
type DeliveryAttemptRepository struct {
	*Repository
}

// NewDeliveryAttemptRepository creates a new delivery attempt repository
func NewDeliveryAttemptRepository(db *DB) *DeliveryAttemptRepository {
	return &DeliveryAttemptRepository{Repository: NewRepository(db)}
}

// Create inserts a new delivery attempt
func (r *DeliveryAttemptRepository) Create(attempt *DeliveryAttempt) error {
	query := `
		INSERT INTO delivery_attempts (message_id, recipient, timestamp, host, ip_address, status, smtp_code, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	attempt.CreatedAt = time.Now()

	result, err := r.db.Exec(query, attempt.MessageID, attempt.Recipient, attempt.Timestamp, attempt.Host, attempt.IPAddress, attempt.Status, attempt.SMTPCode, attempt.ErrorMessage, attempt.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	attempt.ID = id
	return nil
}

// GetByMessageID retrieves all delivery attempts for a message
func (r *DeliveryAttemptRepository) GetByMessageID(messageID string) ([]DeliveryAttempt, error) {
	query := `
		SELECT id, message_id, recipient, timestamp, host, ip_address, status, smtp_code, error_message, created_at
		FROM delivery_attempts WHERE message_id = ? ORDER BY timestamp`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []DeliveryAttempt
	for rows.Next() {
		var attempt DeliveryAttempt
		err := rows.Scan(&attempt.ID, &attempt.MessageID, &attempt.Recipient, &attempt.Timestamp, &attempt.Host, &attempt.IPAddress, &attempt.Status, &attempt.SMTPCode, &attempt.ErrorMessage, &attempt.CreatedAt)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, rows.Err()
}

// LogEntryRepository handles log entry operations
type LogEntryRepository struct {
	*Repository
}

// NewLogEntryRepository creates a new log entry repository
func NewLogEntryRepository(db *DB) *LogEntryRepository {
	return &LogEntryRepository{
		Repository: &Repository{db: db},
	}
}

// Create inserts a new log entry
func (r *LogEntryRepository) Create(entry *LogEntry) error {
	// Marshal recipients to JSON
	if err := entry.MarshalRecipients(); err != nil {
		return fmt.Errorf("failed to marshal recipients: %w", err)
	}

	query := `
		INSERT INTO log_entries (timestamp, message_id, log_type, event, host, sender, recipients, size, status, error_code, error_text, raw_line, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	entry.CreatedAt = time.Now()

	result, err := r.db.Exec(query, entry.Timestamp, entry.MessageID, entry.LogType, entry.Event, entry.Host, entry.Sender, entry.RecipientsDB, entry.Size, entry.Status, entry.ErrorCode, entry.ErrorText, entry.RawLine, entry.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	entry.ID = id
	return nil
}

// GetByMessageID retrieves all log entries for a message
func (r *LogEntryRepository) GetByMessageID(messageID string) ([]LogEntry, error) {
	query := `
		SELECT id, timestamp, message_id, log_type, event, host, sender, recipients, size, status, error_code, error_text, raw_line, created_at
		FROM log_entries WHERE message_id = ? ORDER BY timestamp`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.MessageID, &entry.LogType, &entry.Event, &entry.Host, &entry.Sender, &entry.RecipientsDB, &entry.Size, &entry.Status, &entry.ErrorCode, &entry.ErrorText, &entry.RawLine, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal recipients from JSON
		if err := entry.UnmarshalRecipients(); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// List retrieves log entries with pagination and filtering
func (r *LogEntryRepository) List(limit, offset int, logType, event string, startTime, endTime *time.Time) ([]LogEntry, error) {
	query := `
		SELECT id, timestamp, message_id, log_type, event, host, sender, recipients, size, status, error_code, error_text, raw_line, created_at
		FROM log_entries`

	var conditions []string
	var args []interface{}

	if logType != "" {
		conditions = append(conditions, "log_type = ?")
		args = append(args, logType)
	}

	if event != "" {
		conditions = append(conditions, "event = ?")
		args = append(args, event)
	}

	if startTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, *startTime)
	}

	if endTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, *endTime)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.MessageID, &entry.LogType, &entry.Event, &entry.Host, &entry.Sender, &entry.RecipientsDB, &entry.Size, &entry.Status, &entry.ErrorCode, &entry.ErrorText, &entry.RawLine, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal recipients from JSON
		if err := entry.UnmarshalRecipients(); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// AuditLogRepository handles audit log operations
type AuditLogRepository struct {
	*Repository
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *DB) *AuditLogRepository {
	return &AuditLogRepository{Repository: NewRepository(db)}
}

// Create inserts a new audit log entry
func (r *AuditLogRepository) Create(entry *AuditLog) error {
	query := `
		INSERT INTO audit_log (timestamp, action, message_id, user_id, details, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	entry.Timestamp = now
	entry.CreatedAt = now

	result, err := r.db.Exec(query, entry.Timestamp, entry.Action, entry.MessageID, entry.UserID, entry.Details, entry.IPAddress, entry.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	entry.ID = id
	return nil
}

// List retrieves audit log entries with pagination
func (r *AuditLogRepository) List(limit, offset int, action, userID string) ([]AuditLog, error) {
	query := `
		SELECT id, timestamp, action, message_id, user_id, details, ip_address, created_at
		FROM audit_log`

	var conditions []string
	var args []interface{}

	if action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, action)
	}

	if userID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditLog
	for rows.Next() {
		var entry AuditLog
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.Action, &entry.MessageID, &entry.UserID, &entry.Details, &entry.IPAddress, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// QueueSnapshotRepository handles queue snapshot operations
type QueueSnapshotRepository struct {
	*Repository
}

// NewQueueSnapshotRepository creates a new queue snapshot repository
func NewQueueSnapshotRepository(db *DB) *QueueSnapshotRepository {
	return &QueueSnapshotRepository{Repository: NewRepository(db)}
}

// Create inserts a new queue snapshot
func (r *QueueSnapshotRepository) Create(snapshot *QueueSnapshot) error {
	query := `
		INSERT INTO queue_snapshots (timestamp, total_messages, deferred_messages, frozen_messages, oldest_message_age, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	snapshot.Timestamp = now
	snapshot.CreatedAt = now

	result, err := r.db.Exec(query, snapshot.Timestamp, snapshot.TotalMessages, snapshot.DeferredMessages, snapshot.FrozenMessages, snapshot.OldestMessageAge, snapshot.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	snapshot.ID = id
	return nil
}

// GetLatest retrieves the most recent queue snapshot
func (r *QueueSnapshotRepository) GetLatest() (*QueueSnapshot, error) {
	query := `
		SELECT id, timestamp, total_messages, deferred_messages, frozen_messages, oldest_message_age, created_at
		FROM queue_snapshots ORDER BY timestamp DESC LIMIT 1`

	snapshot := &QueueSnapshot{}
	err := r.db.QueryRow(query).Scan(
		&snapshot.ID, &snapshot.Timestamp, &snapshot.TotalMessages, &snapshot.DeferredMessages, &snapshot.FrozenMessages, &snapshot.OldestMessageAge, &snapshot.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return snapshot, err
}

// List retrieves queue snapshots with pagination
func (r *QueueSnapshotRepository) List(limit, offset int, startTime, endTime *time.Time) ([]QueueSnapshot, error) {
	query := `
		SELECT id, timestamp, total_messages, deferred_messages, frozen_messages, oldest_message_age, created_at
		FROM queue_snapshots`

	var conditions []string
	var args []interface{}

	if startTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, *startTime)
	}

	if endTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, *endTime)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []QueueSnapshot
	for rows.Next() {
		var snapshot QueueSnapshot
		err := rows.Scan(&snapshot.ID, &snapshot.Timestamp, &snapshot.TotalMessages, &snapshot.DeferredMessages, &snapshot.FrozenMessages, &snapshot.OldestMessageAge, &snapshot.CreatedAt)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

// DeleteOlderThan removes queue snapshots older than the specified time
func (r *QueueSnapshotRepository) DeleteOlderThan(cutoff time.Time) (int64, error) {
	query := "DELETE FROM queue_snapshots WHERE timestamp < ?"

	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// MessageTraceRepository handles message delivery tracing operations
type MessageTraceRepository struct {
	*Repository
	messageRepo         *MessageRepository
	recipientRepo       *RecipientRepository
	deliveryAttemptRepo *DeliveryAttemptRepository
	logEntryRepo        *LogEntryRepository
	auditLogRepo        *AuditLogRepository
}

// NewMessageTraceRepository creates a new message trace repository
func NewMessageTraceRepository(db *DB) *MessageTraceRepository {
	return &MessageTraceRepository{
		Repository:          NewRepository(db),
		messageRepo:         NewMessageRepository(db),
		recipientRepo:       NewRecipientRepository(db),
		deliveryAttemptRepo: NewDeliveryAttemptRepository(db),
		logEntryRepo:        NewLogEntryRepository(db),
		auditLogRepo:        NewAuditLogRepository(db),
	}
}

// GetMessageDeliveryTrace generates a comprehensive delivery trace for a message
func (r *MessageTraceRepository) GetMessageDeliveryTrace(messageID string) (*MessageDeliveryTrace, error) {
	// Get message details
	message, err := r.messageRepo.GetByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Get recipients
	recipients, err := r.recipientRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipients: %w", err)
	}

	// Get delivery attempts
	attempts, err := r.deliveryAttemptRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery attempts: %w", err)
	}

	// Get log entries
	logEntries, err := r.logEntryRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get log entries: %w", err)
	}

	// Get audit log entries
	auditEntries, err := r.getAuditLogByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit entries: %w", err)
	}

	// Build recipient delivery status
	recipientStatuses := r.buildRecipientDeliveryStatuses(recipients, attempts)

	// Build delivery timeline
	timeline := r.buildDeliveryTimeline(logEntries, attempts, auditEntries)

	// Build retry schedule
	retrySchedule := r.buildRetrySchedule(recipientStatuses, attempts)

	// Calculate summary
	summary := r.calculateDeliveryTraceSummary(recipientStatuses, attempts, message)

	trace := &MessageDeliveryTrace{
		MessageID:        messageID,
		Message:          message,
		Recipients:       recipientStatuses,
		DeliveryTimeline: timeline,
		RetrySchedule:    retrySchedule,
		Summary:          summary,
		GeneratedAt:      time.Now(),
	}

	return trace, nil
}

// buildRecipientDeliveryStatuses creates per-recipient delivery status tracking
func (r *MessageTraceRepository) buildRecipientDeliveryStatuses(recipients []Recipient, attempts []DeliveryAttempt) []RecipientDeliveryStatus {
	recipientMap := make(map[string]*RecipientDeliveryStatus)

	// Initialize recipient statuses
	for _, recipient := range recipients {
		recipientMap[recipient.Recipient] = &RecipientDeliveryStatus{
			Recipient:       recipient.Recipient,
			Status:          recipient.Status,
			DeliveredAt:     recipient.DeliveredAt,
			AttemptCount:    0,
			DeliveryHistory: []DeliveryAttempt{},
		}
	}

	// Process delivery attempts
	for _, attempt := range attempts {
		if status, exists := recipientMap[attempt.Recipient]; exists {
			status.AttemptCount++
			status.DeliveryHistory = append(status.DeliveryHistory, attempt)
			status.LastAttemptAt = &attempt.Timestamp
			status.LastSMTPCode = attempt.SMTPCode
			status.LastErrorText = attempt.ErrorMessage

			// Update status based on latest attempt
			if attempt.Status == AttemptStatusSuccess {
				status.Status = RecipientStatusDelivered
				status.DeliveredAt = &attempt.Timestamp
			} else if attempt.Status == AttemptStatusBounce {
				status.Status = RecipientStatusBounced
			} else if attempt.Status == AttemptStatusDefer {
				status.Status = RecipientStatusDeferred
				// Estimate next retry (simplified calculation)
				nextRetry := attempt.Timestamp.Add(time.Duration(status.AttemptCount*30) * time.Minute)
				status.NextRetryAt = &nextRetry
			}
		}
	}

	// Convert map to slice
	var result []RecipientDeliveryStatus
	for _, status := range recipientMap {
		result = append(result, *status)
	}

	return result
}

// buildDeliveryTimeline creates a chronological timeline of all delivery events
func (r *MessageTraceRepository) buildDeliveryTimeline(logEntries []LogEntry, attempts []DeliveryAttempt, auditEntries []AuditLog) []DeliveryTimelineEvent {
	var timeline []DeliveryTimelineEvent

	// Add log entries to timeline
	for _, entry := range logEntries {
		event := DeliveryTimelineEvent{
			Timestamp:   entry.Timestamp,
			EventType:   entry.Event,
			Host:        entry.Host,
			SMTPCode:    entry.ErrorCode,
			ErrorText:   entry.ErrorText,
			Description: r.formatLogEventDescription(entry),
			Source:      "log",
			SourceID:    &entry.ID,
		}

		// Add recipient if available
		if len(entry.Recipients) > 0 {
			event.Recipient = &entry.Recipients[0]
		}

		timeline = append(timeline, event)
	}

	// Add delivery attempts to timeline
	for _, attempt := range attempts {
		event := DeliveryTimelineEvent{
			Timestamp:   attempt.Timestamp,
			EventType:   "attempt",
			Recipient:   &attempt.Recipient,
			Host:        attempt.Host,
			IPAddress:   attempt.IPAddress,
			SMTPCode:    attempt.SMTPCode,
			ErrorText:   attempt.ErrorMessage,
			Description: r.formatAttemptDescription(attempt),
			Source:      "queue",
			SourceID:    &attempt.ID,
		}
		timeline = append(timeline, event)
	}

	// Add audit entries to timeline
	for _, audit := range auditEntries {
		event := DeliveryTimelineEvent{
			Timestamp:   audit.Timestamp,
			EventType:   audit.Action,
			Description: r.formatAuditDescription(audit),
			Source:      "audit",
			SourceID:    &audit.ID,
		}
		timeline = append(timeline, event)
	}

	// Sort timeline by timestamp
	for i := 0; i < len(timeline)-1; i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[i].Timestamp.After(timeline[j].Timestamp) {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}

	return timeline
}

// buildRetrySchedule creates estimated retry schedule for deferred messages
func (r *MessageTraceRepository) buildRetrySchedule(recipientStatuses []RecipientDeliveryStatus, attempts []DeliveryAttempt) []RetryScheduleEntry {
	var schedule []RetryScheduleEntry

	for _, status := range recipientStatuses {
		if status.Status == RecipientStatusDeferred && status.NextRetryAt != nil {
			entry := RetryScheduleEntry{
				Recipient:     status.Recipient,
				ScheduledAt:   *status.NextRetryAt,
				AttemptNumber: status.AttemptCount + 1,
				Reason:        "Automatic retry after deferral",
				IsEstimated:   true,
			}

			if status.LastErrorText != nil {
				entry.Reason = fmt.Sprintf("Retry after: %s", *status.LastErrorText)
			}

			schedule = append(schedule, entry)
		}
	}

	return schedule
}

// calculateDeliveryTraceSummary calculates summary statistics for the trace
func (r *MessageTraceRepository) calculateDeliveryTraceSummary(recipientStatuses []RecipientDeliveryStatus, attempts []DeliveryAttempt, message *Message) DeliveryTraceSummary {
	summary := DeliveryTraceSummary{
		TotalRecipients: len(recipientStatuses),
	}

	var deliveryTimes []float64
	var firstAttempt, lastAttempt *time.Time

	for _, status := range recipientStatuses {
		switch status.Status {
		case RecipientStatusDelivered:
			summary.DeliveredCount++
			if status.DeliveredAt != nil && message != nil {
				deliveryTime := status.DeliveredAt.Sub(message.Timestamp).Seconds()
				deliveryTimes = append(deliveryTimes, deliveryTime)
			}
		case RecipientStatusDeferred:
			summary.DeferredCount++
		case RecipientStatusBounced:
			summary.BouncedCount++
		default:
			summary.PendingCount++
		}

		if status.LastAttemptAt != nil {
			if firstAttempt == nil || status.LastAttemptAt.Before(*firstAttempt) {
				firstAttempt = status.LastAttemptAt
			}
			if lastAttempt == nil || status.LastAttemptAt.After(*lastAttempt) {
				lastAttempt = status.LastAttemptAt
			}
		}
	}

	summary.TotalAttempts = len(attempts)
	summary.FirstAttemptAt = firstAttempt
	summary.LastAttemptAt = lastAttempt

	// Calculate average delivery time
	if len(deliveryTimes) > 0 {
		var total float64
		for _, t := range deliveryTimes {
			total += t
		}
		avg := total / float64(len(deliveryTimes))
		summary.AverageDeliveryTime = &avg
	}

	return summary
}

// Helper methods for formatting descriptions

func (r *MessageTraceRepository) formatLogEventDescription(entry LogEntry) string {
	switch entry.Event {
	case EventArrival:
		return fmt.Sprintf("Message arrived from %s", *entry.Sender)
	case EventDelivery:
		return fmt.Sprintf("Successfully delivered to %s", strings.Join(entry.Recipients, ", "))
	case EventDefer:
		if entry.ErrorText != nil {
			return fmt.Sprintf("Delivery deferred: %s", *entry.ErrorText)
		}
		return "Delivery deferred"
	case EventBounce:
		if entry.ErrorText != nil {
			return fmt.Sprintf("Message bounced: %s", *entry.ErrorText)
		}
		return "Message bounced"
	case EventReject:
		if entry.ErrorText != nil {
			return fmt.Sprintf("Message rejected: %s", *entry.ErrorText)
		}
		return "Message rejected"
	default:
		return fmt.Sprintf("Log event: %s", entry.Event)
	}
}

func (r *MessageTraceRepository) formatAttemptDescription(attempt DeliveryAttempt) string {
	switch attempt.Status {
	case AttemptStatusSuccess:
		return fmt.Sprintf("Successfully delivered to %s", attempt.Recipient)
	case AttemptStatusDefer:
		if attempt.ErrorMessage != nil {
			return fmt.Sprintf("Delivery to %s deferred: %s", attempt.Recipient, *attempt.ErrorMessage)
		}
		return fmt.Sprintf("Delivery to %s deferred", attempt.Recipient)
	case AttemptStatusBounce:
		if attempt.ErrorMessage != nil {
			return fmt.Sprintf("Delivery to %s bounced: %s", attempt.Recipient, *attempt.ErrorMessage)
		}
		return fmt.Sprintf("Delivery to %s bounced", attempt.Recipient)
	case AttemptStatusTimeout:
		return fmt.Sprintf("Delivery to %s timed out", attempt.Recipient)
	default:
		return fmt.Sprintf("Delivery attempt to %s: %s", attempt.Recipient, attempt.Status)
	}
}

func (r *MessageTraceRepository) formatAuditDescription(audit AuditLog) string {
	switch audit.Action {
	case "deliver_now":
		return "Manual delivery triggered"
	case "freeze":
		return "Message frozen by operator"
	case "thaw":
		return "Message thawed by operator"
	case "delete":
		return "Message deleted by operator"
	default:
		return fmt.Sprintf("Administrative action: %s", audit.Action)
	}
}

// getAuditLogByMessageID retrieves audit log entries for a specific message
func (r *MessageTraceRepository) getAuditLogByMessageID(messageID string) ([]AuditLog, error) {
	query := `
		SELECT id, timestamp, action, message_id, user_id, details, ip_address, created_at
		FROM audit_log WHERE message_id = ? ORDER BY timestamp`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditLog
	for rows.Next() {
		var entry AuditLog
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.Action, &entry.MessageID, &entry.UserID, &entry.Details, &entry.IPAddress, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// MessageNoteRepository handles message note operations
type MessageNoteRepository struct {
	*Repository
}

// NewMessageNoteRepository creates a new message note repository
func NewMessageNoteRepository(db *DB) *MessageNoteRepository {
	return &MessageNoteRepository{Repository: NewRepository(db)}
}

// Create inserts a new message note
func (r *MessageNoteRepository) Create(note *MessageNote) error {
	query := `
		INSERT INTO message_notes (message_id, user_id, note, is_public, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now

	result, err := r.db.Exec(query, note.MessageID, note.UserID, note.Note, note.IsPublic, note.CreatedAt, note.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	note.ID = id
	return nil
}

// GetByMessageID retrieves all notes for a message
func (r *MessageNoteRepository) GetByMessageID(messageID string) ([]MessageNote, error) {
	query := `
		SELECT id, message_id, user_id, note, is_public, created_at, updated_at
		FROM message_notes WHERE message_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []MessageNote
	for rows.Next() {
		var note MessageNote
		err := rows.Scan(&note.ID, &note.MessageID, &note.UserID, &note.Note, &note.IsPublic, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// Update updates an existing message note
func (r *MessageNoteRepository) Update(note *MessageNote) error {
	query := `
		UPDATE message_notes 
		SET note = ?, is_public = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`

	note.UpdatedAt = time.Now()

	result, err := r.db.Exec(query, note.Note, note.IsPublic, note.UpdatedAt, note.ID, note.UserID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with ID %d not found or not owned by user", note.ID)
	}

	return nil
}

// Delete removes a message note
func (r *MessageNoteRepository) Delete(noteID int64, userID string) error {
	query := "DELETE FROM message_notes WHERE id = ? AND user_id = ?"

	result, err := r.db.Exec(query, noteID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with ID %d not found or not owned by user", noteID)
	}

	return nil
}

// MessageTagRepository handles message tag operations
type MessageTagRepository struct {
	*Repository
}

// NewMessageTagRepository creates a new message tag repository
func NewMessageTagRepository(db *DB) *MessageTagRepository {
	return &MessageTagRepository{Repository: NewRepository(db)}
}

// Create inserts a new message tag
func (r *MessageTagRepository) Create(tag *MessageTag) error {
	query := `
		INSERT INTO message_tags (message_id, tag, color, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)`

	tag.CreatedAt = time.Now()

	result, err := r.db.Exec(query, tag.MessageID, tag.Tag, tag.Color, tag.UserID, tag.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	tag.ID = id
	return nil
}

// GetByMessageID retrieves all tags for a message
func (r *MessageTagRepository) GetByMessageID(messageID string) ([]MessageTag, error) {
	query := `
		SELECT id, message_id, tag, color, user_id, created_at
		FROM message_tags WHERE message_id = ? ORDER BY created_at`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []MessageTag
	for rows.Next() {
		var tag MessageTag
		err := rows.Scan(&tag.ID, &tag.MessageID, &tag.Tag, &tag.Color, &tag.UserID, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// Delete removes a message tag
func (r *MessageTagRepository) Delete(tagID int64, userID string) error {
	query := "DELETE FROM message_tags WHERE id = ? AND user_id = ?"

	result, err := r.db.Exec(query, tagID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag with ID %d not found or not owned by user", tagID)
	}

	return nil
}

// GetPopularTags retrieves the most commonly used tags
func (r *MessageTagRepository) GetPopularTags(limit int) ([]string, error) {
	query := `
		SELECT tag, COUNT(*) as count
		FROM message_tags 
		GROUP BY tag 
		ORDER BY count DESC 
		LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		var count int
		err := rows.Scan(&tag, &count)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetThreadedTimelineView generates a threaded view of delivery events with notes and tags
func (r *MessageTraceRepository) GetThreadedTimelineView(messageID string) (*ThreadedTimelineView, error) {
	// Get basic delivery trace
	trace, err := r.GetMessageDeliveryTrace(messageID)
	if err != nil {
		return nil, err
	}

	// Get notes
	noteRepo := NewMessageNoteRepository(r.db)
	notes, err := noteRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	// Get tags
	tagRepo := NewMessageTagRepository(r.db)
	tags, err := tagRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	// Create threads from timeline events
	threads := r.createDeliveryThreads(trace.DeliveryTimeline)

	// Get correlated incidents (simplified implementation)
	incidents := r.getCorrelatedIncidents(messageID)

	return &ThreadedTimelineView{
		MessageID: messageID,
		Threads:   threads,
		Notes:     notes,
		Tags:      tags,
		Incidents: incidents,
	}, nil
}

// createDeliveryThreads groups timeline events into logical threads
func (r *MessageTraceRepository) createDeliveryThreads(timeline []DeliveryTimelineEvent) []DeliveryThread {
	// Group events by recipient
	recipientThreads := make(map[string][]DeliveryTimelineEvent)
	hostThreads := make(map[string][]DeliveryTimelineEvent)
	systemEvents := []DeliveryTimelineEvent{}

	for _, event := range timeline {
		if event.Recipient != nil {
			recipientThreads[*event.Recipient] = append(recipientThreads[*event.Recipient], event)
		} else if event.Host != nil {
			hostThreads[*event.Host] = append(hostThreads[*event.Host], event)
		} else {
			systemEvents = append(systemEvents, event)
		}
	}

	var threads []DeliveryThread

	// Create recipient threads
	for recipient, events := range recipientThreads {
		threadID := fmt.Sprintf("recipient-%s", recipient)
		status := r.getThreadStatus(events)
		summary := r.getThreadSummary("recipient", events)

		threads = append(threads, DeliveryThread{
			ThreadID:   threadID,
			Recipient:  &recipient,
			ThreadType: "recipient",
			Events:     events,
			Summary:    summary,
			Status:     status,
		})
	}

	// Create host threads
	for host, events := range hostThreads {
		threadID := fmt.Sprintf("host-%s", host)
		status := r.getThreadStatus(events)
		summary := r.getThreadSummary("host", events)

		threads = append(threads, DeliveryThread{
			ThreadID:   threadID,
			ThreadType: "host",
			Events:     events,
			Summary:    summary,
			Status:     status,
		})
	}

	// Create system thread if there are system events
	if len(systemEvents) > 0 {
		threads = append(threads, DeliveryThread{
			ThreadID:   "system",
			ThreadType: "system",
			Events:     systemEvents,
			Summary:    r.getThreadSummary("system", systemEvents),
			Status:     "info",
		})
	}

	return threads
}

// getThreadStatus determines the overall status of a thread
func (r *MessageTraceRepository) getThreadStatus(events []DeliveryTimelineEvent) string {
	hasError := false
	hasSuccess := false

	for _, event := range events {
		switch event.EventType {
		case "delivery":
			hasSuccess = true
		case "bounce", "reject":
			return "error"
		case "defer":
			hasError = true
		}
	}

	if hasSuccess {
		return "success"
	} else if hasError {
		return "warning"
	}

	return "info"
}

// getThreadSummary creates a summary for a thread
func (r *MessageTraceRepository) getThreadSummary(threadType string, events []DeliveryTimelineEvent) string {
	if len(events) == 0 {
		return "No events"
	}

	switch threadType {
	case "recipient":
		return fmt.Sprintf("%d delivery events", len(events))
	case "host":
		return fmt.Sprintf("%d host interactions", len(events))
	case "system":
		return fmt.Sprintf("%d system events", len(events))
	default:
		return fmt.Sprintf("%d events", len(events))
	}
}

// getCorrelatedIncidents finds related incidents (simplified implementation)
func (r *MessageTraceRepository) getCorrelatedIncidents(messageID string) []CorrelatedIncident {
	// This is a simplified implementation
	// In a real system, this would analyze patterns and correlate with other messages
	return []CorrelatedIncident{}
}

// GetMessageContent retrieves safe message content preview
func (r *MessageTraceRepository) GetMessageContent(messageID string) (*MessageContent, error) {
	// This is a placeholder implementation
	// In a real system, this would safely parse and preview message content
	// from the Exim spool files

	content := &MessageContent{
		MessageID:      messageID,
		Headers:        make(map[string]string),
		Attachments:    []MessageAttachment{},
		ContentSafe:    true,
		SizeBytes:      0,
		PreviewLimited: false,
	}

	// Add warning that this is not implemented
	textContent := "Message content preview is not yet implemented. This would require parsing Exim spool files safely."
	content.TextContent = &textContent

	return content, nil
}
