package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/queue"
)

// MessageTraceHandlers contains handlers for message tracing endpoints
type MessageTraceHandlers struct {
	traceRepository *database.MessageTraceRepository
	queueService    *queue.Service
	logService      *logprocessor.Service
	repository      *database.Repository
}

// NewMessageTraceHandlers creates a new message trace handlers instance
func NewMessageTraceHandlers(repository *database.Repository, queueService *queue.Service, logService *logprocessor.Service) *MessageTraceHandlers {
	traceRepo := database.NewMessageTraceRepository(repository.GetDB())

	return &MessageTraceHandlers{
		traceRepository: traceRepo,
		queueService:    queueService,
		logService:      logService,
		repository:      repository,
	}
}

// handleMessageDeliveryTrace handles GET /api/v1/messages/{id}/delivery-trace - Enhanced message delivery tracing
func (h *MessageTraceHandlers) handleMessageDeliveryTrace(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Validate message ID format
	if err := h.validateMessageID(messageID); err != nil {
		WriteBadRequestResponse(w, "Invalid message ID format: "+err.Error())
		return
	}

	// Get comprehensive delivery trace
	trace, err := h.traceRepository.GetMessageDeliveryTrace(messageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Message not found")
		} else {
			WriteInternalErrorResponse(w, "Failed to generate delivery trace: "+err.Error())
		}
		return
	}

	WriteSuccessResponse(w, trace)
}

// handleRecipientDeliveryHistory handles GET /api/v1/messages/{id}/recipients/{recipient}/history - Per-recipient delivery history
func (h *MessageTraceHandlers) handleRecipientDeliveryHistory(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	recipient := GetPathParam(r, "recipient")

	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	if recipient == "" {
		WriteBadRequestResponse(w, "Recipient is required")
		return
	}

	// Get delivery trace
	trace, err := h.traceRepository.GetMessageDeliveryTrace(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get delivery trace: "+err.Error())
		return
	}

	// Find specific recipient
	var recipientStatus *database.RecipientDeliveryStatus
	for _, r := range trace.Recipients {
		if r.Recipient == recipient {
			recipientStatus = &r
			break
		}
	}

	if recipientStatus == nil {
		WriteNotFoundResponse(w, "Recipient not found for this message")
		return
	}

	// Filter timeline for this recipient
	var recipientTimeline []database.DeliveryTimelineEvent
	for _, event := range trace.DeliveryTimeline {
		if event.Recipient != nil && *event.Recipient == recipient {
			recipientTimeline = append(recipientTimeline, event)
		}
	}

	response := map[string]interface{}{
		"message_id":     messageID,
		"recipient":      recipient,
		"status":         recipientStatus,
		"timeline":       recipientTimeline,
		"retry_schedule": h.getRecipientRetrySchedule(trace.RetrySchedule, recipient),
	}

	WriteSuccessResponse(w, response)
}

// handleDeliveryTimeline handles GET /api/v1/messages/{id}/timeline - Delivery timeline visualization
func (h *MessageTraceHandlers) handleDeliveryTimeline(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get optional filters
	eventType := GetQueryParam(r, "event_type", "")
	recipient := GetQueryParam(r, "recipient", "")
	source := GetQueryParam(r, "source", "")

	// Get delivery trace
	trace, err := h.traceRepository.GetMessageDeliveryTrace(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get delivery trace: "+err.Error())
		return
	}

	// Apply filters
	timeline := trace.DeliveryTimeline
	if eventType != "" || recipient != "" || source != "" {
		timeline = h.filterTimeline(trace.DeliveryTimeline, eventType, recipient, source)
	}

	response := map[string]interface{}{
		"message_id":   messageID,
		"timeline":     timeline,
		"summary":      trace.Summary,
		"generated_at": trace.GeneratedAt,
	}

	WriteSuccessResponse(w, response)
}

// handleRetrySchedule handles GET /api/v1/messages/{id}/retry-schedule - Retry timeline visualization
func (h *MessageTraceHandlers) handleRetrySchedule(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get delivery trace
	trace, err := h.traceRepository.GetMessageDeliveryTrace(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get delivery trace: "+err.Error())
		return
	}

	// Get enhanced retry information from queue service if available
	var queueRetryInfo map[string]interface{}
	if h.queueService != nil {
		if details, err := h.queueService.GetMessageDetails(messageID); err == nil {
			queueRetryInfo = map[string]interface{}{
				"queue_status": details.Status,
				"retry_count":  details.RetryCount,
				"last_attempt": details.LastAttempt,
				"next_retry":   details.NextRetry,
			}
		}
	}

	response := map[string]interface{}{
		"message_id":     messageID,
		"retry_schedule": trace.RetrySchedule,
		"queue_info":     queueRetryInfo,
		"deferred_count": trace.Summary.DeferredCount,
		"generated_at":   trace.GeneratedAt,
	}

	WriteSuccessResponse(w, response)
}

