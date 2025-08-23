package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/andreitelteu/exim-pilot/internal/queue"
)

// QueueHandlers contains handlers for queue management endpoints
type QueueHandlers struct {
	queueService *queue.Service
}

// NewQueueHandlers creates a new queue handlers instance
func NewQueueHandlers(queueService *queue.Service) *QueueHandlers {
	return &QueueHandlers{
		queueService: queueService,
	}
}

// handleQueueList handles GET /api/v1/queue - List queue messages with pagination
func (h *QueueHandlers) handleQueueList(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	page, perPage, err := GetPaginationParams(r)
	if err != nil {
		WriteBadRequestResponse(w, err.Error())
		return
	}

	// Get queue status
	status, err := h.queueService.GetQueueStatus()
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve queue status")
		return
	}

	// Apply pagination
	total := len(status.Messages)
	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedMessages := status.Messages[start:end]

	// Create response with metadata
	response := map[string]interface{}{
		"messages":           paginatedMessages,
		"total_messages":     status.TotalMessages,
		"deferred_messages":  status.DeferredMessages,
		"frozen_messages":    status.FrozenMessages,
		"oldest_message_age": status.OldestMessageAge.String(),
	}

	meta := CalculatePagination(page, perPage, total)
	WriteSuccessResponseWithMeta(w, response, meta)
}

// handleQueueSearch handles POST /api/v1/queue/search - Search queue messages
func (h *QueueHandlers) handleQueueSearch(w http.ResponseWriter, r *http.Request) {
	var searchRequest struct {
		Criteria queue.SearchCriteria `json:"criteria"`
		Page     int                  `json:"page"`
		PerPage  int                  `json:"per_page"`
	}

	// Parse request body
	if err := ParseJSONBody(r, &searchRequest); err != nil {
		WriteBadRequestResponse(w, "Invalid JSON: "+err.Error())
		return
	}

	// Set default pagination if not provided
	if searchRequest.Page <= 0 {
		searchRequest.Page = 1
	}
	if searchRequest.PerPage <= 0 {
		searchRequest.PerPage = 50
	}

	// Perform search
	messages, err := h.queueService.SearchQueueMessages(&searchRequest.Criteria)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to search queue messages")
		return
	}

	// Apply pagination to search results
	total := len(messages)
	start := (searchRequest.Page - 1) * searchRequest.PerPage
	end := start + searchRequest.PerPage

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedMessages := messages[start:end]

	// Create response
	response := map[string]interface{}{
		"messages": paginatedMessages,
		"criteria": searchRequest.Criteria,
	}

	meta := CalculatePagination(searchRequest.Page, searchRequest.PerPage, total)
	WriteSuccessResponseWithMeta(w, response, meta)
}

// handleQueueDetails handles GET /api/v1/queue/{id} - Get message details
func (h *QueueHandlers) handleQueueDetails(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Validate message ID
	if err := h.queueService.ValidateMessageID(messageID); err != nil {
		WriteBadRequestResponse(w, "Invalid message ID: "+err.Error())
		return
	}

	// Get message details
	details, err := h.queueService.GetMessageDetails(messageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteNotFoundResponse(w, "Message not found")
		} else {
			WriteInternalErrorResponse(w, "Failed to retrieve message details")
		}
		return
	}

	WriteSuccessResponse(w, details)
}

// handleQueueDeliver handles POST /api/v1/queue/{id}/deliver - Force delivery
func (h *QueueHandlers) handleQueueDeliver(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get user context (placeholder - would be from authentication middleware)
	userID := h.getUserID(r)
	ipAddress := h.getClientIP(r)

	// Perform deliver now operation
	result, err := h.queueService.DeliverNow(messageID, userID, ipAddress)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to deliver message: "+err.Error())
		return
	}

	if result.Success {
		WriteSuccessResponse(w, result)
	} else {
		WriteBadRequestResponse(w, result.Error)
	}
}

// handleQueueFreeze handles POST /api/v1/queue/{id}/freeze - Freeze message
func (h *QueueHandlers) handleQueueFreeze(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	userID := h.getUserID(r)
	ipAddress := h.getClientIP(r)

	result, err := h.queueService.FreezeMessage(messageID, userID, ipAddress)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to freeze message: "+err.Error())
		return
	}

	if result.Success {
		WriteSuccessResponse(w, result)
	} else {
		WriteBadRequestResponse(w, result.Error)
	}
}

