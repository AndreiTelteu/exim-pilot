package logprocessor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// LogAggregator handles log entry aggregation and message correlation
type LogAggregator struct {
	repository *database.Repository
}

// NewLogAggregator creates a new log aggregator
func NewLogAggregator(repository *database.Repository) *LogAggregator {
	return &LogAggregator{
		repository: repository,
	}
}

// MessageCorrelation represents correlated log entries for a message
type MessageCorrelation struct {
	MessageID  string                     `json:"message_id"`
	Message    *database.Message          `json:"message,omitempty"`
	Recipients []database.Recipient       `json:"recipients,omitempty"`
	LogEntries []database.LogEntry        `json:"log_entries"`
	Attempts   []database.DeliveryAttempt `json:"attempts,omitempty"`
	Timeline   []TimelineEvent            `json:"timeline"`
	Summary    MessageSummary             `json:"summary"`
}

// TimelineEvent represents an event in the message timeline
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Event       string    `json:"event"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Details     string    `json:"details,omitempty"`
}

// MessageSummary provides aggregated information about a message
type MessageSummary struct {
	FirstSeen       time.Time `json:"first_seen"`
	LastActivity    time.Time `json:"last_activity"`
	TotalRecipients int       `json:"total_recipients"`
	DeliveredCount  int       `json:"delivered_count"`
	DeferredCount   int       `json:"deferred_count"`
	BouncedCount    int       `json:"bounced_count"`
	AttemptCount    int       `json:"attempt_count"`
	FinalStatus     string    `json:"final_status"`
	Duration        string    `json:"duration"`
}

// AggregateMessageData correlates all data for a specific message
func (a *LogAggregator) AggregateMessageData(ctx context.Context, messageID string) (*MessageCorrelation, error) {
	correlation := &MessageCorrelation{
		MessageID: messageID,
		Timeline:  make([]TimelineEvent, 0),
	}

	// Get message record if it exists
	messageRepo := database.NewMessageRepository(a.repository.GetDB())
	message, err := messageRepo.GetByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	correlation.Message = message

	// Get all log entries for this message
	logRepo := database.NewLogEntryRepository(a.repository.GetDB())
	logEntries, err := logRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get log entries: %w", err)
	}
	correlation.LogEntries = logEntries

	// Get recipients if message exists
	if message != nil {
		recipientRepo := database.NewRecipientRepository(a.repository.GetDB())
		recipients, err := recipientRepo.GetByMessageID(messageID)
		if err != nil {
			return nil, fmt.Errorf("failed to get recipients: %w", err)
		}
		correlation.Recipients = recipients
	}

	// Get delivery attempts
	attemptRepo := database.NewDeliveryAttemptRepository(a.repository.GetDB())
	attempts, err := attemptRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery attempts: %w", err)
	}
	correlation.Attempts = attempts

	// Build timeline from log entries and attempts
	a.buildTimeline(correlation)

	// Generate summary
	a.generateSummary(correlation)

	return correlation, nil
}

// buildTimeline creates a chronological timeline of events
func (a *LogAggregator) buildTimeline(correlation *MessageCorrelation) {
	events := make([]TimelineEvent, 0)

	// Add log entries to timeline
	for _, entry := range correlation.LogEntries {
		event := TimelineEvent{
			Timestamp: entry.Timestamp,
			Event:     entry.Event,
			Status:    getEventStatus(entry.Event),
		}

		switch entry.Event {
		case database.EventArrival:
			event.Description = fmt.Sprintf("Message received from %s", getStringValue(entry.Sender))
			if entry.Size != nil {
				event.Details = fmt.Sprintf("Size: %d bytes", *entry.Size)
			}
		case database.EventDelivery:
			recipients := getRecipientsString(entry.Recipients)
			event.Description = fmt.Sprintf("Delivered to %s", recipients)
			if entry.Host != nil {
				event.Details = fmt.Sprintf("Host: %s", *entry.Host)
			}
		case database.EventDefer:
			recipients := getRecipientsString(entry.Recipients)
			event.Description = fmt.Sprintf("Deferred for %s", recipients)
			if entry.ErrorText != nil {
				event.Details = *entry.ErrorText
			}
		case database.EventBounce:
			recipients := getRecipientsString(entry.Recipients)
			event.Description = fmt.Sprintf("Bounced for %s", recipients)
			if entry.ErrorText != nil {
				event.Details = *entry.ErrorText
			}
		case database.EventReject:
			event.Description = "Message rejected"
			if entry.ErrorText != nil {
				event.Details = *entry.ErrorText
			}
		default:
			event.Description = fmt.Sprintf("Event: %s", entry.Event)
			if entry.ErrorText != nil {
				event.Details = *entry.ErrorText
			}
		}

		events = append(events, event)
	}

	// Add delivery attempts to timeline
	for _, attempt := range correlation.Attempts {
		event := TimelineEvent{
			Timestamp:   attempt.Timestamp,
			Event:       "delivery_attempt",
			Description: fmt.Sprintf("Delivery attempt to %s", attempt.Recipient),
			Status:      attempt.Status,
		}

		details := make([]string, 0)
		if attempt.Host != nil {
			details = append(details, fmt.Sprintf("Host: %s", *attempt.Host))
		}
		if attempt.IPAddress != nil {
			details = append(details, fmt.Sprintf("IP: %s", *attempt.IPAddress))
		}
		if attempt.SMTPCode != nil {
			details = append(details, fmt.Sprintf("SMTP: %s", *attempt.SMTPCode))
		}
		if attempt.ErrorMessage != nil {
			details = append(details, *attempt.ErrorMessage)
		}

		if len(details) > 0 {
			event.Details = fmt.Sprintf("%s", details[0])
			if len(details) > 1 {
				for _, detail := range details[1:] {
					event.Details += ", " + detail
				}
			}
		}

		events = append(events, event)
	}

	// Sort timeline by timestamp
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.After(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	correlation.Timeline = events
}

// generateSummary creates a summary of the message correlation
func (a *LogAggregator) generateSummary(correlation *MessageCorrelation) {
	summary := MessageSummary{
		FinalStatus: "unknown",
	}

	// Find first and last activity
	if len(correlation.Timeline) > 0 {
		summary.FirstSeen = correlation.Timeline[0].Timestamp
		summary.LastActivity = correlation.Timeline[len(correlation.Timeline)-1].Timestamp
		summary.Duration = summary.LastActivity.Sub(summary.FirstSeen).String()
	}

	// Count recipients and their statuses
	recipientMap := make(map[string]string)
	for _, recipient := range correlation.Recipients {
		recipientMap[recipient.Recipient] = recipient.Status
		summary.TotalRecipients++

		switch recipient.Status {
		case database.RecipientStatusDelivered:
			summary.DeliveredCount++
		case database.RecipientStatusDeferred:
			summary.DeferredCount++
		case database.RecipientStatusBounced:
			summary.BouncedCount++
		}
	}

	// If no recipients in database, count from log entries
	if summary.TotalRecipients == 0 {
		recipientSet := make(map[string]bool)
		for _, entry := range correlation.LogEntries {
			for _, recipient := range entry.Recipients {
				recipientSet[recipient] = true
			}
		}
		summary.TotalRecipients = len(recipientSet)
	}

	// Count delivery attempts
	summary.AttemptCount = len(correlation.Attempts)

	// Determine final status
	if summary.DeliveredCount == summary.TotalRecipients && summary.TotalRecipients > 0 {
		summary.FinalStatus = "delivered"
	} else if summary.BouncedCount > 0 {
		summary.FinalStatus = "bounced"
	} else if summary.DeferredCount > 0 {
		summary.FinalStatus = "deferred"
	} else if correlation.Message != nil {
		summary.FinalStatus = correlation.Message.Status
	}

	correlation.Summary = summary
}

// CorrelateLogEntries finds and correlates log entries that belong to the same message
func (a *LogAggregator) CorrelateLogEntries(ctx context.Context, startTime, endTime time.Time) error {
	log.Printf("Starting log entry correlation for period %s to %s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	// Get all log entries in the time range that have message IDs
	logRepo := database.NewLogEntryRepository(a.repository.GetDB())
	entries, err := logRepo.List(10000, 0, "", "", &startTime, &endTime)
	if err != nil {
		return fmt.Errorf("failed to get log entries: %w", err)
	}

	messageMap := make(map[string][]database.LogEntry)
	messageRepo := database.NewMessageRepository(a.repository.GetDB())
	recipientRepo := database.NewRecipientRepository(a.repository.GetDB())
	attemptRepo := database.NewDeliveryAttemptRepository(a.repository.GetDB())

	// Group entries by message ID
	for _, entry := range entries {
		if entry.MessageID != nil && *entry.MessageID != "" {
			messageMap[*entry.MessageID] = append(messageMap[*entry.MessageID], entry)
		}
	}

	log.Printf("Found %d unique messages to correlate", len(messageMap))

	// Process each message
	processedCount := 0
	for messageID, messageEntries := range messageMap {
		if err := a.correlateMessageEntries(ctx, messageID, messageEntries, messageRepo, recipientRepo, attemptRepo); err != nil {
			log.Printf("Failed to correlate message %s: %v", messageID, err)
			continue
		}
		processedCount++

		if processedCount%100 == 0 {
			log.Printf("Processed %d messages", processedCount)
		}
	}

	log.Printf("Completed correlation of %d messages", processedCount)
	return nil
}

// correlateMessageEntries correlates log entries for a single message
func (a *LogAggregator) correlateMessageEntries(ctx context.Context, messageID string, entries []database.LogEntry,
	messageRepo *database.MessageRepository, recipientRepo *database.RecipientRepository, attemptRepo *database.DeliveryAttemptRepository) error {

	// Check if message record exists
	message, err := messageRepo.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("failed to check message: %w", err)
	}

	// Create or update message record based on log entries
	if message == nil {
		message, err = a.createMessageFromLogEntries(messageID, entries)
		if err != nil {
			return fmt.Errorf("failed to create message from log entries: %w", err)
		}

		if err := messageRepo.Create(message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	} else {
		// Update message status based on latest entries
		if err := a.updateMessageFromLogEntries(message, entries); err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}

		if err := messageRepo.Update(message); err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
	}

	// Create or update recipients and delivery attempts
	if err := a.updateRecipientsFromLogEntries(messageID, entries, recipientRepo, attemptRepo); err != nil {
		return fmt.Errorf("failed to update recipients: %w", err)
	}

	return nil
}

// createMessageFromLogEntries creates a message record from log entries
func (a *LogAggregator) createMessageFromLogEntries(messageID string, entries []database.LogEntry) (*database.Message, error) {
	message := &database.Message{
		ID:     messageID,
		Status: database.StatusReceived,
	}

	// Find the earliest timestamp and sender
	for _, entry := range entries {
		if message.Timestamp.IsZero() || entry.Timestamp.Before(message.Timestamp) {
			message.Timestamp = entry.Timestamp
		}

		if entry.Sender != nil && message.Sender == "" {
			message.Sender = *entry.Sender
		}

		if entry.Size != nil && message.Size == nil {
			message.Size = entry.Size
		}

		// Update status based on events
		switch entry.Event {
		case database.EventDelivery:
			message.Status = database.StatusDelivered
		case database.EventDefer:
			if message.Status != database.StatusDelivered {
				message.Status = database.StatusDeferred
			}
		case database.EventBounce:
			message.Status = database.StatusBounced
		}
	}

	return message, nil
}

// updateMessageFromLogEntries updates a message record based on log entries
func (a *LogAggregator) updateMessageFromLogEntries(message *database.Message, entries []database.LogEntry) error {
	// Update status based on latest events
	hasDelivery := false
	hasDefer := false
	hasBounce := false

	for _, entry := range entries {
		switch entry.Event {
		case database.EventDelivery:
			hasDelivery = true
		case database.EventDefer:
			hasDefer = true
		case database.EventBounce:
			hasBounce = true
		}
	}

	// Determine final status
	if hasBounce {
		message.Status = database.StatusBounced
	} else if hasDelivery && !hasDefer {
		message.Status = database.StatusDelivered
	} else if hasDefer {
		message.Status = database.StatusDeferred
	}

	return nil
}

// updateRecipientsFromLogEntries creates or updates recipient records
func (a *LogAggregator) updateRecipientsFromLogEntries(messageID string, entries []database.LogEntry,
	recipientRepo *database.RecipientRepository, attemptRepo *database.DeliveryAttemptRepository) error {

	// Get existing recipients
	existingRecipients, err := recipientRepo.GetByMessageID(messageID)
	if err != nil {
		return fmt.Errorf("failed to get existing recipients: %w", err)
	}

	recipientMap := make(map[string]*database.Recipient)
	for i := range existingRecipients {
		recipientMap[existingRecipients[i].Recipient] = &existingRecipients[i]
	}

	// Process log entries to update recipient status
	for _, entry := range entries {
		for _, recipientAddr := range entry.Recipients {
			recipient, exists := recipientMap[recipientAddr]
			if !exists {
				// Create new recipient
				recipient = &database.Recipient{
					MessageID: messageID,
					Recipient: recipientAddr,
					Status:    database.RecipientStatusPending,
				}
				recipientMap[recipientAddr] = recipient
			}

			// Update status based on event
			switch entry.Event {
			case database.EventDelivery:
				recipient.Status = database.RecipientStatusDelivered
				recipient.DeliveredAt = &entry.Timestamp
			case database.EventDefer:
				if recipient.Status != database.RecipientStatusDelivered {
					recipient.Status = database.RecipientStatusDeferred
				}
			case database.EventBounce:
				recipient.Status = database.RecipientStatusBounced
			}

			// Create delivery attempt record
			if entry.Event == database.EventDelivery || entry.Event == database.EventDefer || entry.Event == database.EventBounce {
				attempt := &database.DeliveryAttempt{
					MessageID: messageID,
					Recipient: recipientAddr,
					Timestamp: entry.Timestamp,
					Host:      entry.Host,
					Status:    getAttemptStatus(entry.Event),
				}

				if entry.ErrorText != nil {
					attempt.ErrorMessage = entry.ErrorText
				}

				if err := attemptRepo.Create(attempt); err != nil {
					log.Printf("Failed to create delivery attempt: %v", err)
				}
			}
		}
	}

	// Save or update recipients
	for _, recipient := range recipientMap {
		if recipient.ID == 0 {
			if err := recipientRepo.Create(recipient); err != nil {
				log.Printf("Failed to create recipient: %v", err)
			}
		} else {
			if err := recipientRepo.Update(recipient); err != nil {
				log.Printf("Failed to update recipient: %v", err)
			}
		}
	}

	return nil
}

// Helper functions

func getEventStatus(event string) string {
	switch event {
	case database.EventArrival:
		return "received"
	case database.EventDelivery:
		return "delivered"
	case database.EventDefer:
		return "deferred"
	case database.EventBounce:
		return "bounced"
	case database.EventReject:
		return "rejected"
	default:
		return "unknown"
	}
}

func getAttemptStatus(event string) string {
	switch event {
	case database.EventDelivery:
		return database.AttemptStatusSuccess
	case database.EventDefer:
		return database.AttemptStatusDefer
	case database.EventBounce:
		return database.AttemptStatusBounce
	default:
		return database.AttemptStatusTimeout
	}
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return "unknown"
	}
	return *ptr
}

func getRecipientsString(recipients []string) string {
	if len(recipients) == 0 {
		return "unknown recipient"
	}
	if len(recipients) == 1 {
		return recipients[0]
	}
	return fmt.Sprintf("%s and %d others", recipients[0], len(recipients)-1)
}
