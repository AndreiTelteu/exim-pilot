package queue

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/exim-pilot/internal/database"
)

// OperationResult represents the result of a queue operation
type OperationResult struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	TotalMessages   int               `json:"total_messages"`
	SuccessfulCount int               `json:"successful_count"`
	FailedCount     int               `json:"failed_count"`
	Results         []OperationResult `json:"results"`
	Operation       string            `json:"operation"`
}

// Operations interface defines queue operation methods
type Operations interface {
	DeliverNow(messageID string, userID string, ipAddress string) (*OperationResult, error)
	FreezeMessage(messageID string, userID string, ipAddress string) (*OperationResult, error)
	ThawMessage(messageID string, userID string, ipAddress string) (*OperationResult, error)
	DeleteMessage(messageID string, userID string, ipAddress string) (*OperationResult, error)
	BulkDeliverNow(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error)
	BulkFreeze(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error)
	BulkThaw(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error)
	BulkDelete(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error)
}

// DeliverNow forces immediate delivery of a message using exim -M
func (m *Manager) DeliverNow(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	result := &OperationResult{
		MessageID: messageID,
		Operation: "deliver_now",
	}

	// Execute exim -M command
	cmd := exec.Command(m.eximPath, "-M", messageID)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Command failed: %v", err)
		result.Message = string(output)
	} else {
		result.Success = true
		result.Message = "Delivery attempt initiated"
	}

	// Log the operation in audit trail
	if err := m.logAuditAction("deliver_now", messageID, userID, ipAddress, result); err != nil {
		// Don't fail the operation if audit logging fails, just log the error
		fmt.Printf("Failed to log audit action: %v\n", err)
	}

	return result, nil
}

// FreezeMessage freezes a message using exim -Mf
func (m *Manager) FreezeMessage(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	result := &OperationResult{
		MessageID: messageID,
		Operation: "freeze",
	}

	// Execute exim -Mf command
	cmd := exec.Command(m.eximPath, "-Mf", messageID)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Command failed: %v", err)
		result.Message = string(output)
	} else {
		result.Success = true
		result.Message = "Message frozen successfully"
	}

	// Log the operation in audit trail
	if err := m.logAuditAction("freeze", messageID, userID, ipAddress, result); err != nil {
		fmt.Printf("Failed to log audit action: %v\n", err)
	}

	return result, nil
}

// ThawMessage thaws a frozen message using exim -Mt
func (m *Manager) ThawMessage(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	result := &OperationResult{
		MessageID: messageID,
		Operation: "thaw",
	}

	// Execute exim -Mt command
	cmd := exec.Command(m.eximPath, "-Mt", messageID)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Command failed: %v", err)
		result.Message = string(output)
	} else {
		result.Success = true
		result.Message = "Message thawed successfully"
	}

	// Log the operation in audit trail
	if err := m.logAuditAction("thaw", messageID, userID, ipAddress, result); err != nil {
		fmt.Printf("Failed to log audit action: %v\n", err)
	}

	return result, nil
}

// DeleteMessage removes a message from the queue using exim -Mrm
func (m *Manager) DeleteMessage(messageID string, userID string, ipAddress string) (*OperationResult, error) {
	result := &OperationResult{
		MessageID: messageID,
		Operation: "delete",
	}

	// Execute exim -Mrm command
	cmd := exec.Command(m.eximPath, "-Mrm", messageID)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Command failed: %v", err)
		result.Message = string(output)
	} else {
		result.Success = true
		result.Message = "Message deleted successfully"
	}

	// Log the operation in audit trail
	if err := m.logAuditAction("delete", messageID, userID, ipAddress, result); err != nil {
		fmt.Printf("Failed to log audit action: %v\n", err)
	}

	return result, nil
}

// BulkDeliverNow performs deliver now operation on multiple messages
func (m *Manager) BulkDeliverNow(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return m.performBulkOperation(messageIDs, "deliver_now", userID, ipAddress, m.DeliverNow)
}

// BulkFreeze performs freeze operation on multiple messages
func (m *Manager) BulkFreeze(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return m.performBulkOperation(messageIDs, "freeze", userID, ipAddress, m.FreezeMessage)
}

// BulkThaw performs thaw operation on multiple messages
func (m *Manager) BulkThaw(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return m.performBulkOperation(messageIDs, "thaw", userID, ipAddress, m.ThawMessage)
}

// BulkDelete performs delete operation on multiple messages
func (m *Manager) BulkDelete(messageIDs []string, userID string, ipAddress string) (*BulkOperationResult, error) {
	return m.performBulkOperation(messageIDs, "delete", userID, ipAddress, m.DeleteMessage)
}

