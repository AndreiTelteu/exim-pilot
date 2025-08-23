# Log Processor Package

The `logprocessor` package provides comprehensive log processing functionality for the Exim Control Panel, including log entry storage, indexing, aggregation, message correlation, background processing, and data retention management.

## Features

### Core Functionality
- **Log Entry Storage**: Efficient storage of parsed log entries with proper indexing
- **Message Correlation**: Automatic correlation of log entries belonging to the same message
- **Advanced Search**: Powerful search capabilities with filtering, sorting, and aggregations
- **Background Processing**: Automated correlation, cleanup, and metrics collection
- **Data Retention**: Configurable retention policies with automatic cleanup

### Key Components

#### 1. Service (`service.go`)
The main service that orchestrates all log processing functionality.

```go
// Create and start the service
config := DefaultServiceConfig()
service := NewService(repository, config)
service.Start()
defer service.Stop()

// Process log entries
ctx := context.Background()
err := service.ProcessLogEntry(ctx, logEntry)

// Search logs
results, err := service.SearchLogs(ctx, searchCriteria)

// Get message correlation
correlation, err := service.GetMessageCorrelation(ctx, messageID)
```

#### 2. Log Aggregator (`aggregator.go`)
Handles message correlation and timeline building.

```go
aggregator := NewLogAggregator(repository)
correlation, err := aggregator.AggregateMessageData(ctx, messageID)
```

#### 3. Background Service (`background_service.go`)
Manages background tasks like correlation, cleanup, and metrics collection.

```go
backgroundService := NewBackgroundService(repository, config)
backgroundService.Start()
```

#### 4. Search Service (`search.go`)
Provides advanced search capabilities with filtering and aggregations.

```go
searchService := NewSearchService(repository)
results, err := searchService.Search(ctx, criteria)
```

## Configuration

### Service Configuration
```go
config := ServiceConfig{
    BackgroundConfig:   backgroundConfig,
    DefaultSearchLimit: 100,
    MaxSearchLimit:     1000,
    BatchSize:          500,
    ProcessingTimeout:  30 * time.Second,
    EnableCorrelation:  true,
    EnableCleanup:      true,
    EnableMetrics:      true,
}
```

### Background Processing Configuration
```go
backgroundConfig := BackgroundConfig{
    CorrelationInterval:    30 * time.Minute,
    CorrelationBatchHours:  24,
    LogRetentionDays:       90,
    AuditRetentionDays:     365,
    SnapshotRetentionDays:  30,
    CleanupInterval:        6 * time.Hour,
    CleanupBatchSize:       1000,
    MaxConcurrentTasks:     3,
}
```

## Usage Examples

### Basic Log Processing
```go
// Initialize service
service := NewService(repository, DefaultServiceConfig())
service.Start()

// Process a log entry
logEntry := &database.LogEntry{
    Timestamp: time.Now(),
    MessageID: stringPtr("1rABC-123456-78"),
    LogType:   database.LogTypeMain,
    Event:     database.EventArrival,
    Sender:    stringPtr("user@example.com"),
    RawLine:   "2024-01-15 10:30:00 1rABC-123456-78 <= user@example.com...",
}

err := service.ProcessLogEntry(ctx, logEntry)
```

### Advanced Search
```go
// Search for messages from specific sender in last 24 hours
criteria := SearchCriteria{
    StartTime: timePtr(time.Now().Add(-24 * time.Hour)),
    EndTime:   timePtr(time.Now()),
    Sender:    "important@example.com",
    LogTypes:  []string{database.LogTypeMain},
    Events:    []string{database.EventArrival, database.EventDelivery},
    Limit:     50,
    SortBy:    "timestamp",
    SortOrder: "desc",
}

results, err := service.SearchLogs(ctx, criteria)
if err == nil {
    fmt.Printf("Found %d entries (total: %d)\n", 
        len(results.Entries), results.TotalCount)
    
    // Access aggregations
    if results.Aggregations != nil {
        fmt.Printf("Event counts: %+v\n", results.Aggregations.EventCounts)
        fmt.Printf("Top senders: %+v\n", results.Aggregations.TopSenders)
    }
}
```

### Message Correlation
```go
// Get complete message history and correlation
correlation, err := service.GetMessageCorrelation(ctx, "1rABC-123456-78")
if err == nil {
    fmt.Printf("Message: %s\n", correlation.MessageID)
    fmt.Printf("Recipients: %d total, %d delivered\n", 
        correlation.Summary.TotalRecipients,
        correlation.Summary.DeliveredCount)
    
    // Print timeline
    for _, event := range correlation.Timeline {
        fmt.Printf("%s: %s - %s\n", 
            event.Timestamp.Format("15:04:05"),
            event.Event, 
            event.Description)
    }
}
```

