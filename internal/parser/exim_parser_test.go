package parser

import (
	"testing"

	"github.com/andreitelteu/exim-pilot/internal/database"
)

func TestEximParser_ParseLogLine(t *testing.T) {
	parser := NewEximParser()

	tests := []struct {
		name     string
		line     string
		logType  string
		wantErr  bool
		wantNil  bool
		expected *database.LogEntry
	}{
		{
			name:    "Message arrival",
			line:    "2024-01-15 10:30:45 1rABCD-123456-78 <= sender@example.com H=mail.example.com [192.168.1.1] P=esmtp S=1234",
			logType: database.LogTypeMain,
			wantErr: false,
			wantNil: false,
			expected: &database.LogEntry{
				MessageID: testStringPtr("1rABCD-123456-78"),
				Event:     database.EventArrival,
				Host:      testStringPtr("mail.example.com"),
				Sender:    testStringPtr("sender@example.com"),
				Size:      testInt64Ptr(1234),
				Status:    testStringPtr("received"),
			},
		},
		{
			name:    "Message delivery",
			line:    "2024-01-15 10:31:00 1rABCD-123456-78 => recipient@example.com R=dnslookup T=remote_smtp H=mx.example.com [192.168.1.2]",
			logType: database.LogTypeMain,
			wantErr: false,
			wantNil: false,
			expected: &database.LogEntry{
				MessageID:  testStringPtr("1rABCD-123456-78"),
				Event:      database.EventDelivery,
				Host:       testStringPtr("mx.example.com"),
				Recipients: []string{"recipient@example.com"},
				Status:     testStringPtr("delivered"),
			},
		},
		{
			name:    "Message deferral",
			line:    "2024-01-15 10:31:00 1rABCD-123456-78 == recipient@example.com R=dnslookup T=remote_smtp defer (-1): Connection refused",
			logType: database.LogTypeMain,
			wantErr: false,
			wantNil: false,
			expected: &database.LogEntry{
				MessageID:  testStringPtr("1rABCD-123456-78"),
				Event:      database.EventDefer,
				Recipients: []string{"recipient@example.com"},
				Status:     testStringPtr("deferred"),
				ErrorCode:  testStringPtr("-1"),
				ErrorText:  testStringPtr("Connection refused"),
			},
		},
		{
			name:    "Connection rejected",
			line:    "2024-01-15 10:30:45 rejected connection from [192.168.1.100]: (tcp wrappers)",
			logType: database.LogTypeReject,
			wantErr: false,
			wantNil: false,
			expected: &database.LogEntry{
				Event:     database.EventReject,
				Host:      testStringPtr("192.168.1.100"),
				Status:    testStringPtr("rejected"),
				ErrorText: testStringPtr("(tcp wrappers)"),
			},
		},
		{
			name:    "Empty line",
			line:    "",
			logType: database.LogTypeMain,
			wantErr: false,
			wantNil: true,
		},
		{
			name:    "Unknown log type",
			line:    "2024-01-15 10:30:45 some log line",
			logType: "unknown",
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseLogLine(tt.line, tt.logType)

			if tt.wantErr && err == nil {
				t.Errorf("ParseLogLine() expected error but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ParseLogLine() unexpected error: %v", err)
				return
			}

			if tt.wantNil && result != nil {
				t.Errorf("ParseLogLine() expected nil result but got: %+v", result)
				return
			}

			if !tt.wantNil && result == nil {
				t.Errorf("ParseLogLine() expected result but got nil")
				return
			}

			if result != nil && tt.expected != nil {
				// Check key fields
				if tt.expected.MessageID != nil && (result.MessageID == nil || *result.MessageID != *tt.expected.MessageID) {
					t.Errorf("ParseLogLine() MessageID = %v, want %v", result.MessageID, tt.expected.MessageID)
				}

				if result.Event != tt.expected.Event {
					t.Errorf("ParseLogLine() Event = %v, want %v", result.Event, tt.expected.Event)
				}

				if tt.expected.Host != nil && (result.Host == nil || *result.Host != *tt.expected.Host) {
					t.Errorf("ParseLogLine() Host = %v, want %v", result.Host, tt.expected.Host)
				}

				if tt.expected.Sender != nil && (result.Sender == nil || *result.Sender != *tt.expected.Sender) {
					t.Errorf("ParseLogLine() Sender = %v, want %v", result.Sender, tt.expected.Sender)
				}

				if tt.expected.Size != nil && (result.Size == nil || *result.Size != *tt.expected.Size) {
					t.Errorf("ParseLogLine() Size = %v, want %v", result.Size, tt.expected.Size)
				}

				if tt.expected.Status != nil && (result.Status == nil || *result.Status != *tt.expected.Status) {
					t.Errorf("ParseLogLine() Status = %v, want %v", result.Status, tt.expected.Status)
				}
			}
		})
	}
}

func TestEximParser_ExtractMessageID(t *testing.T) {
	parser := NewEximParser()

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Valid message ID",
			line:     "2024-01-15 10:30:45 1rABCD-123456-78 <= sender@example.com",
			expected: "1rABCD-123456-78",
		},
		{
			name:     "No message ID",
			line:     "2024-01-15 10:30:45 rejected connection from [192.168.1.100]",
			expected: "",
		},
		{
			name:     "Multiple message IDs",
			line:     "2024-01-15 10:30:45 1rABCD-123456-78 related to 1rDEFG-789012-34",
			expected: "1rABCD-123456-78", // Should return the first one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractMessageID(tt.line)
			if result != tt.expected {
				t.Errorf("ExtractMessageID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper functions for tests

func testStringPtr(s string) *string {
	return &s
}

func testInt64Ptr(i int64) *int64 {
	return &i
}
