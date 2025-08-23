# Requirements Document

## Introduction

The Exim Control Panel (Exim-Pilot) is a comprehensive web-based management interface for Exim mail servers running on Ubuntu/Debian systems. The application provides real-time monitoring, queue management, log analysis, and deliverability reporting through an intuitive web portal. Built with Go backend and SQLite database, it processes Exim logs from standard system locations and offers both individual and bulk operations for efficient mail server administration.

## Requirements

### Requirement 1: Mail Queue Management

**User Story:** As a mail administrator, I want to view and manage the mail queue with comprehensive filtering and search capabilities, so that I can efficiently monitor and troubleshoot mail delivery issues.

#### Acceptance Criteria

1. WHEN the queue viewer is accessed THEN the system SHALL display a paginated list of all queued messages
2. WHEN displaying queue messages THEN the system SHALL show message ID, envelope sender, recipients, size, age, status (queued/deferred/frozen), and retry count
3. WHEN a user searches the queue THEN the system SHALL filter by sender, recipient, message-id, subject, age, status, or retry count
4. WHEN queue health is requested THEN the system SHALL display total queued messages, deferred count, oldest message age, and recent growth trends
5. WHEN bulk actions are selected THEN the system SHALL allow multi-select operations across many messages

### Requirement 2: Message Operations and Control

**User Story:** As a mail administrator, I want to perform individual and bulk operations on queued messages, so that I can manage problematic messages and optimize delivery performance.

#### Acceptance Criteria

1. WHEN a message is selected for inspection THEN the system SHALL display envelope, SMTP transaction log excerpts, full headers, and optionally truncated raw body
2. WHEN "deliver now" is triggered THEN the system SHALL force immediate delivery attempt bypassing scheduler delays
3. WHEN a message is paused/frozen THEN the system SHALL prevent retry attempts until explicitly resumed
4. WHEN a message is resumed/thawed THEN the system SHALL return it to normal retry scheduling
5. WHEN a message is deleted THEN the system SHALL permanently remove it with optional reason logging
6. WHEN bulk operations are performed THEN the system SHALL execute deliver now, freeze, thaw, or delete across selected messages
7. WHEN any queue operation is performed THEN the system SHALL record the action in an audit log with actor and timestamp

### Requirement 3: Logging and Monitoring

**User Story:** As a mail administrator, I want comprehensive access to Exim logs with real-time monitoring and export capabilities, so that I can troubleshoot issues and maintain audit trails.

#### Acceptance Criteria

1. WHEN accessing transaction logs THEN the system SHALL display chronological SMTP transactions and delivery activity
2. WHEN viewing reject logs THEN the system SHALL show rejected messages/connections with reason codes and timestamps
3. WHEN checking panic/error logs THEN the system SHALL display daemon-level errors and critical failures
4. WHEN viewing per-message logs THEN the system SHALL aggregate all log lines for a specific message-id into a timeline
5. WHEN real-time monitoring is enabled THEN the system SHALL stream incoming log entries with keyword/message-id filtering
6. WHEN log export is requested THEN the system SHALL provide CSV/TXT downloads of selected log slices
7. WHEN administrative actions occur THEN the system SHALL maintain an immutable audit trail

### Requirement 4: Delivery History and Tracking

**User Story:** As a mail administrator, I want detailed delivery attempt history and retry visualization, so that I can understand message delivery patterns and troubleshoot delivery failures.

#### Acceptance Criteria

1. WHEN viewing delivery attempts THEN the system SHALL show timestamp, destination IP, SMTP response, outcome, and sequence number for each attempt
2. WHEN displaying retry timeline THEN the system SHALL provide graphical visualization of past attempts and scheduled future retries
3. WHEN classifying delivery outcomes THEN the system SHALL clearly separate permanent bounces from temporary deferrals with rationale
4. WHEN handling multi-recipient messages THEN the system SHALL show individual recipient outcomes and statuses
5. WHEN tracing message delivery THEN the system SHALL provide searchable consolidated delivery path from receive to final status

### Requirement 5: Message Content Inspection

**User Story:** As a mail administrator, I want safe access to message content and headers, so that I can verify routing metadata and diagnose content-related issues without security risks.

#### Acceptance Criteria