### Background Processing
```go
// Manual correlation trigger
err := service.TriggerCorrelation(ctx, 
    time.Now().Add(-1*time.Hour), 
    time.Now())

// Get service status
status := service.GetServiceStatus()
fmt.Printf("Background service running: %t\n", 
    status.BackgroundStatus.Running)

// Get retention information
retentionInfo, err := service.GetRetentionInfo(ctx)
if err == nil {
    fmt.Printf("Log retention: %d days\n", 
        retentionInfo.Config.LogRetentionDays)
    fmt.Printf("Current log entries: %d\n", 
        retentionInfo.CurrentCounts["log_entries"])
}
```

## Data Structures

### Search Criteria
```go
type SearchCriteria struct {
    // Time range
    StartTime *time.Time
    EndTime   *time.Time
    
    // Message filtering
    MessageID    string
    Sender       string
    Recipients   []string
    
    // Log filtering
    LogTypes     []string
    Events       []string
    Status       string
    
    // Content filtering
    Keywords     []string
    ErrorCode    string
    Host         string
    
    // Size filtering
    MinSize      *int64
    MaxSize      *int64
    
    // Pagination and sorting
    Limit        int
    Offset       int
    SortBy       string
    SortOrder    string
}
```

### Message Correlation
```go
type MessageCorrelation struct {
    MessageID    string
    Message      *database.Message
    Recipients   []database.Recipient
    LogEntries   []database.LogEntry
    Attempts     []database.DeliveryAttempt
    Timeline     []TimelineEvent
    Summary      MessageSummary
}

type MessageSummary struct {
    FirstSeen       time.Time
    LastActivity    time.Time
    TotalRecipients int
    DeliveredCount  int
    DeferredCount   int
    BouncedCount    int
    AttemptCount    int
    FinalStatus     string
    Duration        string
}
```

### Search Results
```go
type SearchResult struct {
    Entries      []database.LogEntry
    TotalCount   int
    HasMore      bool
    SearchTime   time.Duration
    Aggregations *SearchAggregations
}

type SearchAggregations struct {
    EventCounts        map[string]int
    LogTypeCounts      map[string]int
    StatusCounts       map[string]int
    HourlyDistribution map[string]int
    TopSenders         []SenderCount
    TopHosts           []HostCount
}
```

## Performance Considerations

### Indexing
The package relies on proper database indexing for performance:
- `idx_log_entries_timestamp` - For time-based queries
- `idx_log_entries_message_id` - For message correlation
- `idx_log_entries_event` - For event filtering
- `idx_log_entries_sender` - For sender searches

### Batch Processing
- Log entries are processed in configurable batches (default: 500)
- Background correlation processes data in time-based batches
- Cleanup operations use batch sizes to avoid long-running transactions

### Memory Management
- Search results are paginated to control memory usage
- Background tasks use streaming processing where possible
- Configurable limits prevent excessive resource consumption

## Error Handling

The package implements comprehensive error handling:
- Retry logic for transient database errors
- Graceful degradation when optional features fail
- Detailed error logging with context
- Circuit breaker patterns for background tasks

## Testing

The package includes comprehensive tests:
- Unit tests for all major components
- Integration tests with test database
- Benchmark tests for performance validation
- Example tests demonstrating usage patterns

Run tests with:
```bash
go test ./internal/logprocessor/...
go test -bench=. ./internal/logprocessor/...
```

## Integration

### With Log Monitor
```go
// Set up log processor in monitor
monitor := logmonitor.NewLogMonitor(config)
monitor.SetLogProcessor(service)
```

### With API Layer
```go
// Use in API handlers
func (h *Handler) SearchLogs(w http.ResponseWriter, r *http.Request) {
    criteria := parseSearchCriteria(r)
    results, err := h.logProcessor.SearchLogs(r.Context(), criteria)
    // ... handle response
}
```

## Monitoring and Metrics

The background service collects metrics:
- Database table sizes
- Processing rates
- Error counts
- Performance statistics

Access metrics through:
```go
status := service.GetServiceStatus()
retentionInfo, _ := service.GetRetentionInfo(ctx)
```

## Best Practices

1. **Configuration**: Use appropriate retention periods based on storage capacity
2. **Batch Sizes**: Adjust batch sizes based on system performance
3. **Correlation**: Enable correlation for better message tracking
4. **Cleanup**: Enable automatic cleanup to manage storage growth
5. **Monitoring**: Regularly check service status and metrics
6. **Search Limits**: Use reasonable search limits to prevent performance issues

## Troubleshooting

### Common Issues

1. **High Memory Usage**: Reduce batch sizes and search limits
2. **Slow Searches**: Check database indexes and query patterns
3. **Background Tasks Failing**: Check database connectivity and permissions
4. **Storage Growth**: Verify cleanup is enabled and retention periods are appropriate

### Debug Information

Enable debug logging to troubleshoot issues:
```go
log.SetLevel(log.DebugLevel)
```

Check service status:
```go
status := service.GetServiceStatus()
fmt.Printf("Service status: %+v\n", status)
```