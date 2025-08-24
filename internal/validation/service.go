package validation

import (
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Service provides input validation functionality
type Service struct {
	// Configuration for validation rules
	maxStringLength   int
	maxArrayLength    int
	allowedOperations map[string]bool
	allowedLogTypes   map[string]bool
	allowedStatuses   map[string]bool
}

// NewService creates a new validation service
func NewService() *Service {
	return &Service{
		maxStringLength: 10000,
		maxArrayLength:  1000,
		allowedOperations: map[string]bool{
			"deliver": true,
			"freeze":  true,
			"thaw":    true,
			"delete":  true,
		},
		allowedLogTypes: map[string]bool{
			"main":   true,
			"reject": true,
			"panic":  true,
		},
		allowedStatuses: map[string]bool{
			"received":  true,
			"queued":    true,
			"delivered": true,
			"deferred":  true,
			"bounced":   true,
			"frozen":    true,
		},
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation errors occurred"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(e.Errors))
}

func (e *ValidationErrors) Add(field, message, value string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ValidateMessageID validates an Exim message ID
func (s *Service) ValidateMessageID(messageID string) error {
	if messageID == "" {
		return &ValidationError{Field: "message_id", Message: "message ID is required"}
	}

	// Exim message ID format: XXXXXX-YYYYYY-ZZ
	// Where X, Y are base-62 characters and Z is base-62
	messageIDPattern := regexp.MustCompile(`^[0-9A-Za-z]{6}-[0-9A-Za-z]{6}-[0-9A-Za-z]{2}$`)
	if !messageIDPattern.MatchString(messageID) {
		return &ValidationError{
			Field:   "message_id",
			Message: "invalid message ID format (expected: XXXXXX-YYYYYY-ZZ)",
			Value:   messageID,
		}
	}

	return nil
}

// ValidateEmailAddress validates an email address
func (s *Service) ValidateEmailAddress(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email address is required"}
	}

	if len(email) > 320 { // RFC 5321 limit
		return &ValidationError{
			Field:   "email",
			Message: "email address too long (max 320 characters)",
			Value:   email,
		}
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return &ValidationError{
			Field:   "email",
			Message: "invalid email address format",
			Value:   email,
		}
	}

	return nil
}

// ValidateIPAddress validates an IP address (IPv4 or IPv6)
func (s *Service) ValidateIPAddress(ip string) error {
	if ip == "" {
		return nil // IP address is optional in most contexts
	}

	if net.ParseIP(ip) == nil {
		return &ValidationError{
			Field:   "ip_address",
			Message: "invalid IP address format",
			Value:   ip,
		}
	}

	return nil
}

// ValidateOperation validates queue operation types
func (s *Service) ValidateOperation(operation string) error {
	if operation == "" {
		return &ValidationError{Field: "operation", Message: "operation is required"}
	}

	if !s.allowedOperations[operation] {
		return &ValidationError{
			Field:   "operation",
			Message: "invalid operation (allowed: deliver, freeze, thaw, delete)",
			Value:   operation,
		}
	}

	return nil
}

// ValidateLogType validates log type values
func (s *Service) ValidateLogType(logType string) error {
	if logType == "" {
		return nil // Log type is optional for filtering
	}

	if !s.allowedLogTypes[logType] {
		return &ValidationError{
			Field:   "log_type",
			Message: "invalid log type (allowed: main, reject, panic)",
			Value:   logType,
		}
	}

	return nil
}

// ValidateStatus validates message status values
func (s *Service) ValidateStatus(status string) error {
	if status == "" {
		return nil // Status is optional for filtering
	}

	if !s.allowedStatuses[status] {
		return &ValidationError{
			Field:   "status",
			Message: "invalid status (allowed: received, queued, delivered, deferred, bounced, frozen)",
			Value:   status,
		}
	}

	return nil
}

// ValidateString validates string fields with length limits
func (s *Service) ValidateString(field, value string, required bool, maxLength int) error {
	if required && value == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if maxLength == 0 {
		maxLength = s.maxStringLength
	}

	if len(value) > maxLength {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s too long (max %d characters)", field, maxLength),
			Value:   value,
		}
	}

	return nil
}

// ValidateStringArray validates string arrays with length limits
func (s *Service) ValidateStringArray(field string, values []string, required bool, maxItems int) error {
	if required && len(values) == 0 {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if maxItems == 0 {
		maxItems = s.maxArrayLength
	}

	if len(values) > maxItems {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s has too many items (max %d)", field, maxItems),
		}
	}

	// Validate each item in the array
	for i, value := range values {
		if err := s.ValidateString(fmt.Sprintf("%s[%d]", field, i), value, true, 0); err != nil {
			return err
		}
	}

	return nil
}

// ValidateInteger validates integer fields with range limits
func (s *Service) ValidateInteger(field string, value int, required bool, min, max int) error {
	if required && value == 0 {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if min != 0 && value < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be at least %d", field, min),
			Value:   strconv.Itoa(value),
		}
	}

	if max != 0 && value > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be at most %d", field, max),
			Value:   strconv.Itoa(value),
		}
	}

	return nil
}

