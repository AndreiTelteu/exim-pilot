package security

import (
	"testing"
)

func TestValidateFileAccess(t *testing.T) {
	service := NewService()

	// Test restricted path access
	err := service.ValidateFileAccess("/etc/passwd", AccessRead)
	if err == nil {
		t.Error("Expected error for restricted path access, got nil")
	}

	// Test non-allowed path access
	err = service.ValidateFileAccess("/tmp/test", AccessRead)
	if err == nil {
		t.Error("Expected error for non-allowed path access, got nil")
	}

	// Test allowed path (would need actual file system setup for full test)
	// This is a basic validation test
}

func TestValidateSystemCommand(t *testing.T) {
	service := NewService()

	// Test allowed command
	err := service.ValidateSystemCommand("exim", []string{"-M", "test-message-id"})
	if err != nil {
		t.Errorf("Expected no error for allowed command, got: %v", err)
	}

	// Test disallowed command
	err = service.ValidateSystemCommand("rm", []string{"-rf", "/"})
	if err == nil {
		t.Error("Expected error for disallowed command, got nil")
	}

	// Test command injection attempt
	err = service.ValidateSystemCommand("exim", []string{"-M", "test; rm -rf /"})
	if err == nil {
		t.Error("Expected error for command injection attempt, got nil")
	}
}

func TestValidateCommandArgument(t *testing.T) {
	service := NewService()

	// Test valid argument
	err := service.validateCommandArgument("1ABC23-DEF456-GH")
	if err != nil {
		t.Errorf("Expected no error for valid argument, got: %v", err)
	}

	// Test dangerous characters
	dangerousArgs := []string{
		"test;rm",
		"test&rm",
		"test|rm",
		"test`rm`",
		"test$(rm)",
		"test../../../etc/passwd",
	}

	for _, arg := range dangerousArgs {
		err := service.validateCommandArgument(arg)
		if err == nil {
			t.Errorf("Expected error for dangerous argument '%s', got nil", arg)
		}
	}
}
