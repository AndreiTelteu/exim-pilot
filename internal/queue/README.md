# Queue Management System

This package provides comprehensive queue management functionality for the Exim Control Panel, including queue listing, message inspection, and queue operations.

## Features

### Queue Interface (`queue.go`)
- **Queue Listing**: Retrieve current queue status using `exim -bp`
- **Message Inspection**: Get detailed message information using `exim -Mvh`, `exim -Mvb`, `exim -Mvl`
- **Queue Snapshots**: Create historical snapshots for tracking queue trends
- **Queue Parsing**: Parse Exim queue output into structured data

### Queue Operations (`operations.go`)
- **Individual Operations**:
  - `DeliverNow`: Force immediate delivery (`exim -M`)
  - `FreezeMessage`: Freeze message (`exim -Mf`)
  - `ThawMessage`: Thaw frozen message (`exim -Mt`)
  - `DeleteMessage`: Remove message (`exim -Mrm`)

- **Bulk Operations**:
  - `BulkDeliverNow`: Deliver multiple messages
  - `BulkFreeze`: Freeze multiple messages
  - `BulkThaw`: Thaw multiple messages
  - `BulkDelete`: Delete multiple messages

- **Audit Logging**: All operations are logged with user context and timestamps

### Service Layer (`service.go`)
- **High-level API**: Simplified interface for queue management
- **Search Functionality**: Search messages by various criteria
- **Health Monitoring**: Queue health metrics and trends
- **Statistics**: Detailed queue statistics and breakdowns
- **Periodic Snapshots**: Background snapshot creation

## Usage Examples

### Basic Queue Operations

```go
// Initialize service
db, _ := database.NewConnection("exim-pilot.db")
service := queue.NewService("/usr/sbin/exim4", db)

// Get queue status
status, err := service.GetQueueStatus()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total messages: %d\n", status.TotalMessages)
```

### Message Operations

```go
// Force delivery of a message
result, err := service.DeliverNow("1a2b3c-000001-AB", "admin", "192.168.1.1")
if err != nil {
    log.Printf("Failed: %v", err)
} else if result.Success {
    fmt.Printf("Delivery initiated: %s\n", result.Message)
}

// Freeze a message
result, err = service.FreezeMessage("1a2b3c-000001-AB", "admin", "192.168.1.1")
```

### Bulk Operations

```go
messageIDs := []string{"msg1", "msg2", "msg3"}
result, err := service.BulkDelete(messageIDs, "admin", "192.168.1.1")

fmt.Printf("Deleted %d/%d messages\n", result.SuccessfulCount, result.TotalMessages)
```

### Search and Filtering

```go
criteria := &queue.SearchCriteria{
    Status: "deferred",
    MinSize: 1024,
    Sender: "example.com",
}

messages, err := service.SearchQueueMessages(criteria)
```

### Health Monitoring

```go
health, err := service.GetQueueHealth()
fmt.Printf("Queue health: %d total, %d deferred, %d frozen\n",
    health.TotalMessages, health.DeferredMessages, health.FrozenMessages)
```

## Data Structures

### QueueMessage
Represents a message in the Exim queue:
```go
type QueueMessage struct {
    ID          string    `json:"id"`
    Size        int64     `json:"size"`
    Age         string    `json:"age"`
    Sender      string    `json:"sender"`
    Recipients  []string  `json:"recipients"`
    Status      string    `json:"status"` // queued, deferred, frozen
    RetryCount  int       `json:"retry_count"`
    LastAttempt time.Time `json:"last_attempt"`
    NextRetry   time.Time `json:"next_retry"`
}
```

### OperationResult
Result of a queue operation:
```go
type OperationResult struct {
    Success   bool   `json:"success"`
    MessageID string `json:"message_id"`
    Operation string `json:"operation"`
    Message   string `json:"message"`
    Error     string `json:"error,omitempty"`
}
```

### BulkOperationResult
Result of a bulk operation:
```go
type BulkOperationResult struct {
    TotalMessages    int               `json:"total_messages"`
    SuccessfulCount  int               `json:"successful_count"`
    FailedCount      int               `json:"failed_count"`
    Results          []OperationResult `json:"results"`
    Operation        string            `json:"operation"`
}
```

## Error Handling

The package implements comprehensive error handling:

- **Command Failures**: Exim command execution errors are captured and returned
- **Parsing Errors**: Malformed queue output is handled gracefully
- **Validation**: Message IDs are validated before operations
- **Audit Logging**: Failed audit logging doesn't fail the operation
- **Bulk Operations**: Partial failures are tracked and reported

## Security Considerations

- **Audit Trail**: All operations are logged with user and IP context
- **Input Validation**: Message IDs are validated before processing
- **Command Injection**: Uses exec.Command with separate arguments to prevent injection
- **Minimal Privileges**: Designed to run with minimal required permissions

## Performance

- **Efficient Parsing**: Streaming approach for large queue outputs
- **Batch Operations**: Bulk operations reduce overhead
- **Database Indexing**: Proper indexing for audit log queries
- **Background Processing**: Snapshots run in background goroutines

## Integration

This package integrates with:
- **Database Layer**: For audit logging and snapshots
- **Log Processing**: For message correlation
- **REST API**: Via service layer methods
- **Web Frontend**: Through API endpoints

## Configuration

The queue manager can be configured with:
- **Exim Path**: Custom path to exim binary (defaults to `/usr/sbin/exim4`)
- **Database Connection**: For audit logging and snapshots
- **Snapshot Interval**: For periodic queue snapshots

## Requirements Satisfied

This implementation satisfies the following requirements:

- **1.1**: Queue listing with pagination and filtering
- **1.2**: Message details display (ID, sender, recipients, size, age, status, retry count)
- **1.3**: Queue search and filtering capabilities
- **1.4**: Queue health metrics and trends
- **2.2**: Deliver now functionality
- **2.3**: Freeze/thaw operations
- **2.4**: Message deletion
- **2.5**: Bulk operations support
- **2.6**: Audit logging for all operations
- **2.7**: Operation result tracking and reporting

## Testing

Example files are provided for testing and demonstration:
- `example_test.go`: Basic functionality examples
- `operations_example.go`: Comprehensive operation examples

Note: Unit tests are not included as per project requirements.