1. WHEN viewing message headers THEN the system SHALL display full RFC-822 headers and raw message source
2. WHEN previewing message content THEN the system SHALL safely render message body and stripped attachments
3. WHEN inspecting routing metadata THEN the system SHALL show SPF/DKIM/DMARC results and Received headers
4. WHEN adding troubleshooting notes THEN the system SHALL allow operator notes/tags on message or recipient records
5. WHEN handling large messages THEN the system SHALL provide rate-limited previews to avoid UI performance issues

### Requirement 6: Deliverability Reporting and Analytics

**User Story:** As a mail administrator, I want comprehensive deliverability reports and trend analysis, so that I can monitor mail server performance and identify delivery issues proactively.

#### Acceptance Criteria

1. WHEN accessing the deliverability dashboard THEN the system SHALL show success/defer/bounce rates over selectable time ranges
2. WHEN viewing volume trends THEN the system SHALL display ranked lists of top senders/recipients by traffic, defers, or bounces
3. WHEN analyzing failures THEN the system SHALL categorize failure types with counts and sample log snippets
4. WHEN reviewing bounce history THEN the system SHALL provide per-domain/account bounce counts and common bounce codes
5. WHEN generating reports THEN the system SHALL support scheduled/on-demand CSV/PDF exports
6. WHEN correlating incidents THEN the system SHALL link queue spikes to recent events using correlated timelines
7. WHEN displaying weekly overview THEN the system SHALL show bar graphs for delivered, failed, pending, and deferred emails with delivered messages removed from pending counts

### Requirement 7: System Integration and Data Collection

**User Story:** As a system administrator, I want the application to automatically collect and process Exim logs from standard Ubuntu/Debian locations, so that I can monitor mail server activity without manual log management.

#### Acceptance Criteria

1. WHEN the system starts THEN it SHALL automatically discover Exim log files in default Ubuntu/Debian paths (/var/log/exim4/)
2. WHEN processing logs THEN the system SHALL parse main log, reject log, and panic log formats
3. WHEN storing log data THEN the system SHALL use SQLite database for efficient querying and reporting
4. WHEN monitoring log files THEN the system SHALL detect new log entries in real-time
5. WHEN handling log rotation THEN the system SHALL continue processing across rotated log files
6. IF log files are inaccessible THEN the system SHALL log appropriate error messages and continue operation

### Requirement 8: Web Interface and User Experience

**User Story:** As a mail administrator, I want an intuitive web-based interface with responsive design, so that I can efficiently manage the mail server from any device.

#### Acceptance Criteria

1. WHEN accessing the web interface THEN the system SHALL provide a responsive design that works on desktop and mobile devices
2. WHEN navigating the interface THEN the system SHALL offer clear menu structure for all major features
3. WHEN performing operations THEN the system SHALL provide immediate feedback and confirmation dialogs for destructive actions
4. WHEN viewing large datasets THEN the system SHALL implement pagination and lazy loading for performance
5. WHEN using search and filters THEN the system SHALL provide real-time filtering with clear filter indicators
6. WHEN displaying the dashboard THEN the system SHALL show weekly overview graphs and key metrics prominently

### Requirement 9: Security and Access Control

**User Story:** As a system administrator, I want secure access controls and audit logging, so that I can ensure only authorized personnel can modify mail server state and track all administrative actions.

#### Acceptance Criteria

1. WHEN accessing the application THEN the system SHALL require authentication for all administrative functions
2. WHEN performing destructive operations THEN the system SHALL require confirmation and log the action with user identity
3. WHEN viewing sensitive message content THEN the system SHALL provide options to mask or redact personal information
4. WHEN handling file system access THEN the system SHALL run with minimal required privileges
5. WHEN logging audit events THEN the system SHALL create immutable records that cannot be modified post-creation

### Requirement 10: Performance and Scalability

**User Story:** As a system administrator, I want the application to handle large mail queues and log volumes efficiently, so that it remains responsive even under heavy mail server load.

#### Acceptance Criteria

1. WHEN processing large queues THEN the system SHALL maintain responsive UI through pagination and background processing
2. WHEN handling high log volumes THEN the system SHALL process logs efficiently without blocking the web interface
3. WHEN storing historical data THEN the system SHALL implement appropriate database indexing for fast queries
4. WHEN displaying real-time data THEN the system SHALL update efficiently without full page refreshes
5. WHEN managing database growth THEN the system SHALL provide configurable data retention policies