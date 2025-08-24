package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations
type Service struct {
	userRepo    *database.UserRepository
	sessionRepo *database.SessionRepository
	auditRepo   *database.AuditLogRepository
}

// NewService creates a new authentication service
func NewService(db *database.DB) *Service {
	return &Service{
		userRepo:    database.NewUserRepository(db),
		sessionRepo: database.NewSessionRepository(db),
		auditRepo:   database.NewAuditLogRepository(db),
	}
}

// Login authenticates a user and creates a session
func (s *Service) Login(username, password, ipAddress, userAgent string) (*database.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		// Log failed login attempt
		s.auditRepo.Create(&database.AuditLog{
			Action:    "login_failed",
			UserID:    &username, // Use username since we don't have user ID
			Details:   stringPtr(`{"reason": "user_not_found"}`),
			IPAddress: &ipAddress,
		})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Log failed login attempt
		userIDStr := fmt.Sprintf("%d", user.ID)
		s.auditRepo.Create(&database.AuditLog{
			Action:    "login_failed",
			UserID:    &userIDStr,
			Details:   stringPtr(`{"reason": "invalid_password"}`),
			IPAddress: &ipAddress,
		})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create session
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session
	session := &database.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
		// Log but don't fail - session is already created
		fmt.Printf("Warning: failed to update last login for user %d: %v\n", user.ID, err)
	}

	// Log successful login
	userIDStr := fmt.Sprintf("%d", user.ID)
	s.auditRepo.Create(&database.AuditLog{
		Action:    "login_success",
		UserID:    &userIDStr,
		Details:   stringPtr(fmt.Sprintf(`{"session_id": "%s"}`, sessionID)),
		IPAddress: &ipAddress,
	})

	return &database.LoginResponse{
		User:      *user,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}, nil
}

// Logout invalidates a session
func (s *Service) Logout(sessionID, ipAddress string) error {
	// Get session to get user ID for audit log
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	// Delete session
	if err := s.sessionRepo.Delete(sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Log logout
	userIDStr := fmt.Sprintf("%d", session.UserID)
	s.auditRepo.Create(&database.AuditLog{
		Action:    "logout",
		UserID:    &userIDStr,
		Details:   stringPtr(fmt.Sprintf(`{"session_id": "%s"}`, sessionID)),
		IPAddress: &ipAddress,
	})

	return nil
}

// ValidateSession validates a session and returns the user
func (s *Service) ValidateSession(sessionID string) (*database.User, error) {
	// Get session
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	// Get user
	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// CreateUser creates a new user (for initial setup)
func (s *Service) CreateUser(username, password, email, fullName string) (*database.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &database.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        &email,
		FullName:     &fullName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log user creation
	userIDStr := fmt.Sprintf("%d", user.ID)
	s.auditRepo.Create(&database.AuditLog{
		Action:  "user_created",
		UserID:  &userIDStr,
		Details: stringPtr(fmt.Sprintf(`{"username": "%s", "email": "%s"}`, username, email)),
	})

	return user, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *Service) CleanupExpiredSessions() error {
	deleted, err := s.sessionRepo.DeleteExpired()
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	if deleted > 0 {
		fmt.Printf("Cleaned up %d expired sessions\n", deleted)
	}

	return nil
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
