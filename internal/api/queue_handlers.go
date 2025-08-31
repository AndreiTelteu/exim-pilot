package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/queue"
	"github.com/andreitelteu/exim-pilot/internal/validation"
	"github.com/andreitelteu/exim-pilot/internal/websocket"
)

// QueueHandlers contains handlers for queue management endpoints
type QueueHandlers struct {
	queueService      *queue.Service
	validationService *validation.Service
	wsService         *websocket.Service
}

// NewQueueHandlers creates a new queue handlers instance
func NewQueueHandlers(queueService *queue.Service, wsService *websocket.Service) *QueueHandlers {
	return &QueueHandlers{
		queueService:      queueService,
		validationService: validation.NewService(),
		wsService:         wsService,
	}
}

// handleQueueList handles GET /api/v1/queue - List queue messages with pagination and search
func (h *QueueHandlers) handleQueueList(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	page, perPage, err := GetPaginationParams(r)
	if err != nil {
		WriteBadRequestResponse(w, err.Error())
		return
	}

	// Parse search criteria from query parameters
	criteria := h.parseSearchCriteria(r)

	var messages []queue.QueueMessage
	var queueStatus *queue.QueueStatus

	// Check if any search criteria are provided
	if h.hasSearchCriteria(criteria) {
		// Use search functionality
		searchResults, err := h.queueService.SearchQueueMessages(criteria)
		if err != nil {
			WriteInternalErrorResponse(w, "Failed to search queue messages")
			return
		}
		messages = searchResults

		// Get basic queue stats for metadata
		queueStatus, err = h.queueService.GetQueueStatus()
		if err != nil {
			WriteInternalErrorResponse(w, "Failed to retrieve queue status")
			return
		}
	} else {
		// Get all messages
		queueStatus, err = h.queueService.GetQueueStatus()
		if err != nil {
			WriteInternalErrorResponse(w, "Failed to retrieve queue status")
			return
		}
		messages = queueStatus.Messages
	}

	// Apply pagination
	total := len(messages)
	start := (page - 1) * perPage
	end := start + perPage

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedMessages := messages[start:end]

	// Create response with metadata
	response := map[string]interface{}{
		"messages":           paginatedMessages,
		"total_messages":     queueStatus.TotalMessages,
		"deferred_messages":  queueStatus.DeferredMessages,
		"frozen_messages":    queueStatus.FrozenMessages,
		"oldest_message_age": queueStatus.OldestMessageAge.String(),
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

	// Validate message ID using validation service
	if err := h.validationService.ValidateMessageID(messageID); err != nil {
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
		// Broadcast queue update to WebSocket clients
		if h.wsService != nil {
			h.wsService.BroadcastQueueUpdate(map[string]interface{}{
				"action":     "deliver",
				"message_id": messageID,
				"status":     "success",
				"timestamp":  time.Now().UTC(),
			})
		}
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
		// Broadcast queue update to WebSocket clients
		if h.wsService != nil {
			h.wsService.BroadcastQueueUpdate(map[string]interface{}{
				"action":     "freeze",
				"message_id": messageID,
				"status":     "success",
				"timestamp":  time.Now().UTC(),
			})
		}
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
		// Broadcast queue update to WebSocket clients
		if h.wsService != nil {
			h.wsService.BroadcastQueueUpdate(map[string]interface{}{
				"action":     "thaw",
				"message_id": messageID,
				"status":     "success",
				"timestamp":  time.Now().UTC(),
			})
		}
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
		// Broadcast queue update to WebSocket clients
		if h.wsService != nil {
			h.wsService.BroadcastQueueUpdate(map[string]interface{}{
				"action":     "delete",
				"message_id": messageID,
				"status":     "success",
				"timestamp":  time.Now().UTC(),
			})
		}
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

	// Validate bulk request using validation service
	if err := h.validationService.ValidateBulkRequest(bulkRequest.Operation, bulkRequest.MessageIDs); err != nil {
		if validationErrors, ok := err.(*validation.ValidationErrors); ok {
			response := APIResponse{
				Success: false,
				Error:   "Validation failed",
				Data:    validationErrors.Errors,
			}
			WriteJSONResponse(w, http.StatusBadRequest, response)
			return
		}
		WriteBadRequestResponse(w, err.Error())
		return
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

	// Broadcast queue update to WebSocket clients for bulk operations
	if h.wsService != nil && result != nil {
		h.wsService.BroadcastQueueUpdate(map[string]interface{}{
			"action":      "bulk_" + bulkRequest.Operation,
			"message_ids": bulkRequest.MessageIDs,
			"status":      "completed",
			"result":      result,
			"timestamp":   time.Now().UTC(),
		})
	}

	WriteSuccessResponse(w, result)
}

// Helper methods

// getUserID extracts user ID from request context
func (h *QueueHandlers) getUserID(r *http.Request) string {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		return "anonymous"
	}
	return fmt.Sprintf("%d", user.ID)
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

// parseSearchCriteria extracts search criteria from query parameters
func (h *QueueHandlers) parseSearchCriteria(r *http.Request) *queue.SearchCriteria {
	criteria := &queue.SearchCriteria{}

	// String parameters
	criteria.Sender = GetQueryParam(r, "sender", "")
	criteria.Recipient = GetQueryParam(r, "recipient", "")
	criteria.MessageID = GetQueryParam(r, "message_id", "")
	criteria.Status = GetQueryParam(r, "status", "")
	criteria.Subject = GetQueryParam(r, "subject", "")
	criteria.MinAge = GetQueryParam(r, "age_min", "")
	criteria.MaxAge = GetQueryParam(r, "age_max", "")

	// Integer parameters with error handling
	if minSizeStr := GetQueryParam(r, "size_min", ""); minSizeStr != "" {
		if minSize, err := strconv.ParseInt(minSizeStr, 10, 64); err == nil {
			criteria.MinSize = minSize
		}
	}

	if maxSizeStr := GetQueryParam(r, "size_max", ""); maxSizeStr != "" {
		if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			criteria.MaxSize = maxSize
		}
	}

	if minRetriesStr := GetQueryParam(r, "retry_count_min", ""); minRetriesStr != "" {
		if minRetries, err := strconv.Atoi(minRetriesStr); err == nil {
			criteria.MinRetries = minRetries
		}
	}

	if maxRetriesStr := GetQueryParam(r, "retry_count_max", ""); maxRetriesStr != "" {
		if maxRetries, err := strconv.Atoi(maxRetriesStr); err == nil {
			criteria.MaxRetries = maxRetries
		}
	}

	return criteria
}

// hasSearchCriteria checks if any search criteria are provided
func (h *QueueHandlers) hasSearchCriteria(criteria *queue.SearchCriteria) bool {
	return criteria.Sender != "" ||
		criteria.Recipient != "" ||
		criteria.MessageID != "" ||
		criteria.Status != "" ||
		criteria.Subject != "" ||
		criteria.MinAge != "" ||
		criteria.MaxAge != "" ||
		criteria.MinSize > 0 ||
		criteria.MaxSize > 0 ||
		criteria.MinRetries > 0 ||
		criteria.MaxRetries > 0
}
