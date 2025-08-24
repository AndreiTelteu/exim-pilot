package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// Service handles audit logging for all administrative actions
type Service struct {
	repository *database.Repository
}

// NewService creates a new audit service instance
func NewService(repository *database.Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ActionType represents the type of administrative action
type ActionType string

const (
	// Queue operations
	ActionQueueDeliver ActionType = "queue_deliver"
	ActionQueueFreeze  ActionType = "queue_freeze"
	ActionQueueThaw    ActionType = "queue_thaw"
	ActionQueueDelete  ActionType = "queue_delete"
	ActionBulkDeliver  ActionType = "bulk_deliver"
	ActionBulkFreeze   ActionType = "bulk_freeze"
	ActionBulkThaw     ActionType = "bulk_thaw"
	ActionBulkDelete   ActionType = "bulk_delete"

	// Authentication actions
	ActionLogin  ActionType = "login"
	ActionLogout ActionType = "logout"

	// Message operations
	ActionMessageView    ActionType = "message_view"
	ActionMessageContent ActionType = "message_content"
	ActionNoteCreate     ActionType = "note_create"
	ActionNoteUpdate     ActionType = "note_update"
	ActionNoteDelete     ActionType = "note_delete"
	ActionTagCreate      ActionType = "tag_create"
	ActionTagDelete      ActionType = "tag_delete"

	// System operations
	ActionConfigChange ActionType = "config_change"
	ActionSystemAccess ActionType = "system_access"
)

// AuditContext contains context information for audit logging
type AuditContext struct {
	UserID    string
	IPAddress string
	UserAgent string
	RequestID string
}

// AuditDetails contains detailed information about the action
type AuditDetails struct {
	MessageIDs    []string               `json:"message_ids,omitempty"`
	Recipients    []string               `json:"recipients,omitempty"`
	Operation     string                 `json:"operation,omitempty"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
	Result        string                 `json:"result,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	ResourcePath  string                 `json:"resource_path,omitempty"`
	PreviousValue interface{}            `json:"previous_value,omitempty"`
	NewValue      interface{}            `json:"new_value,omitempty"`
}

// LogAction logs an administrative action to the audit trail
func (s *Service) LogAction(ctx context.Context, action ActionType, messageID *string, auditCtx *AuditContext, details *AuditDetails) error {
	// Serialize details to JSON
	var detailsJSON *string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			log.Printf("Failed to marshal audit details: %v", err)
			// Continue with logging without details rather than failing
		} else {
			detailsStr := string(detailsBytes)
			detailsJSON = &detailsStr
		}
	}

	// Create audit log entry
	auditEntry := &database.AuditLog{
		Timestamp: time.Now().UTC(),
		Action:    string(action),
		MessageID: messageID,
		UserID:    &auditCtx.UserID,
		Details:   detailsJSON,
		IPAddress: &auditCtx.IPAddress,
		CreatedAt: time.Now().UTC(),
	}

	// Store in database (immutable - no updates allowed)
	err := s.repository.CreateAuditLog(auditEntry)
	if err != nil {
		log.Printf("Failed to create audit log entry: %v", err)
		return fmt.Errorf("failed to create audit log entry: %w", err)
	}

	// Log to system log as well for redundancy
	s.logToSystemLog(action, messageID, auditCtx, details)

	return nil
}

// LogQueueOperation logs queue management operations
func (s *Service) LogQueueOperation(ctx context.Context, operation string, messageID string, auditCtx *AuditContext, success bool, errorMsg string) error {
	var action ActionType
	switch operation {
	case "deliver":
		action = ActionQueueDeliver
	case "freeze":
		action = ActionQueueFreeze
	case "thaw":
		action = ActionQueueThaw
	case "delete":
		action = ActionQueueDelete
	default:
		action = ActionType("queue_" + operation)
	}

	details := &AuditDetails{
		MessageIDs: []string{messageID},
		Operation:  operation,
		Result:     "success",
	}

	if !success {
		details.Result = "failure"
		details.ErrorMessage = errorMsg
	}

	return s.LogAction(ctx, action, &messageID, auditCtx, details)
}

