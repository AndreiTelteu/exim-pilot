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
