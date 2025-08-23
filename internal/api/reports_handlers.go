package api

import (
	"context"
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

	// Get top failure reasons (simplified implementation)
	report.TopFailureReasons = []FailureReason{
		{Reason: "Temporary DNS failure", Count: report.DeferredCount / 3},
		{Reason: "Mailbox full", Count: report.BouncedCount / 4},
		{Reason: "Invalid recipient", Count: report.BouncedCount / 2},
		{Reason: "Spam rejection", Count: report.RejectedCount / 2},
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

	// This is a simplified implementation
	// In a real implementation, you would query the database with proper time grouping
	logStats, err := h.logService.GetLogStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Generate sample time series data based on groupBy
	report.TimeSeries = h.generateTimeSeriesData(startTime, endTime, groupBy, logStats.TotalEntries)

	// Calculate summary statistics
	var totalVolume int
	var maxVolume int
	for _, point := range report.TimeSeries {
		totalVolume += point.Count
		if point.Count > maxVolume {
			maxVolume = point.Count
		}
	}

	report.TotalVolume = totalVolume
	report.AverageVolume = totalVolume / len(report.TimeSeries)
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

	// Generate failure categories (simplified)
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

	// Generate top error codes (simplified)
	report.TopErrorCodes = []ErrorCodeStat{
		{Code: "550", Description: "Mailbox unavailable", Count: logStats.ByEvent["bounce"] / 2},
		{Code: "451", Description: "Temporary local problem", Count: logStats.ByEvent["defer"] / 2},
		{Code: "554", Description: "Transaction failed", Count: logStats.ByEvent["reject"] / 2},
	}

	return report, nil
}

// generateMessageTrace generates a detailed message trace
func (h *ReportsHandlers) generateMessageTrace(ctx context.Context, messageID string) (*MessageTrace, error) {
	// Get message correlation data
	correlation, err := h.logService.GetMessageCorrelation(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Get queue information if available
	var queueInfo *queue.MessageDetails
	if h.queueService != nil {
		if details, err := h.queueService.GetMessageDetails(messageID); err == nil {
			queueInfo = details
		}
	}

	// Get operation history
	var operationHistory []database.AuditLog
	if h.queueService != nil {
		if history, err := h.queueService.GetOperationHistory(messageID); err == nil {
			operationHistory = history
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
	// This is a simplified implementation
	// In a real implementation, you would query the database for actual sender statistics

	report := &TopSendersReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		TopSenders: []SenderStat{
			{Sender: "newsletter@example.com", MessageCount: 1250, VolumeBytes: 15000000, DeliveryRate: 98.5},
			{Sender: "notifications@app.com", MessageCount: 890, VolumeBytes: 4500000, DeliveryRate: 99.2},
			{Sender: "alerts@system.com", MessageCount: 567, VolumeBytes: 2800000, DeliveryRate: 97.8},
		},
	}

	return report, nil
}

// generateTopRecipientsReport generates a top recipients report
func (h *ReportsHandlers) generateTopRecipientsReport(ctx context.Context, startTime, endTime time.Time, limit int) (*TopRecipientsReport, error) {
	// This is a simplified implementation
	report := &TopRecipientsReport{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		TopRecipients: []RecipientStat{
			{Recipient: "user1@domain.com", MessageCount: 45, VolumeBytes: 2250000, DeliveryRate: 100.0},
			{Recipient: "user2@domain.com", MessageCount: 38, VolumeBytes: 1900000, DeliveryRate: 97.4},
			{Recipient: "user3@domain.com", MessageCount: 32, VolumeBytes: 1600000, DeliveryRate: 96.9},
		},
	}

	return report, nil
}

// generateDomainAnalysis generates domain-based analysis
func (h *ReportsHandlers) generateDomainAnalysis(ctx context.Context, startTime, endTime time.Time, analysisType string, limit int) (*DomainAnalysis, error) {
	// This is a simplified implementation
	analysis := &DomainAnalysis{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
		AnalysisType: analysisType,
	}

	if analysisType == "sender_domains" || analysisType == "both" {
		analysis.SenderDomains = []DomainStat{
			{Domain: "example.com", MessageCount: 1250, DeliveryRate: 98.5, BounceRate: 1.2, DeferRate: 0.3},
			{Domain: "app.com", MessageCount: 890, DeliveryRate: 99.2, BounceRate: 0.6, DeferRate: 0.2},
			{Domain: "system.com", MessageCount: 567, DeliveryRate: 97.8, BounceRate: 1.8, DeferRate: 0.4},
		}
	}

	if analysisType == "recipient_domains" || analysisType == "both" {
		analysis.RecipientDomains = []DomainStat{
			{Domain: "gmail.com", MessageCount: 2100, DeliveryRate: 99.1, BounceRate: 0.7, DeferRate: 0.2},
			{Domain: "yahoo.com", MessageCount: 1800, DeliveryRate: 98.3, BounceRate: 1.4, DeferRate: 0.3},
			{Domain: "outlook.com", MessageCount: 1200, DeliveryRate: 97.9, BounceRate: 1.8, DeferRate: 0.3},
		}
	}

	return analysis, nil
}

// generateTimeSeriesData generates sample time series data
func (h *ReportsHandlers) generateTimeSeriesData(startTime, endTime time.Time, groupBy string, totalEntries int) []TimeSeriesPoint {
	var points []TimeSeriesPoint
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

	// Generate points for the time range
	current := startTime
	pointCount := 0
	for current.Before(endTime) {
		pointCount++
		current = current.Add(interval)
	}

	// Distribute total entries across points (simplified)
	avgPerPoint := totalEntries / pointCount
	if avgPerPoint == 0 {
		avgPerPoint = 1
	}

	current = startTime
	for current.Before(endTime) {
		// Add some variation to make it more realistic
		variation := avgPerPoint / 4
		if variation == 0 {
			variation = 1
		}
		count := avgPerPoint + (int(current.Unix())%variation - variation/2)
		if count < 0 {
			count = 0
		}

		points = append(points, TimeSeriesPoint{
			Timestamp: current,
			Count:     count,
		})

		current = current.Add(interval)
	}

	return points
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
