### Project name: Exim-Pilot

#### Mail Queue Viewer
A searchable, paginated list of all queued messages showing ID, envelope sender, recipients, size, age, status (queued/deferred/frozen), and retry count for quick triage.

#### Inspect Queued Message
View a selected message's envelope, SMTP transaction log excerpts, full headers, and (optionally truncated) raw body to diagnose routing, authentication, or content issues.

#### Deliver Now / Force Retry
Trigger an immediate delivery attempt for one or more queued messages (bypassing regular scheduler delays) and record the action in an audit log.

#### Pause / Freeze Message
Temporarily prevent a specific queued message from being retried (hold in queue) so it will not be processed until explicitly resumed.

#### Resume / Thaw Message
Remove the hold on a frozen message so it returns to normal retry scheduling and is eligible for immediate or scheduled delivery attempts.

#### Remove / Delete Message
Permanently remove selected queued messages (with an optional reason), useful for stopping spam runs or clearing corrupted messages.

#### Bulk Queue Actions
Perform multi-select operations (deliver now, freeze, thaw, delete) across many messages to efficiently manage large or problematic queues.

#### Queue Search & Filters
Search and filter queue by sender, recipient, message-id, subject, age, status, or retry count to rapidly locate specific messages or problem classes.

#### Queue Health Indicators
At-a-glance metrics such as total queued messages, number of deferred messages, oldest message age, and recent queue growth trends to surface issues early.

---

### Logging & reports

#### Transaction Log (main mail log)
Chronological record of SMTP transactions and delivery activity for all messages; used to reconstruct delivery flows and inspect SMTP responses from remote hosts.

#### Reject Log
Records messages or connections rejected during SMTP acceptance (e.g., policy/ACL rejections) with reason codes and timestamps for troubleshooting rejections.

#### Panic / Error Log
Captures daemon-level errors and critical Exim failures that require investigation (stack traces, fatal conditions, service errors).

#### Per-Message Log View
Aggregated set of all log lines related to a specific message-id (or queue ID), assembled into a coherent timeline for quick deep-dive analysis.

#### Real-time Log Tail / Live Feed
Streaming view of incoming log entries (with filtering by keyword, message-id or PID) to monitor live delivery attempts and immediate system behavior.

#### Log Export & Download
Ability to export selected log slices or report data (CSV, TXT) for offline analysis, SIEM ingestion, or long-term archiving.

#### Audit Trail of Administrative Actions
Immutable record of UI/CLI actions (force retry, delete, freeze, thaw) including actor, timestamp, and affected message-ids for accountability and post‑mortem reviews.

---

### Retry & delivery history

#### Delivery Attempts History
Detailed list of each delivery attempt for a message with timestamp, destination IP, SMTP response code/text, outcome (success/defer/bounce), and attempt sequence number.

#### Retry Timeline Visualization
Graphical timeline showing past attempts and scheduled future retries for deferred messages to understand retry cadence and next expected action.

#### Bounce vs Deferred Classification
Clear labeling and separation between permanent bounces and temporary deferrals, with the rationale (SMTP response class, error text) surfaced for each event.

#### Per-Recipient Delivery Status
For multi-recipient messages, show individual recipient outcomes, so one message can be partially delivered while other recipients remain deferred or bounced.

---

### Tracking & message inspection

#### Message Trace / Track Delivery
Searchable trace tool (by sender, recipient, message-id, or subject) that returns a consolidated delivery path: receive → queue events → delivery attempts → final status.

#### Threaded Delivery Timeline
Chronological, human-readable timeline of events for a message (receive, queued, retries, remote rejections, final disposition) to simplify root-cause analysis.

#### View Headers & Raw Message
Display the full RFC‑822 headers and raw message source (optionally masked) so operators can verify SPF/DKIM/DMARC results, Received headers, and routing metadata.

#### Attachment & Content Preview (safe mode)
Safe rendering of message body and stripped/previewed attachments for quick inspection while avoiding automatic execution of active content.

#### Recipient-level Troubleshooting Notes
Attach short operator notes or tags to message or recipient records (e.g., “mailbox full — user contacted”) to track remediation steps.

---

### Deliverability reports & history

#### Deliverability Dashboard
Aggregate success/defer/bounce rates over selectable time ranges, broken down by domain, account, IP, or recipient to surface deliverability trends.

#### Top Senders / Top Recipients & Volume Trends
Ranked lists showing which senders or recipients generate most traffic, defers, or bounces to help identify abuse or misconfiguration at the source.

#### Failure Reason Breakdown
Categorized counts of failure types (connection refused, DNS NXDOMAIN, mailbox full, spam rejection, policy deny, timeout) with sample log snippets to speed triage.

#### Bounce Summary & History
Per-domain or per-account history of bounces (counts, recent examples, most common bounce codes) to prioritize mitigation or contact list hygiene.

#### Exportable Historical Reports
Scheduled or on‑demand exports of deliverability metrics and event lists (CSV/PDF) for compliance, reporting, or further analysis in BI tools.

#### Correlated Incident Views
Link queue spikes or failing delivery patterns to recent events (e.g., mass outbound batch at a timestamp) using correlated timelines to assist incident response.

---

### Notes & best-practice UI behaviors

- Every action that alters queue state (deliver now, delete, freeze/thaw) should be logged in the audit trail with actor and optional reason.
- Every action that alters queue state (deliver now, delete, freeze/thaw) should have bulk action capability
- Searchable message-id and SMTP-response fields dramatically reduce mean time to resolution for delivery problems.
- Provide rate-limited previews for large raw messages and attachments to avoid heavy UI load when inspecting many messages.
- Expose per-message log aggregation so operators don’t need to manually stitch log files when troubleshooting.