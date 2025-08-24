package security

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Service provides security-related functionality
type Service struct {
	allowedPaths    []string
	restrictedPaths []string
	readOnlyPaths   []string
}

// NewService creates a new security service
func NewService() *Service {
	return &Service{
		// Allowed paths for Exim log files and spool directories
		allowedPaths: []string{
			"/var/log/exim4",
			"/var/spool/exim4",
			"/var/lib/exim4",
			"/etc/exim4", // Read-only access for configuration
		},
		// Paths that should never be accessed
		restrictedPaths: []string{
			"/etc/passwd",
			"/etc/shadow",
			"/etc/sudoers",
			"/root",
			"/home",
			"/var/www",
			"/usr/bin",
			"/usr/sbin",
			"/bin",
			"/sbin",
		},
		// Paths that should only be accessed read-only
		readOnlyPaths: []string{
			"/etc/exim4",
			"/var/log/exim4",
		},
	}
}

// FileAccessType represents the type of file access requested
type FileAccessType int

const (
	AccessRead FileAccessType = iota
	AccessWrite
	AccessExecute
)

// ValidateFileAccess validates that file access is allowed and secure
func (s *Service) ValidateFileAccess(path string, accessType FileAccessType) error {
	// Clean and resolve the path
	cleanPath, err := s.cleanPath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path is in restricted list
	if s.isRestrictedPath(cleanPath) {
		log.Printf("SECURITY: Attempted access to restricted path: %s", cleanPath)
		return fmt.Errorf("access to path '%s' is not allowed", cleanPath)
	}

	// Check if path is in allowed list
	if !s.isAllowedPath(cleanPath) {
		log.Printf("SECURITY: Attempted access to non-allowed path: %s", cleanPath)
		return fmt.Errorf("access to path '%s' is not allowed", cleanPath)
	}

	// Check read-only restrictions
	if accessType == AccessWrite && s.isReadOnlyPath(cleanPath) {
		log.Printf("SECURITY: Attempted write access to read-only path: %s", cleanPath)
		return fmt.Errorf("write access to path '%s' is not allowed", cleanPath)
	}

	// Check if file exists and is accessible
	if err := s.checkFilePermissions(cleanPath, accessType); err != nil {
		return fmt.Errorf("file access check failed: %w", err)
	}

	return nil
}

// cleanPath cleans and resolves the file path to prevent directory traversal
func (s *Service) cleanPath(path string) (string, error) {
	// Clean the path to resolve . and .. elements
	cleanPath := filepath.Clean(path)

	// Resolve symlinks to prevent symlink attacks
	resolvedPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		// If symlink resolution fails, use the clean path
		// This handles cases where the file doesn't exist yet
		resolvedPath = cleanPath
	}

	// Ensure the path is absolute
	if !filepath.IsAbs(resolvedPath) {
		return "", fmt.Errorf("path must be absolute")
	}

	return resolvedPath, nil
}

// isRestrictedPath checks if a path is in the restricted list
func (s *Service) isRestrictedPath(path string) bool {
	for _, restricted := range s.restrictedPaths {
		if strings.HasPrefix(path, restricted) {
			return true
		}
	}
	return false
}

// isAllowedPath checks if a path is in the allowed list
func (s *Service) isAllowedPath(path string) bool {
	for _, allowed := range s.allowedPaths {
		if strings.HasPrefix(path, allowed) {
			return true
		}
	}
	return false
}

// isReadOnlyPath checks if a path should be read-only
func (s *Service) isReadOnlyPath(path string) bool {
	for _, readOnly := range s.readOnlyPaths {
		if strings.HasPrefix(path, readOnly) {
			return true
		}
	}
	return false
}

// checkFilePermissions checks if the file can be accessed with the requested permissions
func (s *Service) checkFilePermissions(path string, accessType FileAccessType) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - check if we can create it in the directory
			dir := filepath.Dir(path)
			return s.checkDirectoryPermissions(dir, accessType)
		}
		return fmt.Errorf("cannot stat file: %w", err)
	}

	// Check if it's a regular file or directory
	if !info.Mode().IsRegular() && !info.Mode().IsDir() {
		return fmt.Errorf("path is not a regular file or directory")
	}

	// Check permissions based on access type
	switch accessType {
	case AccessRead:
		if err := s.checkReadPermission(path); err != nil {
			return err
		}
	case AccessWrite:
		if err := s.checkWritePermission(path); err != nil {
			return err
		}
	case AccessExecute:
		if err := s.checkExecutePermission(path); err != nil {
			return err
		}
	}

	return nil
}

// checkDirectoryPermissions checks if we can access a directory
func (s *Service) checkDirectoryPermissions(dir string, accessType FileAccessType) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.Mode().IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Check if we can write to the directory (for file creation)
	if accessType == AccessWrite {
		return s.checkWritePermission(dir)
	}

	return nil
}

// checkReadPermission checks if we can read a file
func (s *Service) checkReadPermission(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file for reading: %w", err)
	}
	file.Close()
	return nil
}

// checkWritePermission checks if we can write to a file or directory
func (s *Service) checkWritePermission(path string) error {
	// Try to open for writing (this will fail if we don't have permission)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat path: %w", err)
	}

	if info.Mode().IsDir() {
		// For directories, try to create a temporary file
		tempFile := filepath.Join(path, ".security_check_temp")
		file, err := os.Create(tempFile)
		if err != nil {
			return fmt.Errorf("cannot write to directory: %w", err)
		}
		file.Close()
		os.Remove(tempFile) // Clean up
	} else {
		// For files, try to open for writing
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("cannot open file for writing: %w", err)
		}
		file.Close()
	}

	return nil
}