// handleQueueThaw handles POST /api/v1/queue/{id}/thaw - Thaw message
func (h *QueueHandlers) handleQueueThaw(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	userID := h.getUserID(r)
	ipAddress := h.getClientIP(r)

	result, err := h.queueService.ThawMessage(messageID, userID, ipAddress)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to thaw message: "+err.Error())
		return
	}

	if result.Success {
		WriteSuccessResponse(w, result)
	} else {
		WriteBadRequestResponse(w, result.Error)
	}
}

// handleQueueDelete handles DELETE /api/v1/queue/{id} - Delete message
func (h *QueueHandlers) handleQueueDelete(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	userID := h.getUserID(r)
	ipAddress := h.getClientIP(r)

	result, err := h.queueService.DeleteMessage(messageID, userID, ipAddress)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to delete message: "+err.Error())
		return
	}

	if result.Success {
		WriteSuccessResponse(w, result)
	} else {
		WriteBadRequestResponse(w, result.Error)
	}
}

// handleQueueBulk handles POST /api/v1/queue/bulk - Bulk operations
func (h *QueueHandlers) handleQueueBulk(w http.ResponseWriter, r *http.Request) {
	var bulkRequest struct {
		Operation  string   `json:"operation"`
		MessageIDs []string `json:"message_ids"`
	}

	if err := ParseJSONBody(r, &bulkRequest); err != nil {
		WriteBadRequestResponse(w, "Invalid JSON: "+err.Error())
		return
	}

	// Validate request
	if bulkRequest.Operation == "" {
		WriteBadRequestResponse(w, "Operation is required")
		return
	}

	if len(bulkRequest.MessageIDs) == 0 {
		WriteBadRequestResponse(w, "Message IDs are required")
		return
	}

	// Validate message IDs
	for _, messageID := range bulkRequest.MessageIDs {
		if err := h.queueService.ValidateMessageID(messageID); err != nil {
			WriteBadRequestResponse(w, "Invalid message ID '"+messageID+"': "+err.Error())
			return
		}
	}

	userID := h.getUserID(r)
	ipAddress := h.getClientIP(r)

	// Perform bulk operation based on operation type
	var result *queue.BulkOperationResult
	var err error

	switch bulkRequest.Operation {
	case "deliver":
		result, err = h.queueService.BulkDeliverNow(bulkRequest.MessageIDs, userID, ipAddress)
	case "freeze":
		result, err = h.queueService.BulkFreeze(bulkRequest.MessageIDs, userID, ipAddress)
	case "thaw":
		result, err = h.queueService.BulkThaw(bulkRequest.MessageIDs, userID, ipAddress)
	case "delete":
		result, err = h.queueService.BulkDelete(bulkRequest.MessageIDs, userID, ipAddress)
	default:
		WriteBadRequestResponse(w, "Invalid operation. Supported operations: deliver, freeze, thaw, delete")
		return
	}

	if err != nil {
		WriteInternalErrorResponse(w, "Failed to perform bulk operation: "+err.Error())
		return
	}

	WriteSuccessResponse(w, result)
}

// Helper methods

// getUserID extracts user ID from request context (placeholder implementation)
func (h *QueueHandlers) getUserID(r *http.Request) string {
	// In a real implementation, this would extract the user ID from JWT token or session
	// For now, return a placeholder
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	return "anonymous"
}

// getClientIP extracts client IP address from request
func (h *QueueHandlers) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// Additional queue endpoints

// handleQueueHealth handles GET /api/v1/queue/health - Get queue health metrics
func (h *QueueHandlers) handleQueueHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.queueService.GetQueueHealth()
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve queue health")
		return
	}

	WriteSuccessResponse(w, health)
}

// handleQueueStatistics handles GET /api/v1/queue/statistics - Get detailed queue statistics
func (h *QueueHandlers) handleQueueStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.queueService.GetQueueStatistics()
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve queue statistics")
		return
	}

	WriteSuccessResponse(w, stats)
}

// handleQueueHistory handles GET /api/v1/queue/{id}/history - Get operation history for a message
func (h *QueueHandlers) handleQueueHistory(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get limit parameter
	limitStr := GetQueryParam(r, "limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	history, err := h.queueService.GetOperationHistory(messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve operation history")
		return
	}

	// Apply limit
	if len(history) > limit {
		history = history[:limit]
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message_id": messageID,
		"history":    history,
	})
}
