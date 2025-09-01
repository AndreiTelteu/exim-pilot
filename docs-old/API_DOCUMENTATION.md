# Exim Control Panel API Documentation

## Overview

The Exim Control Panel provides a RESTful API for programmatic access to all mail server management functions. The API follows REST conventions and returns JSON responses for all endpoints.

### Base URL

```
http://your-server:8080/api/v1
```

### Authentication

All API endpoints require authentication. Include credentials in your requests using one of these methods:

**Session-based (recommended for web applications):**
```bash
# Login first to establish session
curl -X POST http://your-server:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}' \
  -c cookies.txt

# Use session cookie for subsequent requests
curl -H "Content-Type: application/json" \
  -b cookies.txt \
  http://your-server:8080/api/v1/dashboard
```

**Basic Authentication:**
```bash
curl -u username:password http://your-server:8080/api/v1/dashboard
```

### Response Format

All API responses follow this standard format:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 150,
    "total_pages": 3
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Error message description",
  "code": "ERROR_CODE"
}
```

### Rate Limiting

API requests are limited to 1000 requests per hour per authenticated user. Rate limit headers are included in responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Authentication Endpoints

### POST /auth/login

Authenticate user and establish session.

**Request:**
```json
{
  "username": "admin",
  "password": "your-password"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "username": "admin",
      "role": "administrator"
    },
    "session_expires": "2024-01-01T12:00:00Z"
  }
}
```

### POST /auth/logout

Terminate current session.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully"
  }
}
```

## Dashboard Endpoints

### GET /dashboard

Get dashboard metrics and overview data.

**Response:**
```json
{
  "success": true,
  "data": {
    "queue": {
      "total": 1250,
      "deferred": 45,
      "frozen": 3,
      "oldest_message_age": 3600,
      "recent_growth": 15
    },
    "delivery": {
      "delivered_today": 8945,
      "failed_today": 123,
      "pending_today": 67,
      "success_rate": 98.6
    },
    "system": {
      "log_entries_today": 15678,
      "uptime": 86400,
      "last_updated": "2024-01-01T12:00:00Z"
    }
  }
}
```

### GET /dashboard/weekly

Get weekly overview data for charts.

**Query Parameters:**
- `weeks` (optional): Number of weeks to include (default: 4)

**Response:**
```json
{
  "success": true,
  "data": {
    "dates": ["2024-01-01", "2024-01-02", "2024-01-03"],
    "delivered": [1200, 1350, 1180],
    "failed": [45, 67, 32],
    "pending": [23, 45, 12],
    "deferred": [12, 8, 15]
  }
}
```

## Queue Management Endpoints

### GET /queue

List messages in the mail queue.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 50, max: 200)
- `status` (optional): Filter by status (queued, deferred, frozen)
- `sort` (optional): Sort field (age, size, sender, recipients)
- `order` (optional): Sort order (asc, desc)

**Example Request:**
```bash
curl -u admin:password \
  "http://your-server:8080/api/v1/queue?page=1&per_page=25&status=deferred&sort=age&order=desc"
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "1hKj4x-0008Oi-3r",
      "size": 2048,
      "age": "2h 15m",
      "sender": "user@example.com",
      "recipients": ["recipient@domain.com"],
      "status": "deferred",
      "retry_count": 3,
      "last_attempt": "2024-01-01T10:30:00Z",
      "next_retry": "2024-01-01T12:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 45,
    "total_pages": 2
  }
}
```

### POST /queue/search

Advanced queue search with multiple filters.

**Request:**
```json
{
  "filters": {
    "sender": "user@example.com",
    "recipient": "*@domain.com",
    "age_min": "1h",
    "age_max": "24h",
    "status": ["deferred", "frozen"],
    "retry_count_min": 2
  },
  "page": 1,
  "per_page": 50,
  "sort": "age",
  "order": "desc"
}
```

**Response:** Same format as GET /queue

### GET /queue/{id}

