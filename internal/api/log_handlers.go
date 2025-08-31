package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/websocket"
)

// LogHandlers contains handlers for log and monitoring endpoints
type LogHandlers struct {
	logService *logprocessor.Service
	wsService  *websocket.Service
}

// NewLogHandlers creates a new log handlers instance
func NewLogHandlers(logService *logprocessor.Service, wsService *websocket.Service) *LogHandlers {
	return &LogHandlers{
		logService: logService,
		wsService:  wsService,
	}
}

// handleLogsList handles GET /api/v1/logs - List log entries with pagination and filtering
func (h *LogHandlers) handleLogsList(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	page, perPage, err := GetPaginationParams(r)
	if err != nil {
		WriteBadRequestResponse(w, err.Error())
		return
	}

	// Build search criteria from query parameters
	criteria := logprocessor.SearchCriteria{
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	}

	// Parse time range
	if startTimeStr := GetQueryParam(r, "start_time", ""); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			criteria.StartTime = &startTime
		}
	}

	if endTimeStr := GetQueryParam(r, "end_time", ""); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			criteria.EndTime = &endTime
		}
	}

	// Parse other filters
	if messageID := GetQueryParam(r, "message_id", ""); messageID != "" {
		criteria.MessageID = messageID
	}

	if sender := GetQueryParam(r, "sender", ""); sender != "" {
		criteria.Sender = sender
	}

	if logType := GetQueryParam(r, "log_type", ""); logType != "" {
		criteria.LogTypes = []string{logType}
	}

	if event := GetQueryParam(r, "event", ""); event != "" {
		criteria.Events = []string{event}
	}

	if status := GetQueryParam(r, "status", ""); status != "" {
		criteria.Status = status
	}

	if host := GetQueryParam(r, "host", ""); host != "" {
		criteria.Host = host
	}

	if errorCode := GetQueryParam(r, "error_code", ""); errorCode != "" {
		criteria.ErrorCode = errorCode
	}

	// Parse size filters
	if minSizeStr := GetQueryParam(r, "min_size", ""); minSizeStr != "" {
		if minSize, err := strconv.ParseInt(minSizeStr, 10, 64); err == nil {
			criteria.MinSize = &minSize
		}
	}

	if maxSizeStr := GetQueryParam(r, "max_size", ""); maxSizeStr != "" {
		if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			criteria.MaxSize = &maxSize
		}
	}

	// Parse sorting
	criteria.SortBy = GetQueryParam(r, "sort_by", "timestamp")
	criteria.SortOrder = GetQueryParam(r, "sort_order", "desc")

	// Perform search
	result, err := h.logService.SearchLogs(r.Context(), criteria)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to search log entries")
		return
	}

	// Create response
	response := map[string]interface{}{
		"entries":      result.Entries,
		"search_time":  result.SearchTime.String(),
		"aggregations": result.Aggregations,
	}

	meta := CalculatePagination(page, perPage, result.TotalCount)
	WriteSuccessResponseWithMeta(w, response, meta)
}

// handleLogsSearch handles POST /api/v1/logs/search - Advanced log search
func (h *LogHandlers) handleLogsSearch(w http.ResponseWriter, r *http.Request) {
	var searchRequest struct {
		Criteria logprocessor.SearchCriteria `json:"criteria"`
		Page     int                         `json:"page"`
		PerPage  int                         `json:"per_page"`
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
		searchRequest.PerPage = 100
	}

	// Apply pagination to criteria
	searchRequest.Criteria.Limit = searchRequest.PerPage
	searchRequest.Criteria.Offset = (searchRequest.Page - 1) * searchRequest.PerPage

	// Perform search
	result, err := h.logService.SearchLogs(r.Context(), searchRequest.Criteria)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to perform advanced search")
		return
	}

	// Create response
	response := map[string]interface{}{
		"entries":      result.Entries,
		"search_time":  result.SearchTime.String(),
		"aggregations": result.Aggregations,
		"criteria":     searchRequest.Criteria,
	}

	meta := CalculatePagination(searchRequest.Page, searchRequest.PerPage, result.TotalCount)
	WriteSuccessResponseWithMeta(w, response, meta)
}

// handleLogsTail handles WebSocket endpoint for real-time log tail
func (h *LogHandlers) handleLogsTail(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response indicating WebSocket support is needed
	// In a full implementation, this would upgrade the connection to WebSocket
	// and stream real-time log entries

	response := map[string]interface{}{
		"message": "WebSocket endpoint for real-time log tail",
		"note":    "This endpoint requires WebSocket implementation",
		"filters": map[string]string{
			"message_id": GetQueryParam(r, "message_id", ""),
			"log_type":   GetQueryParam(r, "log_type", ""),
			"event":      GetQueryParam(r, "event", ""),
			"keywords":   GetQueryParam(r, "keywords", ""),
		},
	}

	WriteSuccessResponse(w, response)
}