// performBulkOperation is a helper function for bulk operations
func (m *Manager) performBulkOperation(
	messageIDs []string,
	operation string,
	userID string,
	ipAddress string,
	operationFunc func(string, string, string) (*OperationResult, error),
) (*BulkOperationResult, error) {

	bulkResult := &BulkOperationResult{
		TotalMessages: len(messageIDs),
		Operation:     operation,
		Results:       make([]OperationResult, 0, len(messageIDs)),
	}

	// Perform operation on each message
	for _, messageID := range messageIDs {
		result, err := operationFunc(messageID, userID, ipAddress)
		if err != nil {
			// Create error result if operation function failed
			result = &OperationResult{
				Success:   false,
				MessageID: messageID,
				Operation: operation,
				Error:     err.Error(),
			}
		}

		bulkResult.Results = append(bulkResult.Results, *result)

		if result.Success {
			bulkResult.SuccessfulCount++
		} else {
			bulkResult.FailedCount++
		}
	}

	// Log bulk operation in audit trail
	if err := m.logBulkAuditAction(operation, messageIDs, userID, ipAddress, bulkResult); err != nil {
		fmt.Printf("Failed to log bulk audit action: %v\n", err)
	}

	return bulkResult, nil
}

// logAuditAction logs a single queue operation to the audit trail
func (m *Manager) logAuditAction(action, messageID, userID, ipAddress string, result *OperationResult) error {
	if m.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Create audit log entry
	auditEntry := &database.AuditLog{
		Action:    fmt.Sprintf("queue_%s", action),
		MessageID: &messageID,
		UserID:    &userID,
		IPAddress: &ipAddress,
	}

	// Add operation details as JSON
	details := map[string]interface{}{
		"operation":  action,
		"message_id": messageID,
		"success":    result.Success,
		"message":    result.Message,
	}

	if result.Error != "" {
		details["error"] = result.Error
	}

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal audit details: %w", err)
	}

	detailsStr := string(detailsJSON)
	auditEntry.Details = &detailsStr

	// Save to database
	repo := database.NewAuditLogRepository(m.db)
	return repo.Create(auditEntry)
}

// logBulkAuditAction logs a bulk queue operation to the audit trail
func (m *Manager) logBulkAuditAction(action string, messageIDs []string, userID, ipAddress string, result *BulkOperationResult) error {
	if m.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Create audit log entry for bulk operation
	auditEntry := &database.AuditLog{
		Action:    fmt.Sprintf("queue_bulk_%s", action),
		UserID:    &userID,
		IPAddress: &ipAddress,
	}

	// Add bulk operation details as JSON
	details := map[string]interface{}{
		"operation":        fmt.Sprintf("bulk_%s", action),
		"message_ids":      messageIDs,
		"total_messages":   result.TotalMessages,
		"successful_count": result.SuccessfulCount,
		"failed_count":     result.FailedCount,
	}

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal bulk audit details: %w", err)
	}

	detailsStr := string(detailsJSON)
	auditEntry.Details = &detailsStr

	// Save to database
	repo := database.NewAuditLogRepository(m.db)
	return repo.Create(auditEntry)
}

// ValidateMessageID checks if a message ID is valid format
func (m *Manager) ValidateMessageID(messageID string) error {
	if messageID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}

	// Basic validation for Exim message ID format
	// Exim message IDs are typically in format: XXXXXX-XXXXXX-XX
	if len(messageID) < 10 {
		return fmt.Errorf("message ID too short: %s", messageID)
	}

	// Additional validation could be added here
	return nil
}

// GetOperationHistory retrieves the operation history for a message
func (m *Manager) GetOperationHistory(messageID string) ([]database.AuditLog, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	repo := database.NewAuditLogRepository(m.db)

	// Get audit logs for this message
	// Note: This is a simplified query - in practice you'd want to filter by message_id
	logs, err := repo.List(100, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve operation history: %w", err)
	}

	// Filter logs for this specific message
	var messageLogs []database.AuditLog
	for _, log := range logs {
		if log.MessageID != nil && *log.MessageID == messageID {
			messageLogs = append(messageLogs, log)
		}
	}

	return messageLogs, nil
}

// GetRecentOperations retrieves recent queue operations
func (m *Manager) GetRecentOperations(limit int) ([]database.AuditLog, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	repo := database.NewAuditLogRepository(m.db)

	// Get recent audit logs with queue operations
	logs, err := repo.List(limit, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve recent operations: %w", err)
	}

	// Filter for queue operations only
	var queueLogs []database.AuditLog
	for _, log := range logs {
		if isQueueOperation(log.Action) {
			queueLogs = append(queueLogs, log)
		}
	}

	return queueLogs, nil
}

// isQueueOperation checks if an action is a queue operation
func isQueueOperation(action string) bool {
	queueActions := []string{
		"queue_deliver_now",
		"queue_freeze",
		"queue_thaw",
		"queue_delete",
		"queue_bulk_deliver_now",
		"queue_bulk_freeze",
		"queue_bulk_thaw",
		"queue_bulk_delete",
	}

	for _, queueAction := range queueActions {
		if action == queueAction {
			return true
		}
	}
	return false
}