Get detailed information about a specific message.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "1hKj4x-0008Oi-3r",
    "envelope": {
      "sender": "user@example.com",
      "recipients": ["recipient@domain.com"],
      "size": 2048,
      "received": "2024-01-01T08:15:00Z"
    },
    "headers": {
      "From": "User <user@example.com>",
      "To": "recipient@domain.com",
      "Subject": "Test Message",
      "Message-ID": "<123456@example.com>",
      "Date": "Mon, 1 Jan 2024 08:15:00 +0000"
    },
    "delivery_attempts": [
      {
        "timestamp": "2024-01-01T08:16:00Z",
        "host": "mx.domain.com",
        "ip": "192.168.1.100",
        "status": "defer",
        "smtp_code": "451",
        "error_message": "Temporary failure"
      }
    ],
    "status": "deferred",
    "retry_count": 3,
    "next_retry": "2024-01-01T12:30:00Z"
  }
}
```

### POST /queue/{id}/deliver

Force immediate delivery of a message.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Delivery initiated for message 1hKj4x-0008Oi-3r",
    "operation_id": "op_123456"
  }
}
```

### POST /queue/{id}/freeze

Freeze (pause) a message to prevent retry attempts.

**Request (optional):**
```json
{
  "reason": "Investigating delivery issues"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Message 1hKj4x-0008Oi-3r has been frozen",
    "operation_id": "op_123457"
  }
}
```

### POST /queue/{id}/thaw

Thaw (resume) a frozen message.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Message 1hKj4x-0008Oi-3r has been thawed",
    "operation_id": "op_123458"
  }
}
```

### DELETE /queue/{id}

Permanently delete a message from the queue.

**Request (optional):**
```json
{
  "reason": "Spam message removal"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Message 1hKj4x-0008Oi-3r has been deleted",
    "operation_id": "op_123459"
  }
}
```

### POST /queue/bulk

Perform bulk operations on multiple messages.

**Request:**
```json
{
  "message_ids": ["1hKj4x-0008Oi-3r", "1hKj4x-0008Oj-4s"],
  "operation": "deliver",
  "reason": "Manual intervention required"
}
```

**Available Operations:**
- `deliver`: Force delivery of all selected messages
- `freeze`: Freeze all selected messages
- `thaw`: Thaw all selected messages
- `delete`: Delete all selected messages

**Response:**
```json
{
  "success": true,
  "data": {
    "operation_id": "bulk_op_123460",
    "total_messages": 2,
    "status": "initiated",
    "results": [
      {
        "message_id": "1hKj4x-0008Oi-3r",
        "success": true
      },
      {
        "message_id": "1hKj4x-0008Oj-4s",
        "success": true
      }
    ]
  }
}
```

## Log Management Endpoints

### GET /logs

Retrieve log entries with filtering and pagination.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 50, max: 200)
- `log_type` (optional): Filter by log type (main, reject, panic)
- `event` (optional): Filter by event type (arrival, delivery, defer, bounce)
- `message_id` (optional): Filter by message ID
- `start_date` (optional): Start date (ISO 8601 format)
- `end_date` (optional): End date (ISO 8601 format)
- `search` (optional): Text search within log entries

**Example Request:**
```bash
curl -u admin:password \
  "http://your-server:8080/api/v1/logs?log_type=main&event=defer&start_date=2024-01-01T00:00:00Z"
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 12345,
      "timestamp": "2024-01-01T08:15:30Z",
      "message_id": "1hKj4x-0008Oi-3r",
      "log_type": "main",
      "event": "defer",
      "host": "mx.domain.com",
      "sender": "user@example.com",
      "recipients": ["recipient@domain.com"],
      "error_code": "451",
      "error_text": "Temporary failure",
      "raw_line": "2024-01-01 08:15:30 1hKj4x-0008Oi-3r == recipient@domain.com R=dnslookup T=remote_smtp defer (-44) H=mx.domain.com [192.168.1.100]: SMTP error from remote mail server after RCPT TO:<recipient@domain.com>: 451 Temporary failure"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 1250,
    "total_pages": 25
  }
}
```

### POST /logs/search

Advanced log search with complex filters.

**Request:**
```json
{
  "filters": {
    "log_types": ["main", "reject"],
    "events": ["defer", "bounce"],
    "message_id": "1hKj4x-0008Oi-3r",
    "sender": "*@example.com",
    "recipient": "user@domain.com",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-02T00:00:00Z",
    "text_search": "temporary failure",
    "error_codes": ["451", "452"]
  },
  "page": 1,
  "per_page": 100,
  "sort": "timestamp",
  "order": "desc"
}
```

**Response:** Same format as GET /logs

### GET /logs/tail

WebSocket endpoint for real-time log monitoring.

**WebSocket URL:**
```
ws://your-server:8080/api/v1/logs/tail
```

**Connection Parameters:**
```javascript
const ws = new WebSocket('ws://your-server:8080/api/v1/logs/tail?log_type=main&message_id=1hKj4x-0008Oi-3r');