// ValidateTimeRange validates time range parameters
func (s *Service) ValidateTimeRange(startTime, endTime *time.Time) error {
	if startTime != nil && endTime != nil {
		if startTime.After(*endTime) {
			return &ValidationError{
				Field:   "time_range",
				Message: "start time must be before end time",
			}
		}

		// Prevent excessively large time ranges (more than 1 year)
		if endTime.Sub(*startTime) > 365*24*time.Hour {
			return &ValidationError{
				Field:   "time_range",
				Message: "time range too large (max 1 year)",
			}
		}
	}

	return nil
}

// ValidatePagination validates pagination parameters
func (s *Service) ValidatePagination(page, perPage int) error {
	errors := &ValidationErrors{}

	if err := s.ValidateInteger("page", page, false, 1, 10000); err != nil {
		errors.Add(err.(*ValidationError).Field, err.(*ValidationError).Message, err.(*ValidationError).Value)
	}

	if err := s.ValidateInteger("per_page", perPage, false, 1, 1000); err != nil {
		errors.Add(err.(*ValidationError).Field, err.(*ValidationError).Message, err.(*ValidationError).Value)
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateSearchCriteria validates queue search criteria
func (s *Service) ValidateSearchCriteria(criteria map[string]interface{}) error {
	errors := &ValidationErrors{}

	// Validate sender if provided
	if sender, ok := criteria["sender"].(string); ok && sender != "" {
		if err := s.ValidateEmailAddress(sender); err != nil {
			errors.Add("sender", err.(*ValidationError).Message, sender)
		}
	}

	// Validate recipient if provided
	if recipient, ok := criteria["recipient"].(string); ok && recipient != "" {
		if err := s.ValidateEmailAddress(recipient); err != nil {
			errors.Add("recipient", err.(*ValidationError).Message, recipient)
		}
	}

	// Validate message ID if provided
	if messageID, ok := criteria["message_id"].(string); ok && messageID != "" {
		if err := s.ValidateMessageID(messageID); err != nil {
			errors.Add("message_id", err.(*ValidationError).Message, messageID)
		}
	}

	// Validate status if provided
	if status, ok := criteria["status"].(string); ok && status != "" {
		if err := s.ValidateStatus(status); err != nil {
			errors.Add("status", err.(*ValidationError).Message, status)
		}
	}

	// Validate age range if provided
	if minAge, ok := criteria["min_age"].(int); ok {
		if err := s.ValidateInteger("min_age", minAge, false, 0, 365*24*3600); err != nil {
			errors.Add("min_age", err.(*ValidationError).Message, strconv.Itoa(minAge))
		}
	}

	if maxAge, ok := criteria["max_age"].(int); ok {
		if err := s.ValidateInteger("max_age", maxAge, false, 0, 365*24*3600); err != nil {
			errors.Add("max_age", err.(*ValidationError).Message, strconv.Itoa(maxAge))
		}
	}

	// Validate size range if provided
	if minSize, ok := criteria["min_size"].(int); ok {
		if err := s.ValidateInteger("min_size", minSize, false, 0, 100*1024*1024); err != nil {
			errors.Add("min_size", err.(*ValidationError).Message, strconv.Itoa(minSize))
		}
	}

	if maxSize, ok := criteria["max_size"].(int); ok {
		if err := s.ValidateInteger("max_size", maxSize, false, 0, 100*1024*1024); err != nil {
			errors.Add("max_size", err.(*ValidationError).Message, strconv.Itoa(maxSize))
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateBulkRequest validates bulk operation requests
func (s *Service) ValidateBulkRequest(operation string, messageIDs []string) error {
	errors := &ValidationErrors{}

	// Validate operation
	if err := s.ValidateOperation(operation); err != nil {
		errors.Add(err.(*ValidationError).Field, err.(*ValidationError).Message, err.(*ValidationError).Value)
	}

	// Validate message IDs array
	if err := s.ValidateStringArray("message_ids", messageIDs, true, 100); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			errors.Add(validationErr.Field, validationErr.Message, validationErr.Value)
		}
	} else {
		// Validate each message ID
		for i, messageID := range messageIDs {
			if err := s.ValidateMessageID(messageID); err != nil {
				errors.Add(fmt.Sprintf("message_ids[%d]", i), err.(*ValidationError).Message, messageID)
			}
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateUserInput validates user input for notes and tags
func (s *Service) ValidateUserInput(input map[string]interface{}) error {
	errors := &ValidationErrors{}

	// Validate note content if provided
	if note, ok := input["note"].(string); ok {
		if err := s.ValidateString("note", note, true, 5000); err != nil {
			errors.Add(err.(*ValidationError).Field, err.(*ValidationError).Message, err.(*ValidationError).Value)
		}

		// Check for potentially malicious content
		if s.containsSuspiciousContent(note) {
			errors.Add("note", "note contains potentially malicious content", "")
		}
	}

	// Validate tag if provided
	if tag, ok := input["tag"].(string); ok {
		if err := s.ValidateString("tag", tag, true, 50); err != nil {
			errors.Add(err.(*ValidationError).Field, err.(*ValidationError).Message, err.(*ValidationError).Value)
		}

		// Tags should only contain alphanumeric characters, hyphens, and underscores
		tagPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
		if !tagPattern.MatchString(tag) {
			errors.Add("tag", "tag can only contain letters, numbers, hyphens, and underscores", tag)
		}
	}

	// Validate color if provided
	if color, ok := input["color"].(string); ok && color != "" {
		colorPattern := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
		if !colorPattern.MatchString(color) {
			errors.Add("color", "color must be a valid hex color code (e.g., #FF0000)", color)
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// containsSuspiciousContent checks for potentially malicious content
func (s *Service) containsSuspiciousContent(content string) bool {
	// Check for common script injection patterns
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"onclick=",
		"eval(",
		"document.cookie",
		"window.location",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}

	return false
}

// SanitizeString removes potentially dangerous characters from strings
func (s *Service) SanitizeString(input string) string {
	// Remove null bytes and control characters
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove other control characters except newlines and tabs
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ValidateFilePath validates file paths to prevent directory traversal
func (s *Service) ValidateFilePath(path string) error {
	if path == "" {
		return &ValidationError{Field: "file_path", Message: "file path is required"}
	}

	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return &ValidationError{
			Field:   "file_path",
			Message: "file path contains invalid characters (..)",
			Value:   path,
		}
	}

	// Check for absolute paths (should be relative)
	if strings.HasPrefix(path, "/") || strings.Contains(path, ":") {
		return &ValidationError{
			Field:   "file_path",
			Message: "absolute file paths are not allowed",
			Value:   path,
		}
	}

	return nil
}
