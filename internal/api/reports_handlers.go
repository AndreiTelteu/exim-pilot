package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/queue"
)

// ReportsHandlers contains handlers for reporting and analytics endpoints
type ReportsHandlers struct {
	logService   *logprocessor.Service
	queueService *queue.Service
	repository   *database.Repository
}

// NewReportsHandlers creates a new reports handlers instance
func NewReportsHandlers(logService *logprocessor.Service, queueService *queue.Service, repository *database.Repository) *ReportsHandlers {
	return &ReportsHandlers{
		logService:   logService,
		queueService: queueService,
		repository:   repository,
	}
}

// handleDeliverabilityReport handles GET /api/v1/reports/deliverability - Deliverability metrics
func (h *ReportsHandlers) handleDeliverabilityReport(w http.ResponseWriter, r *http.Request) {
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

	// Get deliverability metrics
	deliverabilityReport, err := h.generateDeliverabilityReport(r.Context(), startTime, endTime)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate deliverability report")
		return
	}

	WriteSuccessResponse(w, deliverabilityReport)
}

// handleVolumeReport handles GET /api/v1/reports/volume - Volume analysis
func (h *ReportsHandlers) handleVolumeReport(w http.ResponseWriter, r *http.Request) {
	// Get time range (default to last 30 days)
	endTime := time.Now()
	startTime := endTime.Add(-30 * 24 * time.Hour)

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

	// Get grouping parameter (hour, day, week, month)
	groupBy := GetQueryParam(r, "group_by", "day")
	if groupBy != "hour" && groupBy != "day" && groupBy != "week" && groupBy != "month" {
		groupBy = "day"
	}

	// Generate volume report
	volumeReport, err := h.generateVolumeReport(r.Context(), startTime, endTime, groupBy)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate volume report")
		return
	}

	WriteSuccessResponse(w, volumeReport)
}

// handleFailureReport handles GET /api/v1/reports/failures - Failure breakdown
func (h *ReportsHandlers) handleFailureReport(w http.ResponseWriter, r *http.Request) {
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

	// Get limit for top failures
	limit, err := GetQueryParamInt(r, "limit", 20)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid limit parameter")
		return
	}

	// Generate failure report
	failureReport, err := h.generateFailureReport(r.Context(), startTime, endTime, limit)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate failure report")
		return
	}

	WriteSuccessResponse(w, failureReport)
}

// handleMessageTrace handles GET /api/v1/messages/{id}/trace - Message tracing
func (h *ReportsHandlers) handleMessageTrace(w http.ResponseWriter, r *http.Request) {
	messageID := GetPathParam(r, "id")
	if messageID == "" {
		WriteBadRequestResponse(w, "Message ID is required")
		return
	}

	// Get message trace
	trace, err := h.generateMessageTrace(r.Context(), messageID)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate message trace")
		return
	}

	WriteSuccessResponse(w, trace)
}

// Additional reporting endpoints

// handleTopSenders handles GET /api/v1/reports/top-senders - Top senders analysis
func (h *ReportsHandlers) handleTopSenders(w http.ResponseWriter, r *http.Request) {
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

	// Get limit
	limit, err := GetQueryParamInt(r, "limit", 50)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid limit parameter")
		return
	}

	// Generate top senders report
	topSenders, err := h.generateTopSendersReport(r.Context(), startTime, endTime, limit)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate top senders report")
		return
	}

	WriteSuccessResponse(w, topSenders)
}

// handleTopRecipients handles GET /api/v1/reports/top-recipients - Top recipients analysis
func (h *ReportsHandlers) handleTopRecipients(w http.ResponseWriter, r *http.Request) {
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

	// Get limit
	limit, err := GetQueryParamInt(r, "limit", 50)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid limit parameter")
		return
	}

	// Generate top recipients report
	topRecipients, err := h.generateTopRecipientsReport(r.Context(), startTime, endTime, limit)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate top recipients report")
		return
	}

	WriteSuccessResponse(w, topRecipients)
}