ws.onmessage = function(event) {
  const logEntry = JSON.parse(event.data);
  console.log('New log entry:', logEntry);
};
```

**Message Format:**
```json
{
  "type": "log_entry",
  "data": {
    "timestamp": "2024-01-01T08:15:30Z",
    "message_id": "1hKj4x-0008Oi-3r",
    "log_type": "main",
    "event": "delivery",
    "raw_line": "..."
  }
}
```

### GET /logs/export

Export filtered log entries.

**Query Parameters:**
- Same as GET /logs for filtering
- `format` (required): Export format (csv, txt)
- `filename` (optional): Custom filename

**Response:**
- Content-Type: application/octet-stream
- Content-Disposition: attachment; filename="logs_export.csv"

## Reporting Endpoints

### GET /reports/deliverability

Get deliverability metrics and statistics.

**Query Parameters:**
- `start_date` (optional): Start date (default: 7 days ago)
- `end_date` (optional): End date (default: now)
- `group_by` (optional): Group results by (hour, day, week)

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_messages": 10000,
      "delivered": 9850,
      "deferred": 100,
      "bounced": 50,
      "success_rate": 98.5,
      "defer_rate": 1.0,
      "bounce_rate": 0.5
    },
    "timeline": [
      {
        "date": "2024-01-01",
        "delivered": 1200,
        "deferred": 15,
        "bounced": 5
      }
    ],
    "top_senders": [
      {
        "sender": "newsletter@company.com",
        "total": 5000,
        "delivered": 4950,
        "success_rate": 99.0
      }
    ],
    "top_recipients": [
      {
        "domain": "gmail.com",
        "total": 3000,
        "delivered": 2980,
        "success_rate": 99.3
      }
    ]
  }
}
```

### GET /reports/volume

Get volume and traffic analysis.

**Query Parameters:**
- `start_date` (optional): Start date
- `end_date` (optional): End date
- `interval` (optional): Time interval (hour, day, week)

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_volume": 50000,
      "average_per_day": 7142,
      "peak_hour": "14:00",
      "peak_volume": 500
    },
    "timeline": [
      {
        "timestamp": "2024-01-01T00:00:00Z",
        "volume": 120,
        "delivered": 118,
        "failed": 2
      }
    ],
    "hourly_distribution": [
      {"hour": 0, "volume": 50},
      {"hour": 1, "volume": 30}
    ]
  }
}
```

### GET /reports/failures

Get failure analysis and bounce statistics.

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_failures": 500,
      "temporary_failures": 400,
      "permanent_failures": 100
    },
    "failure_categories": [
      {
        "category": "DNS Resolution",
        "count": 150,
        "percentage": 30.0,
        "sample_errors": [
          "Host not found",
          "DNS timeout"
        ]
      }
    ],
    "bounce_codes": [
      {
        "code": "550",
        "description": "Mailbox unavailable",
        "count": 75,
        "percentage": 15.0
      }
    ],
    "problem_domains": [
      {
        "domain": "problematic-domain.com",
        "total_attempts": 100,
        "failures": 95,
        "failure_rate": 95.0
      }
    ]
  }
}
```

## Message Tracing Endpoints

### GET /messages/{id}/trace

Get complete delivery trace for a message.

