package database

import (
	"database/sql"
	"fmt"
)

// TxManager provides transaction management utilities
type TxManager struct {
	db *DB
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *DB) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (tm *TxManager) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := tm.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TxRepository provides repository operations within a transaction
type TxRepository struct {
	tx *sql.Tx
}

// NewTxRepository creates a new transaction repository
func NewTxRepository(tx *sql.Tx) *TxRepository {
	return &TxRepository{tx: tx}
}

// CreateMessage inserts a message within a transaction
func (r *TxRepository) CreateMessage(msg *Message) error {
	query := `
		INSERT INTO messages (id, timestamp, sender, size, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.tx.Exec(query, msg.ID, msg.Timestamp, msg.Sender, msg.Size, msg.Status, msg.CreatedAt, msg.UpdatedAt)
	return err
}

// CreateRecipient inserts a recipient within a transaction
func (r *TxRepository) CreateRecipient(recipient *Recipient) error {
	query := `
		INSERT INTO recipients (message_id, recipient, status, delivered_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.tx.Exec(query, recipient.MessageID, recipient.Recipient, recipient.Status, recipient.DeliveredAt, recipient.CreatedAt, recipient.UpdatedAt)
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

// CreateDeliveryAttempt inserts a delivery attempt within a transaction
func (r *TxRepository) CreateDeliveryAttempt(attempt *DeliveryAttempt) error {
	query := `
		INSERT INTO delivery_attempts (message_id, recipient, timestamp, host, ip_address, status, smtp_code, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.tx.Exec(query, attempt.MessageID, attempt.Recipient, attempt.Timestamp, attempt.Host, attempt.IPAddress, attempt.Status, attempt.SMTPCode, attempt.ErrorMessage, attempt.CreatedAt)
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

// CreateLogEntry inserts a log entry within a transaction
func (r *TxRepository) CreateLogEntry(entry *LogEntry) error {
	// Marshal recipients to JSON
	if err := entry.MarshalRecipients(); err != nil {
		return fmt.Errorf("failed to marshal recipients: %w", err)
	}

	query := `
		INSERT INTO log_entries (timestamp, message_id, log_type, event, host, sender, recipients, size, status, error_code, error_text, raw_line, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.tx.Exec(query, entry.Timestamp, entry.MessageID, entry.LogType, entry.Event, entry.Host, entry.Sender, entry.RecipientsDB, entry.Size, entry.Status, entry.ErrorCode, entry.ErrorText, entry.RawLine, entry.CreatedAt)
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

// CreateAuditLog inserts an audit log entry within a transaction
func (r *TxRepository) CreateAuditLog(entry *AuditLog) error {
	query := `
		INSERT INTO audit_log (timestamp, action, message_id, user_id, details, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.tx.Exec(query, entry.Timestamp, entry.Action, entry.MessageID, entry.UserID, entry.Details, entry.IPAddress, entry.CreatedAt)
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
