package database

import (
	"encoding/json"
	"time"
)

// Message represents a mail message in the system
type Message struct {
	ID        string    `json:"id" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Sender    string    `json:"sender" db:"sender"`
	Size      *int64    `json:"size" db:"size"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MessageStatus constants
const (
	StatusReceived  = "received"
	StatusQueued    = "queued"
	StatusDelivered = "delivered"
	StatusDeferred  = "deferred"
	StatusBounced   = "bounced"
	StatusFrozen    = "frozen"
)

// Recipient represents a message recipient
type Recipient struct {
	ID          int64      `json:"id" db:"id"`
	MessageID   string     `json:"message_id" db:"message_id"`
	Recipient   string     `json:"recipient" db:"recipient"`
	Status      string     `json:"status" db:"status"`
	DeliveredAt *time.Time `json:"delivered_at" db:"delivered_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// RecipientStatus constants
const (
	RecipientStatusDelivered = "delivered"
	RecipientStatusDeferred  = "deferred"
	RecipientStatusBounced   = "bounced"
	RecipientStatusPending   = "pending"
)

// DeliveryAttempt represents a delivery attempt for a message
type DeliveryAttempt struct {
	ID           int64     `json:"id" db:"id"`
	MessageID    string    `json:"message_id" db:"message_id"`
	Recipient    string    `json:"recipient" db:"recipient"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	Host         *string   `json:"host" db:"host"`
	IPAddress    *string   `json:"ip_address" db:"ip_address"`
	Status       string    `json:"status" db:"status"`
	SMTPCode     *string   `json:"smtp_code" db:"smtp_code"`
	ErrorMessage *string   `json:"error_message" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// DeliveryAttemptStatus constants
const (
	AttemptStatusSuccess = "success"
	AttemptStatusDefer   = "defer"
	AttemptStatusBounce  = "bounce"
	AttemptStatusTimeout = "timeout"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	ID           int64     `json:"id" db:"id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	MessageID    *string   `json:"message_id" db:"message_id"`
	LogType      string    `json:"log_type" db:"log_type"`
	Event        string    `json:"event" db:"event"`
	Host         *string   `json:"host" db:"host"`
	Sender       *string   `json:"sender" db:"sender"`
	Recipients   []string  `json:"recipients" db:"-"`
	RecipientsDB *string   `json:"-" db:"recipients"` // JSON string for database
	Size         *int64    `json:"size" db:"size"`
	Status       *string   `json:"status" db:"status"`
	ErrorCode    *string   `json:"error_code" db:"error_code"`
	ErrorText    *string   `json:"error_text" db:"error_text"`
	RawLine      string    `json:"raw_line" db:"raw_line"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// LogType constants
const (
	LogTypeMain   = "main"
	LogTypeReject = "reject"
	LogTypePanic  = "panic"
)

// LogEvent constants
const (
	EventArrival  = "arrival"
	EventDelivery = "delivery"
	EventDefer    = "defer"
	EventBounce   = "bounce"
	EventReject   = "reject"
	EventPanic    = "panic"
)

// MarshalRecipients converts the Recipients slice to JSON for database storage
func (l *LogEntry) MarshalRecipients() error {
	if l.Recipients == nil {
		l.RecipientsDB = nil
		return nil
	}

	data, err := json.Marshal(l.Recipients)
	if err != nil {
		return err
	}

	str := string(data)
	l.RecipientsDB = &str
	return nil
}

// UnmarshalRecipients converts the JSON string from database to Recipients slice
func (l *LogEntry) UnmarshalRecipients() error {
	if l.RecipientsDB == nil {
		l.Recipients = nil
		return nil
	}

	return json.Unmarshal([]byte(*l.RecipientsDB), &l.Recipients)
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        int64     `json:"id" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Action    string    `json:"action" db:"action"`
	MessageID *string   `json:"message_id" db:"message_id"`
	UserID    *string   `json:"user_id" db:"user_id"`
	Details   *string   `json:"details" db:"details"` // JSON string
	IPAddress *string   `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// QueueSnapshot represents a point-in-time snapshot of the queue
type QueueSnapshot struct {
	ID               int64     `json:"id" db:"id"`
	Timestamp        time.Time `json:"timestamp" db:"timestamp"`
	TotalMessages    int       `json:"total_messages" db:"total_messages"`
	DeferredMessages int       `json:"deferred_messages" db:"deferred_messages"`
	FrozenMessages   int       `json:"frozen_messages" db:"frozen_messages"`
	OldestMessageAge *int      `json:"oldest_message_age" db:"oldest_message_age"` // seconds
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// MessageWithRecipients represents a message with its recipients
type MessageWithRecipients struct {
	Message    Message     `json:"message"`
	Recipients []Recipient `json:"recipients"`
}

// DeliveryTrace represents the complete delivery history for a message
type DeliveryTrace struct {
	Message    Message           `json:"message"`
	Attempts   []DeliveryAttempt `json:"attempts"`
	LogEntries []LogEntry        `json:"log_entries"`
}
