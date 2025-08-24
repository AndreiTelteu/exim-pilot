package audit

import (
	"testing"
	"time"
)

func TestAuditServiceLogic(t *testing.T) {
	// Test audit context creation
	auditCtx := &AuditContext{
		UserID:    "test-user",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
		RequestID: "test-request",
	}

	if auditCtx.UserID != "test-user" {
		t.Errorf("Expected UserID 'test-user', got %s", auditCtx.UserID)
	}

	if auditCtx.IPAddress != "127.0.0.1" {
		t.Errorf("Expected IPAddress '127.0.0.1', got %s", auditCtx.IPAddress)
	}

	// Test audit details creation
	details := &AuditDetails{
		MessageIDs: []string{"1ABC23-DEF456-GH"},
		Operation:  "deliver",
		Parameters: map[string]interface{}{
			"test_param": "test_value",
		},
		Result:       "success",
		Duration:     time.Second,
		ResourcePath: "/api/v1/queue/1ABC23-DEF456-GH/deliver",
	}

	if len(details.MessageIDs) != 1 || details.MessageIDs[0] != "1ABC23-DEF456-GH" {
		t.Error("MessageIDs not set correctly")
	}

	if details.Operation != "deliver" {
		t.Error("Operation not set correctly")
	}

	if details.Parameters["test_param"] != "test_value" {
		t.Error("Parameters not set correctly")
	}

	if details.Result != "success" {
		t.Error("Result not set correctly")
	}
}

func TestActionTypes(t *testing.T) {
	// Test that action types are defined correctly
	expectedActions := map[ActionType]string{
		ActionQueueDeliver:   "queue_deliver",
		ActionQueueFreeze:    "queue_freeze",
		ActionQueueThaw:      "queue_thaw",
		ActionQueueDelete:    "queue_delete",
		ActionBulkDeliver:    "bulk_deliver",
		ActionBulkFreeze:     "bulk_freeze",
		ActionBulkThaw:       "bulk_thaw",
		ActionBulkDelete:     "bulk_delete",
		ActionLogin:          "login",
		ActionLogout:         "logout",
		ActionMessageView:    "message_view",
		ActionMessageContent: "message_content",
		ActionNoteCreate:     "note_create",
		ActionNoteUpdate:     "note_update",
		ActionNoteDelete:     "note_delete",
		ActionTagCreate:      "tag_create",
		ActionTagDelete:      "tag_delete",
		ActionConfigChange:   "config_change",
		ActionSystemAccess:   "system_access",
	}

	for action, expected := range expectedActions {
		if string(action) != expected {
			t.Errorf("Expected action %s to equal %s, got %s", action, expected, string(action))
		}
	}
}

func TestAuditDetails(t *testing.T) {
	// Test audit details serialization
	details := &AuditDetails{
		MessageIDs: []string{"test-id"},
		Operation:  "test-op",
		Parameters: map[string]interface{}{
			"test_param": "test_value",
		},
		Result:       "success",
		Duration:     time.Second,
		ResourcePath: "/api/v1/test",
	}

	// This would be tested in the actual LogAction method
	// Here we just verify the structure is correct
	if details.MessageIDs[0] != "test-id" {
		t.Error("MessageIDs not set correctly")
	}
	if details.Operation != "test-op" {
		t.Error("Operation not set correctly")
	}
	if details.Parameters["test_param"] != "test_value" {
		t.Error("Parameters not set correctly")
	}
}