// LogBulkOperation logs bulk queue operations
func (s *Service) LogBulkOperation(ctx context.Context, operation string, messageIDs []string, auditCtx *AuditContext, successCount, failureCount int, errors []string) error {
	var action ActionType
	switch operation {
	case "deliver":
		action = ActionBulkDeliver
	case "freeze":
		action = ActionBulkFreeze
	case "thaw":
		action = ActionBulkThaw
	case "delete":
		action = ActionBulkDelete
	default:
		action = ActionType("bulk_" + operation)
	}

	details := &AuditDetails{
		MessageIDs: messageIDs,
		Operation:  operation,
		Parameters: map[string]interface{}{
			"total_messages": len(messageIDs),
			"success_count":  successCount,
			"failure_count":  failureCount,
			"success_rate":   float64(successCount) / float64(len(messageIDs)),
		},
	}

	if len(errors) > 0 {
		details.ErrorMessage = fmt.Sprintf("Bulk operation completed with %d failures", failureCount)
		details.Parameters["errors"] = errors
	}

	if successCount == len(messageIDs) {
		details.Result = "success"
	} else if successCount > 0 {
		details.Result = "partial_success"
	} else {
		details.Result = "failure"
	}

	return s.LogAction(ctx, action, nil, auditCtx, details)
}

// LogAuthentication logs authentication events
func (s *Service) LogAuthentication(ctx context.Context, action ActionType, username string, auditCtx *AuditContext, success bool, errorMsg string) error {
	details := &AuditDetails{
		Parameters: map[string]interface{}{
			"username": username,
		},
		Result: "success",
	}

	if !success {
		details.Result = "failure"
		details.ErrorMessage = errorMsg
	}

	return s.LogAction(ctx, action, nil, auditCtx, details)
}

// LogMessageAccess logs message content access
func (s *Service) LogMessageAccess(ctx context.Context, messageID string, accessType string, auditCtx *AuditContext) error {
	var action ActionType
	switch accessType {
	case "content":
		action = ActionMessageContent
	case "view":
		action = ActionMessageView
	default:
		action = ActionMessageView
	}

	details := &AuditDetails{
		MessageIDs:   []string{messageID},
		Operation:    accessType,
		ResourcePath: fmt.Sprintf("/messages/%s/%s", messageID, accessType),
	}

	return s.LogAction(ctx, action, &messageID, auditCtx, details)
}

// LogSystemAccess logs system-level access attempts
func (s *Service) LogSystemAccess(ctx context.Context, resourcePath string, auditCtx *AuditContext, success bool, errorMsg string) error {
	details := &AuditDetails{
		ResourcePath: resourcePath,
		Result:       "success",
	}

	if !success {
		details.Result = "failure"
		details.ErrorMessage = errorMsg
	}

	return s.LogAction(ctx, ActionSystemAccess, nil, auditCtx, details)
}

// GetAuditTrail retrieves audit log entries with filtering
func (s *Service) GetAuditTrail(ctx context.Context, filters *AuditFilters) ([]*database.AuditLog, error) {
	return s.repository.GetAuditLogs(filters)
}

// AuditFilters contains filtering options for audit log retrieval
type AuditFilters struct {
	StartTime *time.Time
	EndTime   *time.Time
	UserID    *string
	Action    *string
	MessageID *string
	IPAddress *string
	Limit     int
	Offset    int
}

// logToSystemLog writes audit events to system log for redundancy
func (s *Service) logToSystemLog(action ActionType, messageID *string, auditCtx *AuditContext, details *AuditDetails) {
	logEntry := fmt.Sprintf("AUDIT: action=%s user=%s ip=%s", action, auditCtx.UserID, auditCtx.IPAddress)

	if messageID != nil {
		logEntry += fmt.Sprintf(" message_id=%s", *messageID)
	}

	if details != nil && details.Result != "" {
		logEntry += fmt.Sprintf(" result=%s", details.Result)
	}

	if details != nil && details.ErrorMessage != "" {
		logEntry += fmt.Sprintf(" error=%s", details.ErrorMessage)
	}

	log.Printf(logEntry)
}

// ValidateAuditIntegrity performs basic integrity checks on audit logs
func (s *Service) ValidateAuditIntegrity(ctx context.Context) error {
	// Check for gaps in audit log sequence
	// Check for suspicious patterns
	// This is a placeholder for more sophisticated integrity checks
	return nil
}
