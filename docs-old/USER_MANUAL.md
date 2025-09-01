# Exim Control Panel (Exim-Pilot) User Manual

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Dashboard Overview](#dashboard-overview)
4. [Queue Management](#queue-management)
5. [Log Viewer](#log-viewer)
6. [Reports and Analytics](#reports-and-analytics)
7. [Message Tracing](#message-tracing)
8. [Security and Authentication](#security-and-authentication)
9. [Troubleshooting](#troubleshooting)
10. [FAQ](#faq)

## Introduction

Exim Control Panel (Exim-Pilot) is a comprehensive web-based management interface for Exim mail servers running on Ubuntu/Debian systems. It provides real-time monitoring, queue management, log analysis, and deliverability reporting through an intuitive web portal.

### Key Features

- **Real-time Queue Management**: View, search, and manage mail queue with bulk operations
- **Log Monitoring**: Real-time log viewing with filtering and search capabilities
- **Deliverability Reports**: Comprehensive analytics and trend analysis
- **Message Tracing**: Detailed delivery history and troubleshooting tools
- **Security**: Audit logging and secure access controls
- **Performance**: Optimized for large mail queues and high log volumes

## Getting Started

### System Requirements

- Ubuntu 18.04+ or Debian 9+
- Exim 4.90+
- 512MB RAM minimum (1GB recommended)
- Web browser with JavaScript enabled

### First Login

1. Open your web browser and navigate to `http://your-server:8080`
2. Enter your username and password (configured during installation)
3. You'll be redirected to the main dashboard

### Navigation

The interface consists of several main sections accessible via the navigation menu:

- **Dashboard**: Overview of mail server status and metrics
- **Queue**: Mail queue management and operations
- **Logs**: Log viewing and real-time monitoring
- **Reports**: Analytics and deliverability reports
- **Messages**: Message tracing and history

## Dashboard Overview

The dashboard provides a quick overview of your mail server's current status and recent activity.

### Key Metrics

- **Total Queue Messages**: Current number of messages in the queue
- **Deferred Messages**: Messages that failed delivery and are scheduled for retry
- **Frozen Messages**: Messages that have been paused and won't be retried
- **Oldest Message**: Age of the oldest message in the queue

### Weekly Overview Chart

The weekly chart displays email activity over the past 7 days:

- **Green bars**: Successfully delivered messages
- **Red bars**: Failed/bounced messages
- **Yellow bars**: Currently pending messages
- **Gray bars**: Deferred messages awaiting retry

### Real-time Updates

The dashboard automatically updates every 30 seconds to show current status. You can also click the refresh button for immediate updates.

## Queue Management

The Queue section allows you to view and manage all messages currently in the mail queue.

### Queue List

The queue list displays all queued messages with the following information:

- **Message ID**: Unique identifier for the message
- **Sender**: Envelope sender address
- **Recipients**: List of recipient addresses
- **Size**: Message size in bytes
- **Age**: How long the message has been in the queue
- **Status**: Current status (queued, deferred, frozen)
- **Retry Count**: Number of delivery attempts made

### Searching and Filtering

Use the search bar to filter messages by:

- Sender email address
- Recipient email address
- Message ID
- Subject line
- Age (e.g., ">1h", "<30m")
- Status (queued, deferred, frozen)
- Retry count

**Search Examples:**
- `sender:user@example.com` - Find messages from specific sender
- `recipient:*@domain.com` - Find messages to specific domain
- `age:>2h` - Find messages older than 2 hours
- `status:deferred` - Find only deferred messages
- `retry:>3` - Find messages with more than 3 retry attempts

### Message Operations

#### Individual Message Actions

Click on any message to view detailed information and perform actions:

- **Deliver Now**: Force immediate delivery attempt
- **Freeze**: Pause the message (stops retry attempts)
- **Thaw**: Resume a frozen message
- **Delete**: Permanently remove the message from queue

#### Bulk Operations

Select multiple messages using checkboxes and perform bulk actions:

1. Select messages using individual checkboxes or "Select All"
2. Choose an action from the bulk actions dropdown
3. Confirm the operation in the dialog box

**Available Bulk Actions:**
- Deliver Now (all selected messages)
- Freeze (pause all selected messages)
- Thaw (resume all selected messages)
- Delete (permanently remove all selected messages)

### Message Details

Click on a message ID to view detailed information:

- **Envelope Information**: Sender, recipients, message size
- **Headers**: Full RFC-822 message headers
- **SMTP Transaction Log**: Delivery attempt history
- **Content Preview**: Safe preview of message content (if enabled)
- **Delivery Timeline**: Visual timeline of delivery attempts

## Log Viewer

The Log Viewer provides access to Exim log files with powerful search and filtering capabilities.

### Log Types

- **Main Log**: Message arrivals, deliveries, and deferrals
- **Reject Log**: Rejected messages and connections
- **Panic Log**: Daemon errors and critical failures

### Viewing Logs

The log viewer displays entries in chronological order with:

- **Timestamp**: When the log entry was created
- **Message ID**: Associated message (if applicable)
- **Event Type**: Type of log event (arrival, delivery, defer, etc.)
- **Details**: Full log entry text

### Search and Filtering

Use the advanced search to filter logs by:

- **Date Range**: Specify start and end dates
- **Log Type**: Filter by main, reject, or panic logs
- **Message ID**: Find all entries for a specific message
- **Event Type**: Filter by event type (arrival, delivery, defer, bounce)
- **Text Search**: Search within log entry text

### Real-time Log Tail

The real-time tail feature shows new log entries as they occur:

1. Click "Start Tail" to begin real-time monitoring
2. Use filters to show only relevant entries
3. Click "Pause" to stop auto-scrolling
4. Click "Stop Tail" to end real-time monitoring

### Exporting Logs

Export filtered log entries for external analysis:

1. Apply desired filters
2. Click "Export" button
3. Choose format (CSV or TXT)
4. Download will begin automatically

## Reports and Analytics

The Reports section provides comprehensive analytics about your mail server's performance.

### Deliverability Report

Shows success rates and delivery metrics:

- **Success Rate**: Percentage of successfully delivered messages
- **Defer Rate**: Percentage of temporarily failed messages
- **Bounce Rate**: Percentage of permanently failed messages
- **Top Senders**: Highest volume senders
- **Top Recipients**: Highest volume recipients
- **Domain Analysis**: Per-domain deliverability statistics

### Volume Report

Displays traffic trends and patterns:

- **Hourly Volume**: Messages per hour over selected time period
- **Daily Volume**: Messages per day
- **Peak Hours**: Busiest times of day
- **Growth Trends**: Volume changes over time

### Failure Analysis

Analyzes delivery failures and bounce patterns:

- **Failure Categories**: Common failure types and counts
- **Bounce Codes**: SMTP response codes and frequencies
- **Problem Domains**: Domains with high failure rates
- **Retry Patterns**: Analysis of retry behavior

### Custom Reports

Generate custom reports by:

1. Selecting date range
2. Choosing metrics to include
3. Applying filters (sender, recipient, domain)
4. Exporting results (CSV, PDF)

## Message Tracing

The Message Tracing feature provides detailed delivery history for individual messages.

### Delivery Timeline

View the complete delivery path for any message:

- **Receipt**: When the message was received
- **Queue Entry**: When it entered the queue
- **Delivery Attempts**: Each attempt with timestamp and result
- **Final Status**: Current status or final delivery result

### Per-Recipient Tracking

For messages with multiple recipients, view individual recipient status:

- **Delivered**: Successfully delivered recipients
- **Deferred**: Recipients with temporary failures
- **Bounced**: Recipients with permanent failures

### SMTP Transaction Details

View detailed SMTP transaction information:

- **Connecting Host**: Remote server details
- **SMTP Responses**: Server responses for each command
- **Error Messages**: Detailed error information for failures
- **Timing Information**: Connection and transfer times

### Troubleshooting Notes

Add notes and tags to messages for troubleshooting:

1. Click "Add Note" on message details page
2. Enter troubleshooting information
3. Add tags for categorization
4. Notes are preserved in audit log

## Security and Authentication

### User Authentication

Exim Control Panel uses secure authentication:

- **Login Required**: All features require authentication
- **Session Management**: Automatic logout after inactivity
- **Secure Cookies**: Session cookies are secure and HTTP-only

### Audit Logging

All administrative actions are logged:

- **Queue Operations**: All message operations (deliver, freeze, delete)
- **User Actions**: Login/logout events
- **Configuration Changes**: Any system configuration modifications
- **Timestamps**: All actions include precise timestamps
- **User Context**: Actions are associated with the performing user

### Access Control

- **Read-Only Access**: View queues and logs without modification rights
- **Administrative Access**: Full queue management and system operations
- **Audit Access**: View audit logs and security events

### Data Security

- **Input Validation**: All user input is validated and sanitized
- **SQL Injection Protection**: Parameterized queries prevent SQL injection
- **XSS Prevention**: Output encoding prevents cross-site scripting
- **File System Security**: Minimal file system privileges

## Troubleshooting

### Common Issues

#### Queue Not Loading

**Symptoms**: Queue page shows "Loading..." indefinitely

**Possible Causes**:
- Exim not running or not accessible
- Permission issues accessing queue files
- Database connectivity problems

**Solutions**:
1. Check Exim status: `systemctl status exim4`
2. Verify file permissions on `/var/spool/exim4`
3. Check application logs for database errors
4. Restart Exim Control Panel service

#### Logs Not Updating

**Symptoms**: Log viewer shows old entries, real-time tail not working

**Possible Causes**:
- Log file permissions
- Log file rotation issues
- File system watcher problems

**Solutions**:
1. Check log file permissions: `ls -la /var/log/exim4/`
2. Verify log rotation configuration
3. Restart the application to reinitialize file watchers
4. Check disk space availability

#### Performance Issues

**Symptoms**: Slow page loading, timeouts

**Possible Causes**:
- Large queue size
- High log volume
- Insufficient system resources

**Solutions**:
1. Increase system memory if possible
2. Configure data retention policies
3. Use more specific search filters
4. Consider queue cleanup for old messages

#### Authentication Problems

**Symptoms**: Cannot log in, session expires quickly

**Possible Causes**:
- Incorrect credentials
- Session configuration issues
- Clock synchronization problems

**Solutions**:
1. Verify username and password
2. Check system time synchronization
3. Clear browser cookies and cache
4. Check application configuration

### Getting Help

If you encounter issues not covered in this guide:

1. Check the application logs for error messages
2. Review the system logs (`journalctl -u exim-pilot`)
3. Verify Exim configuration and status
4. Consult the API documentation for integration issues

### Log File Locations

- **Application Logs**: `/opt/exim-pilot/logs/exim-pilot.log`
- **Exim Logs**: `/var/log/exim4/mainlog`, `/var/log/exim4/rejectlog`, `/var/log/exim4/paniclog`
- **System Logs**: `journalctl -u exim-pilot`

## FAQ

### General Questions

**Q: Can I use Exim Control Panel with other mail servers?**
A: No, it's specifically designed for Exim mail servers and uses Exim-specific commands and log formats.

**Q: Does it work on CentOS/RHEL?**
A: Currently optimized for Ubuntu/Debian. CentOS/RHEL support may require configuration adjustments.

**Q: Can multiple users access the interface simultaneously?**
A: Yes, the interface supports multiple concurrent users with individual sessions.

### Queue Management

**Q: What happens when I delete a message?**
A: The message is permanently removed from the queue and cannot be recovered. This action is logged in the audit trail.

**Q: Can I modify message content?**
A: No, the interface is read-only for message content. You can only manage delivery (deliver, freeze, delete).

**Q: How often does the queue update?**
A: The queue list updates automatically every 30 seconds, or immediately when you perform operations.

### Logs and Monitoring

**Q: How long are logs retained?**
A: By default, log entries are kept for 90 days. This is configurable in the system settings.

**Q: Can I export logs for external analysis?**
A: Yes, you can export filtered logs in CSV or TXT format.

**Q: Does real-time monitoring affect performance?**
A: Real-time monitoring uses efficient WebSocket connections and has minimal performance impact.

### Reports and Analytics

**Q: How accurate are the deliverability statistics?**
A: Statistics are based on actual log entries and are highly accurate. They reflect the server's actual delivery performance.

**Q: Can I schedule automatic reports?**
A: Currently, reports are generated on-demand. Scheduled reporting may be added in future versions.

**Q: What time zone are reports generated in?**
A: Reports use the server's local time zone. Timestamps in the interface show the server's time.

### Security

**Q: Are message contents secure?**
A: Message content viewing is optional and includes safety measures. Full content is never stored in the database.

**Q: Can I integrate with LDAP/Active Directory?**
A: Currently uses simple authentication. LDAP integration is planned for future versions.

**Q: Are audit logs tamper-proof?**
A: Audit logs are designed to be immutable once created, providing a reliable audit trail.

---

For additional support or feature requests, please consult your system administrator or refer to the technical documentation.