**Response:**
```json
{
  "success": true,
  "data": {
    "message_id": "1hKj4x-0008Oi-3r",
    "timeline": [
      {
        "timestamp": "2024-01-01T08:15:00Z",
        "event": "arrival",
        "description": "Message received from user@example.com",
        "details": {
          "sender": "user@example.com",
          "size": 2048,
          "host": "mail.example.com"
        }
      },
      {
        "timestamp": "2024-01-01T08:16:00Z",
        "event": "delivery_attempt",
        "description": "Delivery attempt to recipient@domain.com",
        "details": {
          "recipient": "recipient@domain.com",
          "host": "mx.domain.com",
          "ip": "192.168.1.100",
          "result": "defer",
          "smtp_code": "451",
          "error": "Temporary failure"
        }
      }
    ],
    "recipients": [
      {
        "address": "recipient@domain.com",
        "status": "deferred",
        "attempts": 3,
        "last_attempt": "2024-01-01T10:30:00Z",
        "next_retry": "2024-01-01T12:30:00Z"
      }
    ],
    "current_status": "deferred"
  }
}
```

### GET /messages/{id}/content

Get safe message content preview.

**Query Parameters:**
- `include_body` (optional): Include message body (default: false)
- `max_body_length` (optional): Maximum body length (default: 1000)

**Response:**
```json
{
  "success": true,
  "data": {
    "headers": {
      "From": "User <user@example.com>",
      "To": "recipient@domain.com",
      "Subject": "Test Message",
      "Date": "Mon, 1 Jan 2024 08:15:00 +0000"
    },
    "body_preview": "This is the beginning of the message content...",
    "body_truncated": true,
    "attachments": [
      {
        "filename": "document.pdf",
        "size": 1024000,
        "content_type": "application/pdf"
      }
    ],
    "safety_info": {
      "content_filtered": true,
      "suspicious_content": false
    }
  }
}
```

### POST /messages/{id}/notes

Add troubleshooting notes to a message.

**Request:**
```json
{
  "note": "Investigating delivery issues with this domain",
  "tags": ["investigation", "domain-issue"],
  "priority": "high"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "note_id": "note_123456",
    "message": "Note added successfully"
  }
}
```

### GET /messages/{id}/notes

Get all notes for a message.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "note_123456",
      "timestamp": "2024-01-01T12:00:00Z",
      "user": "admin",
      "note": "Investigating delivery issues with this domain",
      "tags": ["investigation", "domain-issue"],
      "priority": "high"
    }
  ]
}
```

## Audit and Security Endpoints

### GET /audit

Get audit log entries.

**Query Parameters:**
- `page` (optional): Page number
- `per_page` (optional): Items per page
- `start_date` (optional): Start date
- `end_date` (optional): End date
- `user` (optional): Filter by username
- `action` (optional): Filter by action type

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 12345,
      "timestamp": "2024-01-01T12:00:00Z",
      "user": "admin",
      "action": "message_delete",
      "message_id": "1hKj4x-0008Oi-3r",
      "details": {
        "reason": "Spam message removal",
        "ip_address": "192.168.1.50"
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 500,
    "total_pages": 10
  }
}
```

## Error Codes

### Common HTTP Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Application Error Codes

- `AUTH_REQUIRED`: Authentication required
- `INVALID_CREDENTIALS`: Invalid username or password
- `SESSION_EXPIRED`: Session has expired
- `PERMISSION_DENIED`: Insufficient permissions
- `MESSAGE_NOT_FOUND`: Message ID not found
- `INVALID_OPERATION`: Operation not allowed for message status
- `EXIM_ERROR`: Error executing Exim command
- `DATABASE_ERROR`: Database operation failed
- `VALIDATION_ERROR`: Input validation failed

## Rate Limiting

API requests are limited to prevent abuse:

- **Authenticated users**: 1000 requests per hour
- **Queue operations**: 100 operations per hour
- **Bulk operations**: 10 operations per hour
- **Export operations**: 5 exports per hour