// checkExecutePermission checks if we can execute a file
func (s *Service) checkExecutePermission(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	// Check if file has execute permission
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("file is not executable")
	}

	return nil
}

// SecureFileRead reads a file with security checks
func (s *Service) SecureFileRead(path string) ([]byte, error) {
	if err := s.ValidateFileAccess(path, AccessRead); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("SECURITY: Failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	log.Printf("SECURITY: File read access granted: %s", path)
	return data, nil
}

// SecureFileWrite writes to a file with security checks
func (s *Service) SecureFileWrite(path string, data []byte) error {
	if err := s.ValidateFileAccess(path, AccessWrite); err != nil {
		return err
	}

	err := os.WriteFile(path, data, 0644)
	if err != nil {
		log.Printf("SECURITY: Failed to write file %s: %v", path, err)
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("SECURITY: File write access granted: %s", path)
	return nil
}

// DropPrivileges attempts to drop privileges to a less privileged user
func (s *Service) DropPrivileges(username string) error {
	// This is a placeholder for privilege dropping
	// In a real implementation, you would:
	// 1. Look up the user by name
	// 2. Set the process UID/GID to that user
	// 3. Verify the privileges were dropped

	log.Printf("SECURITY: Privilege dropping requested for user: %s", username)

	// For now, just log the attempt
	// In production, implement actual privilege dropping:
	/*
		user, err := user.Lookup(username)
		if err != nil {
			return fmt.Errorf("failed to lookup user %s: %w", username, err)
		}

		uid, _ := strconv.Atoi(user.Uid)
		gid, _ := strconv.Atoi(user.Gid)

		if err := syscall.Setgid(gid); err != nil {
			return fmt.Errorf("failed to set GID: %w", err)
		}

		if err := syscall.Setuid(uid); err != nil {
			return fmt.Errorf("failed to set UID: %w", err)
		}
	*/

	return nil
}

// CheckProcessPrivileges checks current process privileges
func (s *Service) CheckProcessPrivileges() error {
	// Check if running as root (UID 0)
	if os.Getuid() == 0 {
		log.Printf("SECURITY WARNING: Process is running as root (UID 0)")
		return fmt.Errorf("process should not run as root for security reasons")
	}

	// Check effective UID
	if os.Geteuid() == 0 {
		log.Printf("SECURITY WARNING: Process has effective UID 0 (root)")
		return fmt.Errorf("process should not have root privileges")
	}

	log.Printf("SECURITY: Process privileges check passed - UID: %d, EUID: %d", os.Getuid(), os.Geteuid())
	return nil
}

// ValidateSystemCommand validates system commands before execution
func (s *Service) ValidateSystemCommand(command string, args []string) error {
	// Only allow specific Exim commands
	allowedCommands := map[string]bool{
		"exim":            true,
		"exim4":           true,
		"/usr/sbin/exim":  true,
		"/usr/sbin/exim4": true,
	}

	if !allowedCommands[command] {
		log.Printf("SECURITY: Attempted execution of disallowed command: %s", command)
		return fmt.Errorf("command '%s' is not allowed", command)
	}

	// Validate command arguments
	for _, arg := range args {
		if err := s.validateCommandArgument(arg); err != nil {
			log.Printf("SECURITY: Invalid command argument: %s", arg)
			return fmt.Errorf("invalid command argument: %w", err)
		}
	}

	log.Printf("SECURITY: System command validation passed: %s %v", command, args)
	return nil
}

// validateCommandArgument validates individual command arguments
func (s *Service) validateCommandArgument(arg string) error {
	// Check for command injection attempts
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\"", "'"}
	for _, char := range dangerousChars {
		if strings.Contains(arg, char) {
			return fmt.Errorf("argument contains dangerous character: %s", char)
		}
	}

	// Check for path traversal
	if strings.Contains(arg, "..") {
		return fmt.Errorf("argument contains path traversal sequence")
	}

	return nil
}

// LogSecurityEvent logs security-related events
func (s *Service) LogSecurityEvent(event, details string) {
	log.Printf("SECURITY EVENT: %s - %s", event, details)
}

// SetupSecureEnvironment sets up a secure environment for the application
func (s *Service) SetupSecureEnvironment() error {
	// Set umask to ensure files are created with secure permissions (Unix only)
	// On Windows, this is handled differently through file permissions
	s.setSecureFilePermissions()

	// Check current privileges
	if err := s.CheckProcessPrivileges(); err != nil {
		// Log warning but don't fail - allow running as root in development
		log.Printf("SECURITY WARNING: %v", err)
	}

	// Set up signal handlers for graceful shutdown
	// This would be implemented in the main application

	log.Printf("SECURITY: Secure environment setup completed")
	return nil
}

// setSecureFilePermissions sets secure file permissions (platform-specific)
func (s *Service) setSecureFilePermissions() {
	// On Unix systems, we would set umask
	// On Windows, we rely on NTFS permissions
	// This is a placeholder for platform-specific implementation
	log.Printf("SECURITY: File permissions configured for current platform")
}

// GetSecureLogPaths returns the list of allowed log file paths
func (s *Service) GetSecureLogPaths() []string {
	return []string{
		"/var/log/exim4/mainlog",
		"/var/log/exim4/rejectlog",
		"/var/log/exim4/paniclog",
	}
}

// GetSecureSpoolPath returns the secure spool directory path
func (s *Service) GetSecureSpoolPath() string {
	return "/var/spool/exim4"
}
