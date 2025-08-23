package logprocessor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

// SearchService provides advanced search capabilities for log entries
type SearchService struct {
	repository *database.Repository
}

// NewSearchService creates a new search service
func NewSearchService(repository *database.Repository) *SearchService {
	return &SearchService{
		repository: repository,
	}
}

// SearchCriteria defines search parameters
type SearchCriteria struct {
	// Time range
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// Message filtering
	MessageID  string   `json:"message_id,omitempty"`
	Sender     string   `json:"sender,omitempty"`
	Recipients []string `json:"recipients,omitempty"`

	// Log filtering
	LogTypes []string `json:"log_types,omitempty"`
	Events   []string `json:"events,omitempty"`
	Status   string   `json:"status,omitempty"`

	// Content filtering
	Keywords  []string `json:"keywords,omitempty"`
	ErrorCode string   `json:"error_code,omitempty"`
	Host      string   `json:"host,omitempty"`

	// Size filtering
	MinSize *int64 `json:"min_size,omitempty"`
	MaxSize *int64 `json:"max_size,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Sorting
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // "asc" or "desc"
}

// SearchResult contains search results and metadata
type SearchResult struct {
	Entries      []database.LogEntry `json:"entries"`
	TotalCount   int                 `json:"total_count"`
	HasMore      bool                `json:"has_more"`
	SearchTime   time.Duration       `json:"search_time"`
	Aggregations *SearchAggregations `json:"aggregations,omitempty"`
}

// SearchAggregations provides summary statistics for search results
type SearchAggregations struct {
	EventCounts        map[string]int `json:"event_counts"`
	LogTypeCounts      map[string]int `json:"log_type_counts"`
	StatusCounts       map[string]int `json:"status_counts"`
	HourlyDistribution map[string]int `json:"hourly_distribution"`
	TopSenders         []SenderCount  `json:"top_senders"`
	TopHosts           []HostCount    `json:"top_hosts"`
}

// SenderCount represents sender statistics
type SenderCount struct {
	Sender string `json:"sender"`
	Count  int    `json:"count"`
}

// HostCount represents host statistics
type HostCount struct {
	Host  string `json:"host"`
	Count int    `json:"count"`
}

// Search performs advanced log entry search
func (s *SearchService) Search(ctx context.Context, criteria SearchCriteria) (*SearchResult, error) {
	start := time.Now()

	// Build the query
	query, countQuery, args := s.buildSearchQuery(criteria)

	// Get total count
	var totalCount int
	if err := s.repository.GetDB().QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Execute search query
	rows, err := s.repository.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var entries []database.LogEntry
	for rows.Next() {
		var entry database.LogEntry
		err := rows.Scan(
			&entry.ID, &entry.Timestamp, &entry.MessageID, &entry.LogType, &entry.Event,
			&entry.Host, &entry.Sender, &entry.RecipientsDB, &entry.Size, &entry.Status,
			&entry.ErrorCode, &entry.ErrorText, &entry.RawLine, &entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Unmarshal recipients
		if err := entry.UnmarshalRecipients(); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Calculate aggregations if requested
	var aggregations *SearchAggregations
	if len(entries) > 0 {
		aggregations = s.calculateAggregations(entries)
	}

	searchTime := time.Since(start)

	return &SearchResult{
		Entries:      entries,
		TotalCount:   totalCount,
		HasMore:      totalCount > criteria.Offset+len(entries),
		SearchTime:   searchTime,
		Aggregations: aggregations,
	}, nil
}

// buildSearchQuery constructs the SQL query based on search criteria
func (s *SearchService) buildSearchQuery(criteria SearchCriteria) (string, string, []interface{}) {
	baseQuery := `
		SELECT id, timestamp, message_id, log_type, event, host, sender, recipients, 
		       size, status, error_code, error_text, raw_line, created_at
		FROM log_entries`

	countQuery := "SELECT COUNT(*) FROM log_entries"

	var conditions []string
	var args []interface{}

	// Time range filtering
	if criteria.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, *criteria.StartTime)
	}

	if criteria.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, *criteria.EndTime)
	}

	// Message ID filtering
	if criteria.MessageID != "" {
		conditions = append(conditions, "message_id = ?")
		args = append(args, criteria.MessageID)
	}

	// Sender filtering
	if criteria.Sender != "" {
		conditions = append(conditions, "sender LIKE ?")
		args = append(args, "%"+criteria.Sender+"%")
	}

	// Recipients filtering
	if len(criteria.Recipients) > 0 {
		recipientConditions := make([]string, len(criteria.Recipients))
		for i, recipient := range criteria.Recipients {
			recipientConditions[i] = "recipients LIKE ?"
			args = append(args, "%"+recipient+"%")
		}
		conditions = append(conditions, "("+strings.Join(recipientConditions, " OR ")+")")
	}

	// Log type filtering
	if len(criteria.LogTypes) > 0 {
		placeholders := make([]string, len(criteria.LogTypes))
		for i, logType := range criteria.LogTypes {
			placeholders[i] = "?"
			args = append(args, logType)
		}
		conditions = append(conditions, "log_type IN ("+strings.Join(placeholders, ",")+")")
	}

	// Event filtering
	if len(criteria.Events) > 0 {
		placeholders := make([]string, len(criteria.Events))
		for i, event := range criteria.Events {
			placeholders[i] = "?"
			args = append(args, event)
		}
		conditions = append(conditions, "event IN ("+strings.Join(placeholders, ",")+")")
	}

	// Status filtering
	if criteria.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, criteria.Status)
	}

	// Keyword filtering
	if len(criteria.Keywords) > 0 {
		keywordConditions := make([]string, len(criteria.Keywords))
		for i, keyword := range criteria.Keywords {
			keywordConditions[i] = "(raw_line LIKE ? OR error_text LIKE ?)"
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
		}
		conditions = append(conditions, "("+strings.Join(keywordConditions, " AND ")+")")
	}

	// Error code filtering
	if criteria.ErrorCode != "" {
		conditions = append(conditions, "error_code LIKE ?")
		args = append(args, "%"+criteria.ErrorCode+"%")
	}

	// Host filtering
	if criteria.Host != "" {
		conditions = append(conditions, "host LIKE ?")
		args = append(args, "%"+criteria.Host+"%")
	}

	// Size filtering
	if criteria.MinSize != nil {
		conditions = append(conditions, "size >= ?")
		args = append(args, *criteria.MinSize)
	}

	if criteria.MaxSize != nil {
		conditions = append(conditions, "size <= ?")
		args = append(args, *criteria.MaxSize)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add WHERE clause to both queries
	baseQuery += whereClause
	countQuery += whereClause

	// Add sorting
	sortBy := "timestamp"
	if criteria.SortBy != "" {
		sortBy = criteria.SortBy
	}

	sortOrder := "DESC"
	if criteria.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	limit := 100
	if criteria.Limit > 0 {
		limit = criteria.Limit
	}

	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, criteria.Offset)

	return baseQuery, countQuery, args
}

// calculateAggregations computes summary statistics for search results
func (s *SearchService) calculateAggregations(entries []database.LogEntry) *SearchAggregations {
	agg := &SearchAggregations{
		EventCounts:        make(map[string]int),
		LogTypeCounts:      make(map[string]int),
		StatusCounts:       make(map[string]int),
		HourlyDistribution: make(map[string]int),
	}

	senderCounts := make(map[string]int)
	hostCounts := make(map[string]int)

	for _, entry := range entries {
		// Count events
		agg.EventCounts[entry.Event]++

		// Count log types
		agg.LogTypeCounts[entry.LogType]++

		// Count statuses
		if entry.Status != nil {
			agg.StatusCounts[*entry.Status]++
		}

		// Hourly distribution
		hour := entry.Timestamp.Format("2006-01-02 15:00")
		agg.HourlyDistribution[hour]++

		// Count senders
		if entry.Sender != nil {
			senderCounts[*entry.Sender]++
		}

		// Count hosts
		if entry.Host != nil {
			hostCounts[*entry.Host]++
		}
	}

	// Get top senders (limit to top 10)
	agg.TopSenders = getTopCounts(senderCounts, 10)

	// Get top hosts (limit to top 10)
	agg.TopHosts = getTopHostCounts(hostCounts, 10)

	return agg
}

// getTopCounts returns the top N senders by count
func getTopCounts(counts map[string]int, limit int) []SenderCount {
	type senderCount struct {
		sender string
		count  int
	}

	var senders []senderCount
	for sender, count := range counts {
		senders = append(senders, senderCount{sender, count})
	}

	// Simple bubble sort (for small datasets)
	for i := 0; i < len(senders)-1; i++ {
		for j := i + 1; j < len(senders); j++ {
			if senders[i].count < senders[j].count {
				senders[i], senders[j] = senders[j], senders[i]
			}
		}
	}

	// Limit results
	if len(senders) > limit {
		senders = senders[:limit]
	}

	result := make([]SenderCount, len(senders))
	for i, s := range senders {
		result[i] = SenderCount{Sender: s.sender, Count: s.count}
	}

	return result
}

// getTopHostCounts returns the top N hosts by count
func getTopHostCounts(counts map[string]int, limit int) []HostCount {
	type hostCount struct {
		host  string
		count int
	}

	var hosts []hostCount
	for host, count := range counts {
		hosts = append(hosts, hostCount{host, count})
	}

	// Simple bubble sort (for small datasets)
	for i := 0; i < len(hosts)-1; i++ {
		for j := i + 1; j < len(hosts); j++ {
			if hosts[i].count < hosts[j].count {
				hosts[i], hosts[j] = hosts[j], hosts[i]
			}
		}
	}

	// Limit results
	if len(hosts) > limit {
		hosts = hosts[:limit]
	}

	result := make([]HostCount, len(hosts))
	for i, h := range hosts {
		result[i] = HostCount{Host: h.host, Count: h.count}
	}

	return result
}

// SearchMessageHistory searches for all log entries related to a message
func (s *SearchService) SearchMessageHistory(ctx context.Context, messageID string) (*MessageCorrelation, error) {
	aggregator := NewLogAggregator(s.repository)
	return aggregator.AggregateMessageData(ctx, messageID)
}

// SearchSimilarMessages finds messages with similar characteristics
func (s *SearchService) SearchSimilarMessages(ctx context.Context, messageID string, limit int) ([]database.LogEntry, error) {
	// Get the reference message's log entries
	logRepo := database.NewLogEntryRepository(s.repository.GetDB())
	refEntries, err := logRepo.GetByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference message entries: %w", err)
	}

	if len(refEntries) == 0 {
		return nil, fmt.Errorf("no log entries found for message %s", messageID)
	}

	// Use the first entry as reference
	refEntry := refEntries[0]

	criteria := SearchCriteria{
		Limit:     limit,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	// Add similarity criteria based on the reference entry
	if refEntry.Sender != nil {
		criteria.Sender = *refEntry.Sender
	}

	if refEntry.Host != nil {
		criteria.Host = *refEntry.Host
	}

	if refEntry.Status != nil {
		criteria.Status = *refEntry.Status
	}

	criteria.Events = []string{refEntry.Event}
	criteria.LogTypes = []string{refEntry.LogType}

	// Search for similar messages
	result, err := s.Search(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar messages: %w", err)
	}

	// Filter out the original message
	var similarEntries []database.LogEntry
	for _, entry := range result.Entries {
		if entry.MessageID == nil || *entry.MessageID != messageID {
			similarEntries = append(similarEntries, entry)
		}
	}

	return similarEntries, nil
}

// GetLogStatistics returns overall log statistics
func (s *SearchService) GetLogStatistics(ctx context.Context, startTime, endTime time.Time) (*LogStatistics, error) {
	stats := &LogStatistics{
		Period: Period{
			Start: startTime,
			End:   endTime,
		},
	}

	// Get total log entries
	query := "SELECT COUNT(*) FROM log_entries WHERE timestamp >= ? AND timestamp <= ?"
	if err := s.repository.GetDB().QueryRowContext(ctx, query, startTime, endTime).Scan(&stats.TotalEntries); err != nil {
		return nil, fmt.Errorf("failed to get total entries: %w", err)
	}

	// Get entries by log type
	query = `
		SELECT log_type, COUNT(*) 
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ? 
		GROUP BY log_type`

	rows, err := s.repository.GetDB().QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get log type counts: %w", err)
	}
	defer rows.Close()

	stats.ByLogType = make(map[string]int)
	for rows.Next() {
		var logType string
		var count int
		if err := rows.Scan(&logType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan log type count: %w", err)
		}
		stats.ByLogType[logType] = count
	}

	// Get entries by event
	query = `
		SELECT event, COUNT(*) 
		FROM log_entries 
		WHERE timestamp >= ? AND timestamp <= ? 
		GROUP BY event`

	rows, err = s.repository.GetDB().QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get event counts: %w", err)
	}
	defer rows.Close()

	stats.ByEvent = make(map[string]int)
	for rows.Next() {
		var event string
		var count int
		if err := rows.Scan(&event, &count); err != nil {
			return nil, fmt.Errorf("failed to scan event count: %w", err)
		}
		stats.ByEvent[event] = count
	}

	return stats, nil
}

// LogStatistics represents log statistics for a time period
type LogStatistics struct {
	Period       Period         `json:"period"`
	TotalEntries int            `json:"total_entries"`
	ByLogType    map[string]int `json:"by_log_type"`
	ByEvent      map[string]int `json:"by_event"`
}

// Period represents a time period
type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