// handleWeeklyOverview handles GET /api/v1/reports/weekly-overview - Weekly overview data
func (h *ReportsHandlers) handleWeeklyOverview(w http.ResponseWriter, r *http.Request) {
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

	// Generate weekly overview report
	weeklyOverview, err := h.generateWeeklyOverviewReport(r.Context(), startTime, endTime)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate weekly overview report")
		return
	}

	WriteSuccessResponse(w, weeklyOverview)
}

// handleDomainAnalysis handles GET /api/v1/reports/domains - Domain-based analysis
func (h *ReportsHandlers) handleDomainAnalysis(w http.ResponseWriter, r *http.Request) {
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

	// Get analysis type (sender_domains, recipient_domains, or both)
	analysisType := GetQueryParam(r, "type", "both")
	if analysisType != "sender_domains" && analysisType != "recipient_domains" && analysisType != "both" {
		analysisType = "both"
	}

	// Get limit
	limit, err := GetQueryParamInt(r, "limit", 50)
	if err != nil {
		WriteBadRequestResponse(w, "Invalid limit parameter")
		return
	}

	// Generate domain analysis
	domainAnalysis, err := h.generateDomainAnalysis(r.Context(), startTime, endTime, analysisType, limit)
	if err != nil {
		WriteInternalErrorResponse(w, "Failed to generate domain analysis")
		return
	}

	WriteSuccessResponse(w, domainAnalysis)
}

// Report generation methods

// generateDeliverabilityReport generates a comprehensive deliverability report
func (h *ReportsHandlers) generateDeliverabilityReport(ctx context.Context, startTime, endTime time.Time) (*DeliverabilityReport, error) {
	// Get log statistics for the period
	logStats, err := h.logService.GetLogStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate deliverability metrics
	report := &DeliverabilityReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		TotalMessages: logStats.TotalEntries,
		EventCounts:   logStats.ByEvent,
		LogTypeCounts: logStats.ByLogType,
	}

	// Calculate rates
	if report.TotalMessages > 0 {
		delivered := logStats.ByEvent["delivery"]
		deferred := logStats.ByEvent["defer"]
		bounced := logStats.ByEvent["bounce"]
		rejected := logStats.ByEvent["reject"]

		report.DeliveredCount = delivered
		report.DeferredCount = deferred
		report.BouncedCount = bounced
		report.RejectedCount = rejected

		report.DeliveryRate = float64(delivered) / float64(report.TotalMessages) * 100
		report.DeferralRate = float64(deferred) / float64(report.TotalMessages) * 100
		report.BounceRate = float64(bounced) / float64(report.TotalMessages) * 100
		report.RejectionRate = float64(rejected) / float64(report.TotalMessages) * 100
	}

	// Get top failure reasons from actual database
	topFailureReasons, err := h.getTopFailureReasons(ctx, startTime, endTime, 10)
	if err != nil {
		// If we can't get real failure reasons, provide empty list instead of fake data
		report.TopFailureReasons = make([]FailureReason, 0)
	} else {
		report.TopFailureReasons = topFailureReasons
	}

	return report, nil
}

// generateVolumeReport generates a volume analysis report
func (h *ReportsHandlers) generateVolumeReport(ctx context.Context, startTime, endTime time.Time, groupBy string) (*VolumeReport, error) {
	report := &VolumeReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		GroupBy: groupBy,
	}

	// Generate real time series data from database
	timeSeries, err := h.generateTimeSeriesData(ctx, startTime, endTime, groupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate time series data: %w", err)
	}
	report.TimeSeries = timeSeries

	// Calculate summary statistics from real data
	var totalVolume int
	var maxVolume int
	for _, point := range report.TimeSeries {
		totalVolume += point.Count
		if point.Count > maxVolume {
			maxVolume = point.Count
		}
	}

	report.TotalVolume = totalVolume
	if len(report.TimeSeries) > 0 {
		report.AverageVolume = totalVolume / len(report.TimeSeries)
	} else {
		report.AverageVolume = 0
	}
	report.PeakVolume = maxVolume

	return report, nil
}