// handleDeliveryAttemptDetails handles GET /api/v1/delivery-attempts/{id} - Detailed delivery attempt information
func (h *MessageTraceHandlers) handleDeliveryAttemptDetails(w http.ResponseWriter, r *http.Request) {
	attemptIDStr := GetPathParam(r, "id")
	if attemptIDStr == "" {
		WriteBadRequestResponse(w, "Delivery attempt ID is required")
		return
	}

	attemptID, err := strconv.ParseInt(attemptIDStr, 10, 64)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid delivery attempt ID")
		return
	}

	// Get delivery attempt details
	attempt, err := h.getDeliveryAttemptByID(attemptID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Delivery attempt not found")
		} else {
			WriteInternalErrorResponse(w, "Failed to get delivery attempt: "+err.Error())
		}
		return
	}

	// Get related log entries for context
	relatedLogs, err := h.getRelatedLogEntries(attempt.MessageID, attempt.Recipient, attempt.Timestamp)
	if err != nil {
		// Log error but don't fail the request
		relatedLogs = []database.LogEntry{}
	}

	response := map[string]interface{}{
		"attempt":      attempt,
		"related_logs": relatedLogs,
	}

	WriteSuccessResponse(w, response)
}

// handleMessageDeliveryStats handles GET /api/v1/messages/{id}/delivery-stats - Delivery statistics summary
func (h *MessageTraceHandlers) handleMessageDeliveryStats(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get delivery trace
	trace, err := h.traceRepository.GetMessageDeliveryTrace(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get delivery trace: "+err.Error())
		return
	}

	// Calculate additional statistics
	stats := h.calculateDeliveryStatistics(trace)

	WriteSuccessResponse(w, stats)
}

// Helper methods

// validateMessageID validates the format of a message ID
func (h *MessageTraceHandlers) validateMessageID(messageID string) error {
	if len(messageID) == 0 {
		return fmt.Errorf("message ID cannot be empty")
	}

	// Basic validation - Exim message IDs are typically alphanumeric with hyphens
	if len(messageID) < 6 || len(messageID) > 50 {
		return fmt.Errorf("message ID length must be between 6 and 50 characters")
	}

	return nil
}

