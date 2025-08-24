package validation

import (
	"testing"
)

func TestValidateMessageID(t *testing.T) {
	service := NewService()

	// Test valid message IDs
	validIDs := []string{
		"1ABC23-DEF456-GH",
		"123456-789ABC-DE",
		"ABCDEF-123456-78",
	}

	for _, id := range validIDs {
		err := service.ValidateMessageID(id)
		if err != nil {
			t.Errorf("Expected no error for valid message ID '%s', got: %v", id, err)
		}
	}

	// Test invalid message IDs
	invalidIDs := []string{
		"",
		"123",
		"123456-789ABC",
		"123456-789ABC-DEF",
		"123456-789ABC-D",
		"123456_789ABC-DE",
		"123456-789ABC_DE",
	}

	for _, id := range invalidIDs {
		err := service.ValidateMessageID(id)
		if err == nil {
			t.Errorf("Expected error for invalid message ID '%s', got nil", id)
		}
	}
}

func TestValidateEmailAddress(t *testing.T) {
	service := NewService()

	// Test valid email addresses
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"admin+tag@company.org",
	}

	for _, email := range validEmails {
		err := service.ValidateEmailAddress(email)
		if err != nil {
			t.Errorf("Expected no error for valid email '%s', got: %v", email, err)
		}
	}

	// Test invalid email addresses
	invalidEmails := []string{
		"",
		"invalid",
		"@domain.com",
		"user@",
		"user space@domain.com",
	}

	for _, email := range invalidEmails {
		err := service.ValidateEmailAddress(email)
		if err == nil {
			t.Errorf("Expected error for invalid email '%s', got nil", email)
		}
	}
}

func TestValidateOperation(t *testing.T) {
	service := NewService()

	// Test valid operations
	validOps := []string{"deliver", "freeze", "thaw", "delete"}
	for _, op := range validOps {
		err := service.ValidateOperation(op)
		if err != nil {
			t.Errorf("Expected no error for valid operation '%s', got: %v", op, err)
		}
	}

	// Test invalid operations
	invalidOps := []string{"", "invalid", "DROP TABLE", "rm -rf"}
	for _, op := range invalidOps {
		err := service.ValidateOperation(op)
		if err == nil {
			t.Errorf("Expected error for invalid operation '%s', got nil", op)
		}
	}
}

func TestValidateBulkRequest(t *testing.T) {
	service := NewService()

	// Test valid bulk request
	validMessageIDs := []string{"1ABC23-DEF456-GH", "123456-789ABC-DE"}
	err := service.ValidateBulkRequest("deliver", validMessageIDs)
	if err != nil {
		t.Errorf("Expected no error for valid bulk request, got: %v", err)
	}

	// Test invalid operation
	err = service.ValidateBulkRequest("invalid", validMessageIDs)
	if err == nil {
		t.Error("Expected error for invalid operation, got nil")
	}

	// Test empty message IDs
	err = service.ValidateBulkRequest("deliver", []string{})
	if err == nil {
		t.Error("Expected error for empty message IDs, got nil")
	}

	// Test invalid message ID in array
	invalidMessageIDs := []string{"1ABC23-DEF456-GH", "invalid-id"}
	err = service.ValidateBulkRequest("deliver", invalidMessageIDs)
	if err == nil {
		t.Error("Expected error for invalid message ID in array, got nil")
	}
}

func TestContainsSuspiciousContent(t *testing.T) {
	service := NewService()

	// Test safe content
	safeContent := []string{
		"This is a normal message",
		"User reported delivery issue",
		"Message contains special chars: @#$%^&*()",
	}

	for _, content := range safeContent {
		if service.containsSuspiciousContent(content) {
			t.Errorf("Expected safe content '%s' to be allowed", content)
		}
	}

	// Test suspicious content
	suspiciousContent := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"onload=alert('xss')",
		"eval(malicious_code)",
		"document.cookie",
	}

	for _, content := range suspiciousContent {
		if !service.containsSuspiciousContent(content) {
			t.Errorf("Expected suspicious content '%s' to be blocked", content)
		}
	}
}