// generateFailureReport generates a failure analysis report
func (h *ReportsHandlers) generateFailureReport(ctx context.Context, startTime, endTime time.Time, limit int) (*FailureReport, error) {
	// Get log statistics
	logStats, err := h.logService.GetLogStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	report := &FailureReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		TotalFailures: logStats.ByEvent["bounce"] + logStats.ByEvent["defer"] + logStats.ByEvent["reject"],
	}

	// Generate failure categories from real data
	if report.TotalFailures > 0 {
		report.FailureCategories = []FailureCategory{
			{
				Category:    "Temporary Failures",
				Count:       logStats.ByEvent["defer"],
				Percentage:  float64(logStats.ByEvent["defer"]) / float64(report.TotalFailures) * 100,
				Description: "Messages deferred due to temporary issues",
			},
			{
				Category:    "Permanent Failures",
				Count:       logStats.ByEvent["bounce"],
				Percentage:  float64(logStats.ByEvent["bounce"]) / float64(report.TotalFailures) * 100,
				Description: "Messages permanently bounced",
			},
			{
				Category:    "Rejections",
				Count:       logStats.ByEvent["reject"],
				Percentage:  float64(logStats.ByEvent["reject"]) / float64(report.TotalFailures) * 100,
				Description: "Messages rejected at SMTP level",
			},
		}
	} else {
		report.FailureCategories = make([]FailureCategory, 0)
	}

	// Get real top error codes from database
	topErrorCodes, err := h.getTopErrorCodes(ctx, startTime, endTime, limit)
	if err != nil {
		// If we can't get real error codes, provide empty list instead of fake data
		report.TopErrorCodes = make([]ErrorCodeStat, 0)
	} else {
		report.TopErrorCodes = topErrorCodes
	}

	return report, nil
}

// generateMessageTrace generates a detailed message trace
func (h *ReportsHandlers) generateMessageTrace(ctx context.Context, messageID string) (*MessageTrace, error) {
	// Create a message trace repository to get comprehensive delivery trace
	traceRepo := database.NewMessageTraceRepository(h.repository.GetDB())

	// Get comprehensive message delivery trace
	_, err := traceRepo.GetMessageDeliveryTrace(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message delivery trace: %w", err)
	}

	// Get message correlation data from log service
	correlation, err := h.logService.GetMessageCorrelation(ctx, messageID)
	if err != nil {
		// If correlation fails, create a basic one from the delivery trace
		correlation = &logprocessor.MessageCorrelation{
			MessageID: messageID,
			Timeline:  make([]logprocessor.TimelineEvent, 0),
		}
	}

	// Get queue information if available
	var queueInfo *queue.MessageDetails
	if h.queueService != nil {
		if details, err := h.queueService.GetMessageDetails(messageID); err == nil {
			queueInfo = details
		}
	}

	// Get operation history from audit logs
	auditLogs, err := h.repository.GetAuditLogs(map[string]interface{}{"message_id": messageID})
	if err != nil {
		auditLogs = make([]*database.AuditLog, 0) // Empty list if query fails
	}

	// Convert to the expected type
	operationHistory := make([]database.AuditLog, 0)
	for _, log := range auditLogs {
		if log != nil {
			operationHistory = append(operationHistory, *log)
		}
	}

	trace := &MessageTrace{
		MessageID:        messageID,
		Correlation:      correlation,
		QueueInfo:        queueInfo,
		OperationHistory: operationHistory,
		GeneratedAt:      time.Now(),
	}

	return trace, nil
}

// generateTopSendersReport generates a top senders report
func (h *ReportsHandlers) generateTopSendersReport(ctx context.Context, startTime, endTime time.Time, limit int) (*TopSendersReport, error) {
	report := &TopSendersReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		TopSenders: make([]SenderStat, 0),
	}

	// Get top senders from database
	topSenders, err := h.getTopSenders(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top senders: %w", err)
	}

	report.TopSenders = topSenders
	return report, nil
}

// generateTopRecipientsReport generates a top recipients report
func (h *ReportsHandlers) generateTopRecipientsReport(ctx context.Context, startTime, endTime time.Time, limit int) (*TopRecipientsReport, error) {
	report := &TopRecipientsReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
	}

	// Get top recipients from database
	topRecipients, err := h.getTopRecipients(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top recipients: %w", err)
	}

	report.TopRecipients = topRecipients
	return report, nil
}