// getRecipientRetrySchedule filters retry schedule for a specific recipient
func (h *MessageTraceHandlers) getRecipientRetrySchedule(schedule []database.RetryScheduleEntry, recipient string) []database.RetryScheduleEntry {
	var filtered []database.RetryScheduleEntry
	for _, entry := range schedule {
		if entry.Recipient == recipient {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// filterTimeline applies filters to the delivery timeline
func (h *MessageTraceHandlers) filterTimeline(timeline []database.DeliveryTimelineEvent, eventType, recipient, source string) []database.DeliveryTimelineEvent {
	var filtered []database.DeliveryTimelineEvent

	for _, event := range timeline {
		// Apply event type filter
		if eventType != "" && event.EventType != eventType {
			continue
		}

		// Apply recipient filter
		if recipient != "" && (event.Recipient == nil || *event.Recipient != recipient) {
			continue
		}

		// Apply source filter
		if source != "" && event.Source != source {
			continue
		}

		filtered = append(filtered, event)
	}

	return filtered
}

// getDeliveryAttemptByID retrieves a delivery attempt by ID
func (h *MessageTraceHandlers) getDeliveryAttemptByID(attemptID int64) (*database.DeliveryAttempt, error) {
	query := `
		SELECT id, message_id, recipient, timestamp, host, ip_address, status, smtp_code, error_message, created_at
		FROM delivery_attempts WHERE id = ?`

	attempt := &database.DeliveryAttempt{}
	err := h.repository.GetDB().QueryRow(query, attemptID).Scan(
		&attempt.ID, &attempt.MessageID, &attempt.Recipient, &attempt.Timestamp,
		&attempt.Host, &attempt.IPAddress, &attempt.Status, &attempt.SMTPCode,
		&attempt.ErrorMessage, &attempt.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("delivery attempt not found")
	}

	return attempt, err
}

// getRelatedLogEntries gets log entries related to a specific delivery attempt
func (h *MessageTraceHandlers) getRelatedLogEntries(messageID, recipient string, attemptTime time.Time) ([]database.LogEntry, error) {
	// Get log entries within a time window around the attempt
	startTime := attemptTime.Add(-5 * time.Minute)
	endTime := attemptTime.Add(5 * time.Minute)

	query := `
		SELECT id, timestamp, message_id, log_type, event, host, sender, recipients, size, status, error_code, error_text, raw_line, created_at
		FROM log_entries 
		WHERE message_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp`

	rows, err := h.repository.GetDB().Query(query, messageID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []database.LogEntry
	for rows.Next() {
		var entry database.LogEntry
		err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.MessageID, &entry.LogType, &entry.Event, &entry.Host, &entry.Sender, &entry.RecipientsDB, &entry.Size, &entry.Status, &entry.ErrorCode, &entry.ErrorText, &entry.RawLine, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal recipients from JSON
		if err := entry.UnmarshalRecipients(); err != nil {
			continue // Skip entries with invalid recipient data
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// calculateDeliveryStatistics calculates additional delivery statistics
func (h *MessageTraceHandlers) calculateDeliveryStatistics(trace *database.MessageDeliveryTrace) map[string]interface{} {
	stats := map[string]interface{}{
		"summary": trace.Summary,
	}

	// Calculate per-recipient statistics
	recipientStats := make(map[string]interface{})
	for _, recipient := range trace.Recipients {
		recipientStats[recipient.Recipient] = map[string]interface{}{
			"status":        recipient.Status,
			"attempt_count": recipient.AttemptCount,
			"delivered_at":  recipient.DeliveredAt,
			"last_attempt":  recipient.LastAttemptAt,
			"next_retry":    recipient.NextRetryAt,
		}
	}
	stats["recipients"] = recipientStats

	// Calculate timeline statistics
	eventCounts := make(map[string]int)
	sourceCounts := make(map[string]int)

	for _, event := range trace.DeliveryTimeline {
		eventCounts[event.EventType]++
		sourceCounts[event.Source]++
	}

	stats["timeline_stats"] = map[string]interface{}{
		"total_events":  len(trace.DeliveryTimeline),
		"event_counts":  eventCounts,
		"source_counts": sourceCounts,
	}

	// Calculate retry statistics
	stats["retry_stats"] = map[string]interface{}{
		"scheduled_retries": len(trace.RetrySchedule),
		"estimated_retries": h.countEstimatedRetries(trace.RetrySchedule),
	}

	return stats
}

// countEstimatedRetries counts how many retries are estimated vs actual
func (h *MessageTraceHandlers) countEstimatedRetries(schedule []database.RetryScheduleEntry) int {
	count := 0
	for _, entry := range schedule {
		if entry.IsEstimated {
			count++
		}
	}
	return count
}

// handleThreadedTimeline handles GET /api/v1/messages/{id}/threaded-timeline - Threaded delivery timeline view
func (h *MessageTraceHandlers) handleThreadedTimeline(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get threaded timeline view
	threadedView, err := h.traceRepository.GetThreadedTimelineView(messageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Message not found")
		} else {
			WriteInternalErrorResponse(w, "Failed to generate threaded timeline: "+err.Error())
		}
		return
	}

	WriteSuccessResponse(w, threadedView)
}

// handleMessageContent handles GET /api/v1/messages/{id}/content - Safe message content preview
func (h *MessageTraceHandlers) handleMessageContent(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get message content
	content, err := h.traceRepository.GetMessageContent(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get message content: "+err.Error())
		return
	}

	WriteSuccessResponse(w, content)
}

// handleMessageNotes handles GET /api/v1/messages/{id}/notes - Get message notes
func (h *MessageTraceHandlers) handleMessageNotes(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	noteRepo := database.NewMessageNoteRepository(h.repository.GetDB())
	notes, err := noteRepo.GetByMessageID(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get message notes: "+err.Error())
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message_id": messageID,
		"notes":      notes,
	})
}

// handleCreateMessageNote handles POST /api/v1/messages/{id}/notes - Create message note
func (h *MessageTraceHandlers) handleCreateMessageNote(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	var noteRequest struct {
		Note     string `json:"note"`
		IsPublic bool   `json:"is_public"`
	}

	if err := ParseJSONBody(r, &noteRequest); err != nil {
		WriteBadRequestResponse(w, "Invalid JSON: "+err.Error())
		return
	}

	if strings.TrimSpace(noteRequest.Note) == "" {
		WriteBadRequestResponse(w, "Note content is required")
		return
	}

	userID := h.getUserID(r)

	note := &database.MessageNote{
		MessageID: messageID,
		UserID:    userID,
		Note:      strings.TrimSpace(noteRequest.Note),
		IsPublic:  noteRequest.IsPublic,
	}

	noteRepo := database.NewMessageNoteRepository(h.repository.GetDB())
	if err := noteRepo.Create(note); err != nil {
		WriteInternalErrorResponse(w, "Failed to create note: "+err.Error())
		return
	}

	WriteSuccessResponse(w, note)
}

// handleUpdateMessageNote handles PUT /api/v1/messages/{id}/notes/{noteId} - Update message note
func (h *MessageTraceHandlers) handleUpdateMessageNote(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	noteIDStr := GetPathParam(r, "noteId")

	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	if noteIDStr == "" {
		WriteBadRequestResponse(w, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid note ID")
		return
	}

	var noteRequest struct {
		Note     string `json:"note"`
		IsPublic bool   `json:"is_public"`
	}

	if err := ParseJSONBody(r, &noteRequest); err != nil {
		WriteBadRequestResponse(w, "Invalid JSON: "+err.Error())
		return
	}

	if strings.TrimSpace(noteRequest.Note) == "" {
		WriteBadRequestResponse(w, "Note content is required")
		return
	}

	userID := h.getUserID(r)

	note := &database.MessageNote{
		ID:        noteID,
		MessageID: messageID,
		UserID:    userID,
		Note:      strings.TrimSpace(noteRequest.Note),
		IsPublic:  noteRequest.IsPublic,
	}

	noteRepo := database.NewMessageNoteRepository(h.repository.GetDB())
	if err := noteRepo.Update(note); err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Note not found or not owned by user")
		} else {
			WriteInternalErrorResponse(w, "Failed to update note: "+err.Error())
		}
		return
	}

	WriteSuccessResponse(w, note)
}

// handleDeleteMessageNote handles DELETE /api/v1/messages/{id}/notes/{noteId} - Delete message note
func (h *MessageTraceHandlers) handleDeleteMessageNote(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	noteIDStr := GetPathParam(r, "noteId")

	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	if noteIDStr == "" {
		WriteBadRequestResponse(w, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid note ID")
		return
	}

	userID := h.getUserID(r)

	noteRepo := database.NewMessageNoteRepository(h.repository.GetDB())
	if err := noteRepo.Delete(noteID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Note not found or not owned by user")
		} else {
			WriteInternalErrorResponse(w, "Failed to delete note: "+err.Error())
		}
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message": "Note deleted successfully",
	})
}

// handleMessageTags handles GET /api/v1/messages/{id}/tags - Get message tags
func (h *MessageTraceHandlers) handleMessageTags(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	tagRepo := database.NewMessageTagRepository(h.repository.GetDB())
	tags, err := tagRepo.GetByMessageID(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get message tags: "+err.Error())
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message_id": messageID,
		"tags":       tags,
	})
}

// handleCreateMessageTag handles POST /api/v1/messages/{id}/tags - Create message tag
func (h *MessageTraceHandlers) handleCreateMessageTag(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	var tagRequest struct {
		Tag   string  `json:"tag"`
		Color *string `json:"color,omitempty"`
	}

	if err := ParseJSONBody(r, &tagRequest); err != nil {
		WriteBadRequestResponse(w, "Invalid JSON: "+err.Error())
		return
	}

	if strings.TrimSpace(tagRequest.Tag) == "" {
		WriteBadRequestResponse(w, "Tag is required")
		return
	}

	userID := h.getUserID(r)

	tag := &database.MessageTag{
		MessageID: messageID,
		Tag:       strings.TrimSpace(tagRequest.Tag),
		Color:     tagRequest.Color,
		UserID:    userID,
	}

	tagRepo := database.NewMessageTagRepository(h.repository.GetDB())
	if err := tagRepo.Create(tag); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			WriteBadRequestResponse(w, "Tag already exists for this message")
		} else {
			WriteInternalErrorResponse(w, "Failed to create tag: "+err.Error())
		}
		return
	}

	WriteSuccessResponse(w, tag)
}