Rate limit headers are included in all responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Window: 3600
```

## WebSocket Events

### Real-time Updates

Connect to WebSocket endpoints for real-time updates:

**Queue Updates:**
```javascript
const ws = new WebSocket('ws://your-server:8080/api/v1/queue/updates');
ws.onmessage = function(event) {
  const update = JSON.parse(event.data);
  // Handle queue changes
};
```

**Log Tail:**
```javascript
const ws = new WebSocket('ws://your-server:8080/api/v1/logs/tail');
ws.onmessage = function(event) {
  const logEntry = JSON.parse(event.data);
  // Handle new log entries
};
```

### Event Types

- `queue_update`: Queue status changed
- `message_operation`: Message operation completed
- `log_entry`: New log entry available
- `system_alert`: System alert or warning

## SDK Examples

### JavaScript/Node.js

```javascript
class EximControlPanelAPI {
  constructor(baseUrl, credentials) {
    this.baseUrl = baseUrl;
    this.credentials = credentials;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}/api/v1${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Basic ${btoa(this.credentials)}`,
        ...options.headers
      }
    });
    
    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }
    
    return response.json();
  }

  async getQueue(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/queue?${query}`);
  }

  async deliverMessage(messageId, reason = '') {
    return this.request(`/queue/${messageId}/deliver`, {
      method: 'POST',
      body: JSON.stringify({ reason })
    });
  }

  async searchLogs(filters) {
    return this.request('/logs/search', {
      method: 'POST',
      body: JSON.stringify(filters)
    });
  }
}

// Usage
const api = new EximControlPanelAPI('http://your-server:8080', 'admin:password');

// Get queue messages
const queue = await api.getQueue({ status: 'deferred', page: 1 });

// Force delivery
await api.deliverMessage('1hKj4x-0008Oi-3r', 'Manual intervention');

// Search logs
const logs = await api.searchLogs({
  filters: { log_type: 'main', event: 'defer' },
  page: 1,
  per_page: 100
});
```

### Python

```python
import requests
import json
from typing import Dict, Any, Optional

class EximControlPanelAPI:
    def __init__(self, base_url: str, username: str, password: str):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.auth = (username, password)
        self.session.headers.update({'Content-Type': 'application/json'})

    def request(self, endpoint: str, method: str = 'GET', data: Optional[Dict] = None) -> Dict[str, Any]:
        url = f"{self.base_url}/api/v1{endpoint}"
        response = self.session.request(method, url, json=data)
        response.raise_for_status()
        return response.json()

    def get_queue(self, **params) -> Dict[str, Any]:
        return self.request('/queue', params=params)

    def deliver_message(self, message_id: str, reason: str = '') -> Dict[str, Any]:
        return self.request(f'/queue/{message_id}/deliver', 'POST', {'reason': reason})

    def search_logs(self, filters: Dict[str, Any]) -> Dict[str, Any]:
        return self.request('/logs/search', 'POST', filters)

# Usage
api = EximControlPanelAPI('http://your-server:8080', 'admin', 'password')

# Get deferred messages
queue = api.get_queue(status='deferred', page=1)

# Force delivery
result = api.deliver_message('1hKj4x-0008Oi-3r', 'Manual intervention')

# Search logs
logs = api.search_logs({
    'filters': {'log_type': 'main', 'event': 'defer'},
    'page': 1,
    'per_page': 100
})
```

## Best Practices

### Authentication
- Use session-based authentication for web applications
- Store credentials securely and never in client-side code
- Implement proper session timeout handling

### Error Handling
- Always check the `success` field in responses
- Implement retry logic for transient errors (5xx status codes)
- Log API errors for debugging and monitoring

### Performance
- Use pagination for large result sets
- Implement client-side caching where appropriate
- Use WebSocket connections for real-time updates instead of polling

### Security
- Always use HTTPS in production
- Validate and sanitize all input data
- Implement proper rate limiting on the client side
- Never log sensitive information like passwords

### Monitoring
- Monitor API response times and error rates
- Set up alerts for high error rates or slow responses
- Track rate limit usage to avoid hitting limits

---

For additional support or questions about the API, please refer to the user manual or contact your system administrator. 