// generateDomainAnalysis generates domain-based analysis
func (h *ReportsHandlers) generateDomainAnalysis(ctx context.Context, startTime, endTime time.Time, analysisType string, limit int) (*DomainAnalysis, error) {
	analysis := &DomainAnalysis{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		AnalysisType: analysisType,
	}

	var err error

	if analysisType == "sender_domains" || analysisType == "both" {
		analysis.SenderDomains, err = h.getSenderDomainStats(ctx, startTime, endTime, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get sender domain stats: %w", err)
		}
	}

	if analysisType == "recipient_domains" || analysisType == "both" {
		analysis.RecipientDomains, err = h.getRecipientDomainStats(ctx, startTime, endTime, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get recipient domain stats: %w", err)
		}
	}

	return analysis, nil
}

// generateTimeSeriesData generates time series data from database
func (h *ReportsHandlers) generateTimeSeriesData(ctx context.Context, startTime, endTime time.Time, groupBy string) ([]TimeSeriesPoint, error) {
	points := make([]TimeSeriesPoint, 0)
	var interval time.Duration

	switch groupBy {
	case "hour":
		interval = time.Hour
	case "day":
		interval = 24 * time.Hour
	case "week":
		interval = 7 * 24 * time.Hour
	case "month":
		interval = 30 * 24 * time.Hour
	default:
		interval = 24 * time.Hour
	}

	// Query database for actual time series data
	var query string
	if groupBy == "hour" {
		query = `
			SELECT 
				strftime('%Y-%m-%d %H:00:00', timestamp) as period,
				COUNT(*) as count
			FROM log_entries 
			WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY strftime('%Y-%m-%d %H', timestamp)
			ORDER BY period
		`
	} else if groupBy == "day" {
		query = `
			SELECT 
				strftime('%Y-%m-%d', timestamp) as period,
				COUNT(*) as count
			FROM log_entries 
			WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY strftime('%Y-%m-%d', timestamp)
			ORDER BY period
		`
	} else if groupBy == "week" {
		query = `
			SELECT 
				strftime('%Y-%W', timestamp) as period,
				COUNT(*) as count
			FROM log_entries 
			WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY strftime('%Y-%W', timestamp)
			ORDER BY period
		`
	} else { // month
		query = `
			SELECT 
				strftime('%Y-%m', timestamp) as period,
				COUNT(*) as count
			FROM log_entries 
			WHERE timestamp >= ? AND timestamp <= ?
			GROUP BY strftime('%Y-%m', timestamp)
			ORDER BY period
		`
	}

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query time series data: %w", err)
	}
	defer rows.Close()

	resultMap := make(map[string]int)
	for rows.Next() {
		var period string
		var count int
		if err := rows.Scan(&period, &count); err != nil {
			return nil, fmt.Errorf("failed to scan time series row: %w", err)
		}
		resultMap[period] = count
	}

	// Generate complete time range with zero values for missing periods
	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		var periodKey string
		if groupBy == "hour" {
			periodKey = current.Format("2006-01-02 15:00:00")
		} else if groupBy == "day" {
			periodKey = current.Format("2006-01-02")
		} else if groupBy == "week" {
			// Format as year-week
			year, week := current.ISOWeek()
			periodKey = fmt.Sprintf("%d-%02d", year, week)
		} else {
			periodKey = current.Format("2006-01")
		}

		count := resultMap[periodKey]
		points = append(points, TimeSeriesPoint{
			Timestamp: current,
			Count:     count,
		})

		current = current.Add(interval)
	}

	return points, nil
}

