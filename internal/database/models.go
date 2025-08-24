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

// MessageDeliveryTrace represents enhanced message tracing with timeline
type MessageDeliveryTrace struct {
	MessageID        string                    `json:"message_id"`
	Message          *Message                  `json:"message,omitempty"`
	Recipients       []RecipientDeliveryStatus `json:"recipients"`
	DeliveryTimeline []DeliveryTimelineEvent   `json:"delivery_timeline"`
	RetrySchedule    []RetryScheduleEntry      `json:"retry_schedule"`
	Summary          DeliveryTraceSummary      `json:"summary"`
	GeneratedAt      time.Time                 `json:"generated_at"`
}

// RecipientDeliveryStatus represents per-recipient delivery tracking
type RecipientDeliveryStatus struct {
	Recipient       string            `json:"recipient"`
	Status          string            `json:"status"` // delivered, deferred, bounced, pending
	DeliveredAt     *time.Time        `json:"delivered_at,omitempty"`
	LastAttemptAt   *time.Time        `json:"last_attempt_at,omitempty"`
	NextRetryAt     *time.Time        `json:"next_retry_at,omitempty"`
	AttemptCount    int               `json:"attempt_count"`
	LastSMTPCode    *string           `json:"last_smtp_code,omitempty"`
	LastErrorText   *string           `json:"last_error_text,omitempty"`
	DeliveryHistory []DeliveryAttempt `json:"delivery_history"`
}

// DeliveryTimelineEvent represents a single event in the delivery timeline
type DeliveryTimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"` // arrival, attempt, delivery, defer, bounce, freeze, thaw, delete
	Recipient   *string   `json:"recipient,omitempty"`
	Host        *string   `json:"host,omitempty"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	SMTPCode    *string   `json:"smtp_code,omitempty"`
	ErrorText   *string   `json:"error_text,omitempty"`
	Description string    `json:"description"`
	Source      string    `json:"source"` // log, queue, audit
	SourceID    *int64    `json:"source_id,omitempty"`
}

// RetryScheduleEntry represents a scheduled retry attempt
type RetryScheduleEntry struct {
	Recipient     string    `json:"recipient"`
	ScheduledAt   time.Time `json:"scheduled_at"`
	AttemptNumber int       `json:"attempt_number"`
	Reason        string    `json:"reason"`
	IsEstimated   bool      `json:"is_estimated"` // true if calculated, false if from queue
}

// DeliveryTraceSummary provides summary statistics for the trace
type DeliveryTraceSummary struct {
	TotalRecipients     int        `json:"total_recipients"`
	DeliveredCount      int        `json:"delivered_count"`
	DeferredCount       int        `json:"deferred_count"`
	BouncedCount        int        `json:"bounced_count"`
	PendingCount        int        `json:"pending_count"`
	TotalAttempts       int        `json:"total_attempts"`
	FirstAttemptAt      *time.Time `json:"first_attempt_at,omitempty"`
	LastAttemptAt       *time.Time `json:"last_attempt_at,omitempty"`
	AverageDeliveryTime *float64   `json:"average_delivery_time_seconds,omitempty"`
}

// MessageNote represents operator notes for messages
type MessageNote struct {
	ID        int64     `json:"id" db:"id"`
	MessageID string    `json:"message_id" db:"message_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Note      string    `json:"note" db:"note"`
	IsPublic  bool      `json:"is_public" db:"is_public"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MessageTag represents tags for messages
type MessageTag struct {
	ID        int64     `json:"id" db:"id"`
	MessageID string    `json:"message_id" db:"message_id"`
	Tag       string    `json:"tag" db:"tag"`
	Color     *string   `json:"color" db:"color"` // hex color code
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MessageContent represents safe message content preview
type MessageContent struct {
	MessageID      string              `json:"message_id"`
	Headers        map[string]string   `json:"headers"`
	TextContent    *string             `json:"text_content,omitempty"`
	HTMLContent    *string             `json:"html_content,omitempty"`
	Attachments    []MessageAttachment `json:"attachments"`
	ContentSafe    bool                `json:"content_safe"`
	SizeBytes      int64               `json:"size_bytes"`
	PreviewLimited bool                `json:"preview_limited"`
}

// MessageAttachment represents a safe attachment preview
type MessageAttachment struct {
	Filename    string  `json:"filename"`
	ContentType string  `json:"content_type"`
	Size        int64   `json:"size"`
	IsSafe      bool    `json:"is_safe"`
	Preview     *string `json:"preview,omitempty"` // safe preview text
}

// CorrelatedIncident represents related incidents or patterns
type CorrelatedIncident struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"` // bounce_pattern, delivery_issue, spam_report
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Severity    string     `json:"severity"` // low, medium, high, critical
	MessageIDs  []string   `json:"message_ids"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Status      string     `json:"status"` // active, resolved, investigating
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ThreadedTimelineView represents a threaded view of delivery events
type ThreadedTimelineView struct {
	MessageID string               `json:"message_id"`
	Threads   []DeliveryThread     `json:"threads"`
	Notes     []MessageNote        `json:"notes"`
	Tags      []MessageTag         `json:"tags"`
	Incidents []CorrelatedIncident `json:"correlated_incidents"`
}

// DeliveryThread represents a thread of related delivery events
type DeliveryThread struct {
	ThreadID   string                  `json:"thread_id"`
	Recipient  *string                 `json:"recipient,omitempty"`
	ThreadType string                  `json:"thread_type"` // recipient, host, error_pattern
	Events     []DeliveryTimelineEvent `json:"events"`
	Summary    string                  `json:"summary"`
	Status     string                  `json:"status"`
}

// User represents a system user for authentication
type User struct {
	ID           int64      `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never include in JSON
	Email        *string    `json:"email" db:"email"`
	FullName     *string    `json:"full_name" db:"full_name"`
	IsActive     bool       `json:"is_active" db:"active"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IPAddress *string   `json:"ip_address" db:"ip_address"`
	UserAgent *string   `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User      User      `json:"user"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