// handleDeleteMessageTag handles DELETE /api/v1/messages/{id}/tags/{tagId} - Delete message tag
func (h *MessageTraceHandlers) handleDeleteMessageTag(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	tagIDStr := GetPathParam(r, "tagId")

	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	if tagIDStr == "" {
		WriteBadRequestResponse(w, "Tag ID is required")
		return
	}

	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid tag ID")
		return
	}

	userID := h.getUserID(r)

	tagRepo := database.NewMessageTagRepository(h.repository.GetDB())
	if err := tagRepo.Delete(tagID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Tag not found or not owned by user")
		} else {
			WriteInternalErrorResponse(w, "Failed to delete tag: "+err.Error())
		}
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message": "Tag deleted successfully",
	})
}

// handlePopularTags handles GET /api/v1/tags/popular - Get popular tags
func (h *MessageTraceHandlers) handlePopularTags(w http.ResponseWriter, r *http.Request) {
	limit, err := GetQueryParamInt(r, "limit", 20)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid limit parameter")
		return
	}

	if limit > 100 {
		limit = 100
	}

	tagRepo := database.NewMessageTagRepository(h.repository.GetDB())
	tags, err := tagRepo.GetPopularTags(limit)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to get popular tags: "+err.Error())
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"tags": tags,
	})
}

// getUserID extracts user ID from request context
func (h *MessageTraceHandlers) getUserID(r *http.Request) string {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		return "anonymous"
	}
	return fmt.Sprintf("%d", user.ID)
}