// getTopFailureReasons retrieves the most common failure reasons from log entries
func (h *ReportsHandlers) getTopFailureReasons(ctx context.Context, startTime, endTime time.Time, limit int) ([]FailureReason, error) {
	query := `
		SELECT 
			COALESCE(error_text, 'Unknown error') as reason,
			COUNT(*) as count
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ?
		AND event IN ('defer', 'bounce', 'reject')
		AND error_text IS NOT NULL
		AND error_text != ''
		GROUP BY error_text
		ORDER BY count DESC
		LIMIT ?
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failure reasons: %w", err)
	}
	defer rows.Close()

	reasons := make([]FailureReason, 0)
	for rows.Next() {
		var reason FailureReason
		if err := rows.Scan(&reason.Reason, &reason.Count); err != nil {
			return nil, fmt.Errorf("failed to scan failure reason: %w", err)
		}
		reasons = append(reasons, reason)
	}

	return reasons, nil
}

// getTopErrorCodes retrieves the most common error codes from log entries
func (h *ReportsHandlers) getTopErrorCodes(ctx context.Context, startTime, endTime time.Time, limit int) ([]ErrorCodeStat, error) {
	query := `
		SELECT 
			COALESCE(error_code, 'Unknown') as code,
			COALESCE(error_text, 'No description') as description,
			COUNT(*) as count
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ?
		AND event IN ('defer', 'bounce', 'reject')
		AND error_code IS NOT NULL
		AND error_code != ''
		GROUP BY error_code, error_text
		ORDER BY count DESC
		LIMIT ?
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query error codes: %w", err)
	}
	defer rows.Close()

	errorCodes := make([]ErrorCodeStat, 0)
	for rows.Next() {
		var errorCode ErrorCodeStat
		if err := rows.Scan(&errorCode.Code, &errorCode.Description, &errorCode.Count); err != nil {
			return nil, fmt.Errorf("failed to scan error code: %w", err)
		}
		errorCodes = append(errorCodes, errorCode)
	}

	return errorCodes, nil
}

// getTopSenders retrieves the top senders by message count from log entries
func (h *ReportsHandlers) getTopSenders(ctx context.Context, startTime, endTime time.Time, limit int) ([]SenderStat, error) {
	query := `
		SELECT 
			COALESCE(sender, 'Unknown') as sender,
			COUNT(*) as message_count,
			COALESCE(SUM(size), 0) as volume_bytes,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'delivery' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as delivery_rate
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ?
		AND sender IS NOT NULL
		AND sender != ''
		AND event IN ('arrival', 'delivery', 'defer', 'bounce')
		GROUP BY sender
		ORDER BY message_count DESC
		LIMIT ?
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top senders: %w", err)
	}
	defer rows.Close()

	senders := make([]SenderStat, 0)
	for rows.Next() {
		var sender SenderStat
		if err := rows.Scan(&sender.Sender, &sender.MessageCount, &sender.VolumeBytes, &sender.DeliveryRate); err != nil {
			return nil, fmt.Errorf("failed to scan sender stat: %w", err)
		}
		senders = append(senders, sender)
	}

	return senders, nil
}

// getTopRecipients retrieves the top recipients by message count from log entries
func (h *ReportsHandlers) getTopRecipients(ctx context.Context, startTime, endTime time.Time, limit int) ([]RecipientStat, error) {
	query := `
		SELECT 
			recipient_email,
			COUNT(*) as message_count,
			COALESCE(SUM(size), 0) as volume_bytes,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'delivery' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as delivery_rate
		FROM (
			SELECT 
				TRIM(json_extract(value, '$')) as recipient_email,
				event,
				size
			FROM log_entries, 
				json_each(recipients)
			WHERE timestamp >= ? AND timestamp <= ?
			AND recipients IS NOT NULL
			AND recipients != ''
			AND recipients != '[]'
			AND event IN ('arrival', 'delivery', 'defer', 'bounce')
		) recipient_data
		WHERE recipient_email IS NOT NULL
		AND recipient_email != ''
		GROUP BY recipient_email
		ORDER BY message_count DESC
		LIMIT ?
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top recipients: %w", err)
	}
	defer rows.Close()

	recipients := make([]RecipientStat, 0)
	for rows.Next() {
		var recipient RecipientStat
		if err := rows.Scan(&recipient.Recipient, &recipient.MessageCount, &recipient.VolumeBytes, &recipient.DeliveryRate); err != nil {
			return nil, fmt.Errorf("failed to scan recipient stat: %w", err)
		}
		recipients = append(recipients, recipient)
	}

	return recipients, nil
}