// handleDashboard handles GET /api/v1/dashboard - Dashboard metrics
func (h *LogHandlers) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get time range for dashboard metrics (default to last 24 hours)
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	// Parse custom time range if provided
	if startTimeStr := GetQueryParam(r, "start_time", ""); startTimeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsedTime
		}
	}

	if endTimeStr := GetQueryParam(r, "end_time", ""); endTimeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsedTime
		}
	}

	// Get log statistics
	logStats, err := h.logService.GetLogStatistics(r.Context(), startTime, endTime)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve log statistics")
		return
	}

	// Get service status
	serviceStatus := h.logService.GetServiceStatus()

	// Get retention information
	retentionInfo, err := h.logService.GetRetentionInfo(r.Context())
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve retention information")
		return
	}

	// Create dashboard response
	dashboard := map[string]interface{}{
		"period": map[string]interface{}{
			"start": startTime,
			"end":   endTime,
		},
		"log_statistics": logStats,
		"service_status": serviceStatus,
		"retention_info": retentionInfo,
		"summary": map[string]interface{}{
			"total_log_entries": logStats.TotalEntries,
			"log_types":         len(logStats.ByLogType),
			"event_types":       len(logStats.ByEvent),
			"service_running":   serviceStatus.Running,
		},
	}

	WriteSuccessResponse(w, dashboard)
}

// Additional log-related endpoints

// handleMessageHistory handles GET /api/v1/logs/messages/{id}/history - Get message history
func (h *LogHandlers) handleMessageHistory(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get message history
	history, err := h.logService.GetMessageHistory(r.Context(), messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve message history")
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message_id": messageID,
		"history":    history,
	})
}

// handleMessageCorrelation handles GET /api/v1/logs/messages/{id}/correlation - Get message correlation
func (h *LogHandlers) handleMessageCorrelation(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get message correlation
	correlation, err := h.logService.GetMessageCorrelation(r.Context(), messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve message correlation")
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message_id":  messageID,
		"correlation": correlation,
	})
}

// handleSimilarMessages handles GET /api/v1/logs/messages/{id}/similar - Find similar messages
func (h *LogHandlers) handleSimilarMessages(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get limit parameter
	limit, err := GetQueryParamInt(r, "limit", 10)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid limit parameter")
		return
	}

	// Find similar messages
	similarMessages, err := h.logService.FindSimilarMessages(r.Context(), messageID, limit)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to find similar messages")
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message_id":       messageID,
		"similar_messages": similarMessages,
		"count":            len(similarMessages),
	})
}

// handleLogStatistics handles GET /api/v1/logs/statistics - Get detailed log statistics
func (h *LogHandlers) handleLogStatistics(w http.ResponseWriter, r *http.Request) {
	// Get time range (default to last 7 days)
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour)

	// Parse custom time range if provided
	if startTimeStr := GetQueryParam(r, "start_time", ""); startTimeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsedTime
		}
	}

	if endTimeStr := GetQueryParam(r, "end_time", ""); endTimeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsedTime
		}
	}

	// Get statistics
	stats, err := h.logService.GetLogStatistics(r.Context(), startTime, endTime)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to retrieve log statistics")
		return
	}

	WriteSuccessResponse(w, stats)
}

// handleServiceStatus handles GET /api/v1/logs/service/status - Get service status
func (h *LogHandlers) handleServiceStatus(w http.ResponseWriter, r *http.Request) {
	status := h.logService.GetServiceStatus()
	WriteSuccessResponse(w, status)
}

// handleTriggerCorrelation handles POST /api/v1/logs/correlation/trigger - Manually trigger correlation
func (h *LogHandlers) handleTriggerCorrelation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}

	if err := ParseJSONBody(r, &request); err != nil {
		WriteBadRequestResponse(w, "Invalid JSON: "+err.Error())
		return
	}

	// Validate time range
	if request.StartTime.IsZero() || request.EndTime.IsZero() {
		WriteBadRequestResponse(w, "Both start_time and end_time are required")
		return
	}

	if request.StartTime.After(request.EndTime) {
		WriteBadRequestResponse(w, "start_time must be before end_time")
		return
	}

	// Trigger correlation
	if err := h.logService.TriggerCorrelation(r.Context(), request.StartTime, request.EndTime); err != nil {
		WriteInternalErrorResponse(w, "Failed to trigger correlation: "+err.Error())
		return
	}

	WriteSuccessResponse(w, map[string]interface{}{
		"message":    "Correlation triggered successfully",
		"start_time": request.StartTime,
		"end_time":   request.EndTime,
	})
}

// handleExportLogs handles GET /api/v1/logs/export - Export logs in various formats
func (h *LogHandlers) handleExportLogs(w http.ResponseWriter, r *http.Request) {
	// Get export format
	format := GetQueryParam(r, "format", "json")
	if format != "json" && format != "csv" && format != "txt" {
		WriteBadRequestResponse(w, "Invalid format. Supported formats: json, csv, txt")
		return
	}

	// Build search criteria (similar to handleLogsList)
	criteria := logprocessor.SearchCriteria{
		Limit: 10000, // Limit exports to prevent memory issues
	}

	// Parse filters (same as handleLogsList)
	if startTimeStr := GetQueryParam(r, "start_time", ""); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			criteria.StartTime = &startTime
		}
	}

	if endTimeStr := GetQueryParam(r, "end_time", ""); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			criteria.EndTime = &endTime
		}
	}

	if messageID := GetQueryParam(r, "message_id", ""); messageID != "" {
		criteria.MessageID = messageID
	}

	// Perform search
	result, err := h.logService.SearchLogs(r.Context(), criteria)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to export logs")
		return
	}

	// For now, just return JSON format
	// In a full implementation, this would handle CSV and TXT formats
	response := map[string]interface{}{
		"format":      format,
		"entries":     result.Entries,
		"total_count": result.TotalCount,
		"export_time": time.Now(),
		"search_time": result.SearchTime.String(),
		"note":        "CSV and TXT export formats would be implemented here",
	}

	WriteSuccessResponse(w, response)
}