// getSenderDomainStats retrieves sender domain statistics from log entries
func (h *ReportsHandlers) getSenderDomainStats(ctx context.Context, startTime, endTime time.Time, limit int) ([]DomainStat, error) {
	query := `
		SELECT 
			CASE 
				WHEN sender LIKE '%@%' THEN LOWER(SUBSTR(sender, INSTR(sender, '@') + 1))
				ELSE 'unknown'
			END as domain,
			COUNT(*) as message_count,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'delivery' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as delivery_rate,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'bounce' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as bounce_rate,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'defer' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as defer_rate
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ?
		AND sender IS NOT NULL
		AND sender != ''
		AND event IN ('arrival', 'delivery', 'defer', 'bounce')
		GROUP BY domain
		HAVING domain != 'unknown'
		ORDER BY message_count DESC
		LIMIT ?
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query sender domain stats: %w", err)
	}
	defer rows.Close()

	domains := make([]DomainStat, 0)
	for rows.Next() {
		var domain DomainStat
		if err := rows.Scan(&domain.Domain, &domain.MessageCount, &domain.DeliveryRate, &domain.BounceRate, &domain.DeferRate); err != nil {
			return nil, fmt.Errorf("failed to scan domain stat: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// getRecipientDomainStats retrieves recipient domain statistics from log entries
func (h *ReportsHandlers) getRecipientDomainStats(ctx context.Context, startTime, endTime time.Time, limit int) ([]DomainStat, error) {
	query := `
		SELECT 
			domain,
			COUNT(*) as message_count,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'delivery' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as delivery_rate,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'bounce' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as bounce_rate,
			COALESCE(
				100.0 * SUM(CASE WHEN event = 'defer' THEN 1 ELSE 0 END) / 
				NULLIF(COUNT(*), 0), 
				0
			) as defer_rate
		FROM (
			SELECT 
				CASE 
					WHEN TRIM(json_extract(value, '$')) LIKE '%@%' 
					THEN LOWER(SUBSTR(TRIM(json_extract(value, '$')), 
						INSTR(TRIM(json_extract(value, '$')), '@') + 1))
					ELSE 'unknown'
				END as domain,
				event
			FROM log_entries, 
				json_each(recipients)
			WHERE timestamp >= ? AND timestamp <= ?
			AND recipients IS NOT NULL
			AND recipients != ''
			AND recipients != '[]'
			AND event IN ('arrival', 'delivery', 'defer', 'bounce')
		) recipient_data
		WHERE domain != 'unknown'
		AND domain != ''
		GROUP BY domain
		ORDER BY message_count DESC
		LIMIT ?
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipient domain stats: %w", err)
	}
	defer rows.Close()

	domains := make([]DomainStat, 0)
	for rows.Next() {
		var domain DomainStat
		if err := rows.Scan(&domain.Domain, &domain.MessageCount, &domain.DeliveryRate, &domain.BounceRate, &domain.DeferRate); err != nil {
			return nil, fmt.Errorf("failed to scan domain stat: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// getDailyBreakdown retrieves daily statistics for the specified time range
func (h *ReportsHandlers) getDailyBreakdown(ctx context.Context, startTime, endTime time.Time) ([]DailyStats, error) {
	query := `
		SELECT 
			strftime('%Y-%m-%d', timestamp) as date,
			COUNT(*) as total_messages,
			SUM(CASE WHEN event = 'delivery' THEN 1 ELSE 0 END) as delivered_messages,
			SUM(CASE WHEN event = 'bounce' THEN 1 ELSE 0 END) as bounced_messages,
			SUM(CASE WHEN event = 'defer' THEN 1 ELSE 0 END) as deferred_messages,
			SUM(CASE WHEN event = 'reject' THEN 1 ELSE 0 END) as rejected_messages
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ?
		AND event IN ('arrival', 'delivery', 'defer', 'bounce', 'reject')
		GROUP BY strftime('%Y-%m-%d', timestamp)
		ORDER BY date
	`

	rows, err := h.repository.GetDB().QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily breakdown: %w", err)
	}
	defer rows.Close()

	dailyStatsMap := make(map[string]DailyStats)
	for rows.Next() {
		var stats DailyStats
		if err := rows.Scan(&stats.Date, &stats.TotalMessages, &stats.DeliveredMessages, &stats.BouncedMessages, &stats.DeferredMessages, &stats.RejectedMessages); err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}
		dailyStatsMap[stats.Date] = stats
	}

	// Generate complete daily breakdown including days with no data
	dailyBreakdown := make([]DailyStats, 0)
	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		dateStr := current.Format("2006-01-02")
		if stats, exists := dailyStatsMap[dateStr]; exists {
			dailyBreakdown = append(dailyBreakdown, stats)
		} else {
			// Add empty stats for days with no data
			dailyBreakdown = append(dailyBreakdown, DailyStats{
				Date:              dateStr,
				TotalMessages:     0,
				DeliveredMessages: 0,
				BouncedMessages:   0,
				DeferredMessages:  0,
				RejectedMessages:  0,
			})
		}
		current = current.Add(24 * time.Hour)
	}

	return dailyBreakdown, nil
}

// Data structures for reports

type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type DeliverabilityReport struct {
	Period            Period          `json:"period"`
	TotalMessages     int             `json:"total_messages"`
	DeliveredCount    int             `json:"delivered_count"`
	DeferredCount     int             `json:"deferred_count"`
	BouncedCount      int             `json:"bounced_count"`
	RejectedCount     int             `json:"rejected_count"`
	DeliveryRate      float64         `json:"delivery_rate"`
	DeferralRate      float64         `json:"deferral_rate"`
	BounceRate        float64         `json:"bounce_rate"`
	RejectionRate     float64         `json:"rejection_rate"`
	EventCounts       map[string]int  `json:"event_counts"`
	LogTypeCounts     map[string]int  `json:"log_type_counts"`
	TopFailureReasons []FailureReason `json:"top_failure_reasons"`
}

type FailureReason struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type VolumeReport struct {
	Period        Period            `json:"period"`
	GroupBy       string            `json:"group_by"`
	TotalVolume   int               `json:"total_volume"`
	AverageVolume int               `json:"average_volume"`
	PeakVolume    int               `json:"peak_volume"`
	TimeSeries    []TimeSeriesPoint `json:"time_series"`
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

type FailureReport struct {
	Period            Period            `json:"period"`
	TotalFailures     int               `json:"total_failures"`
	FailureCategories []FailureCategory `json:"failure_categories"`
	TopErrorCodes     []ErrorCodeStat   `json:"top_error_codes"`
}

type FailureCategory struct {
	Category    string  `json:"category"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	Description string  `json:"description"`
}

type ErrorCodeStat struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

type MessageTrace struct {
	MessageID        string                           `json:"message_id"`
	Correlation      *logprocessor.MessageCorrelation `json:"correlation"`
	QueueInfo        *queue.MessageDetails            `json:"queue_info,omitempty"`
	OperationHistory []database.AuditLog              `json:"operation_history"`
	GeneratedAt      time.Time                        `json:"generated_at"`
}

type TopSendersReport struct {
	Period     Period       `json:"period"`
	TopSenders []SenderStat `json:"top_senders"`
}

type SenderStat struct {
	Sender       string  `json:"sender"`
	MessageCount int     `json:"message_count"`
	VolumeBytes  int64   `json:"volume_bytes"`
	DeliveryRate float64 `json:"delivery_rate"`
}

type TopRecipientsReport struct {
	Period        Period          `json:"period"`
	TopRecipients []RecipientStat `json:"top_recipients"`
}

type RecipientStat struct {
	Recipient    string  `json:"recipient"`
	MessageCount int     `json:"message_count"`
	VolumeBytes  int64   `json:"volume_bytes"`
	DeliveryRate float64 `json:"delivery_rate"`
}

type DomainAnalysis struct {
	Period           Period       `json:"period"`
	AnalysisType     string       `json:"analysis_type"`
	SenderDomains    []DomainStat `json:"sender_domains,omitempty"`
	RecipientDomains []DomainStat `json:"recipient_domains,omitempty"`
}

type DomainStat struct {
	Domain       string  `json:"domain"`
	MessageCount int     `json:"message_count"`
	DeliveryRate float64 `json:"delivery_rate"`
	BounceRate   float64 `json:"bounce_rate"`
	DeferRate    float64 `json:"defer_rate"`
}

// WeeklyOverviewData represents weekly overview data for the dashboard
type WeeklyOverviewData struct {
	Period         Period        `json:"period"`
	Summary        WeeklySummary `json:"summary"`
	DailyBreakdown []DailyStats  `json:"daily_breakdown"`
	TopDomains     []DomainStat  `json:"top_domains"`
	QueueStatus    QueueStatus   `json:"queue_status"`
}

type WeeklySummary struct {
	TotalMessages     int     `json:"total_messages"`
	DeliveredMessages int     `json:"delivered_messages"`
	BouncedMessages   int     `json:"bounced_messages"`
	DeferredMessages  int     `json:"deferred_messages"`
	RejectedMessages  int     `json:"rejected_messages"`
	DeliveryRate      float64 `json:"delivery_rate"`
	BounceRate        float64 `json:"bounce_rate"`
}

type DailyStats struct {
	Date              string `json:"date"`
	TotalMessages     int    `json:"total_messages"`
	DeliveredMessages int    `json:"delivered_messages"`
	BouncedMessages   int    `json:"bounced_messages"`
	DeferredMessages  int    `json:"deferred_messages"`
	RejectedMessages  int    `json:"rejected_messages"`
}

type QueueStatus struct {
	TotalInQueue   int       `json:"total_in_queue"`
	FrozenMessages int       `json:"frozen_messages"`
	OldestMessage  time.Time `json:"oldest_message,omitempty"`
}

// generateWeeklyOverviewReport generates weekly overview data for dashboard
func (h *ReportsHandlers) generateWeeklyOverviewReport(ctx context.Context, startTime, endTime time.Time) (*WeeklyOverviewData, error) {
	// Get log statistics for the period
	logStats, err := h.logService.GetLogStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get log statistics: %w", err)
	}

	// Calculate summary statistics
	summary := WeeklySummary{
		TotalMessages:     logStats.TotalEntries,
		DeliveredMessages: logStats.ByEvent["delivery"],
		BouncedMessages:   logStats.ByEvent["bounce"],
		DeferredMessages:  logStats.ByEvent["defer"],
		RejectedMessages:  logStats.ByEvent["reject"],
	}

	if summary.TotalMessages > 0 {
		summary.DeliveryRate = float64(summary.DeliveredMessages) / float64(summary.TotalMessages) * 100
		summary.BounceRate = float64(summary.BouncedMessages) / float64(summary.TotalMessages) * 100
	}

	// Get daily breakdown for the period
	dailyBreakdown, err := h.getDailyBreakdown(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily breakdown: %w", err)
	}

	// Get top domains for the period
	topDomains, err := h.getRecipientDomainStats(ctx, startTime, endTime, 5)
	if err != nil {
		// Don't fail the whole report if top domains fails
		topDomains = []DomainStat{}
	}

	// Get queue status if queueService is available
	queueStatus := QueueStatus{
		TotalInQueue:   0,
		FrozenMessages: 0,
		OldestMessage:  time.Time{},
	}
	if h.queueService != nil {
		if status, err := h.queueService.GetQueueStatus(); err == nil {
			queueStatus.TotalInQueue = status.TotalMessages
			queueStatus.FrozenMessages = status.FrozenMessages
			// Calculate oldest message time from age duration
			if status.OldestMessageAge > 0 {
				queueStatus.OldestMessage = time.Now().Add(-status.OldestMessageAge)
			}
		}
	}

	overview := &WeeklyOverviewData{
		Period:         Period{Start: startTime, End: endTime},
		Summary:        summary,
		DailyBreakdown: dailyBreakdown,
		TopDomains:     topDomains,
		QueueStatus:    queueStatus,
	}

	return overview, nil
}
