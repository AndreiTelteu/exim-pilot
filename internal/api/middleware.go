package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/audit"
	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/validation"
)

// loggingMiddleware logs all HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log the request
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %v %s %s",
			r.Method,
			r.RequestURI,
			lrw.statusCode,
			duration,
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

// errorHandlingMiddleware provides centralized error handling
func (s *Server) errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic in handler: %v", err)

				response := APIResponse{
					Success: false,
					Error:   "Internal server error",
				}

				WriteJSONResponse(w, http.StatusInternalServerError, response)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates authentication for protected routes
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for login and health endpoints
		if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Get session ID from cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			response := APIResponse{
				Success: false,
				Error:   "Authentication required",
			}
			WriteJSONResponse(w, http.StatusUnauthorized, response)
			return
		}

		// Validate session
		user, err := s.authService.ValidateSession(cookie.Value)
		if err != nil {
			response := APIResponse{
				Success: false,
				Error:   "Invalid session",
			}
			WriteJSONResponse(w, http.StatusUnauthorized, response)
			return
		}

		// Add user to request context
		ctx := r.Context()
		ctx = SetUserInContext(ctx, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write ensures WriteHeader is called if it hasn't been already
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		lrw.WriteHeader(http.StatusOK)
	}
	return lrw.ResponseWriter.Write(b)
}

// Context utilities for user authentication
type contextKey string

const userContextKey contextKey = "user"

// SetUserInContext adds a user to the request context
func SetUserInContext(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*database.User, bool) {
	user, ok := ctx.Value(userContextKey).(*database.User)
	return user, ok
}

// contentTypeMiddleware ensures proper content-type headers are set
func (s *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set default content-type for API responses, not static files
		if strings.HasPrefix(r.URL.Path, "/api/") && w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// validationMiddleware provides input validation for API requests
func (s *Server) validationMiddleware(next http.Handler) http.Handler {
	validator := validation.NewService()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate common request parameters
		if err := s.validateCommonParams(r, validator); err != nil {
			if validationErrors, ok := err.(*validation.ValidationErrors); ok {
				response := APIResponse{
					Success: false,
					Error:   "Validation failed",
					Data:    validationErrors.Errors,
				}
				WriteJSONResponse(w, http.StatusBadRequest, response)
				return
			}

			response := APIResponse{
				Success: false,
				Error:   err.Error(),
			}
			WriteJSONResponse(w, http.StatusBadRequest, response)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// auditMiddleware logs all administrative actions
func (s *Server) auditMiddleware(next http.Handler) http.Handler {
	auditService := audit.NewService(s.repository)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip audit logging for read-only operations
		if r.Method == "GET" && !s.isAuditableEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get user context
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			// If no user context, still proceed but log as anonymous
			user = &database.User{ID: 0}
		}

		// Create audit context
		auditCtx := &audit.AuditContext{
			UserID:    getUserIDString(user),
			IPAddress: getClientIPFromRequest(r),
			UserAgent: r.UserAgent(),
			RequestID: generateRequestID(),
		}

		// Determine action type based on request
		action := s.determineAuditAction(r)

		// Log the action (before processing to capture attempts)
		if action != "" {
			details := &audit.AuditDetails{
				ResourcePath: r.URL.Path,
				Parameters: map[string]interface{}{
					"method": r.Method,
					"query":  r.URL.RawQuery,
				},
			}

			// Extract message ID if present in path
			var messageID *string
			if mid := extractMessageIDFromPath(r.URL.Path); mid != "" {
				messageID = &mid
			}

			err := auditService.LogAction(r.Context(), action, messageID, auditCtx, details)
			if err != nil {
				log.Printf("Failed to log audit action: %v", err)
			}
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// securityHeadersMiddleware adds security headers to responses
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Remove server information
		w.Header().Set("Server", "Exim-Pilot")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Helper functions for middleware

func (s *Server) validateCommonParams(r *http.Request, validator *validation.Service) error {
	// Validate pagination parameters if present
	if page := r.URL.Query().Get("page"); page != "" {
		if perPage := r.URL.Query().Get("per_page"); perPage != "" {
			pageInt, err := getIntParam(page, 1)
			if err != nil {
				return &validation.ValidationError{Field: "page", Message: "invalid page parameter"}
			}
			perPageInt, err := getIntParam(perPage, 50)
			if err != nil {
				return &validation.ValidationError{Field: "per_page", Message: "invalid per_page parameter"}
			}

			if err := validator.ValidatePagination(pageInt, perPageInt); err != nil {
				return err
			}
		}
	}

	// Validate message ID in path if present
	if messageID := extractMessageIDFromPath(r.URL.Path); messageID != "" {
		if err := validator.ValidateMessageID(messageID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) isAuditableEndpoint(path string) bool {
	// Define endpoints that should be audited even for GET requests
	auditableGETEndpoints := []string{
		"/api/v1/messages/",
		"/api/v1/queue/",
	}

	for _, endpoint := range auditableGETEndpoints {
		if contains(path, endpoint) {
			return true
		}
	}

	return false
}

func (s *Server) determineAuditAction(r *http.Request) audit.ActionType {
	path := r.URL.Path
	method := r.Method

	// Queue operations
	if contains(path, "/api/v1/queue/") {
		if contains(path, "/deliver") && method == "POST" {
			return audit.ActionQueueDeliver
		}
		if contains(path, "/freeze") && method == "POST" {
			return audit.ActionQueueFreeze
		}
		if contains(path, "/thaw") && method == "POST" {
			return audit.ActionQueueThaw
		}
		if method == "DELETE" {
			return audit.ActionQueueDelete
		}
		if contains(path, "/bulk") && method == "POST" {
			return audit.ActionBulkDeliver // Will be refined based on operation
		}
	}

	// Authentication operations
	if contains(path, "/api/v1/auth/login") && method == "POST" {
		return audit.ActionLogin
	}
	if contains(path, "/api/v1/auth/logout") && method == "POST" {
		return audit.ActionLogout
	}

	// Message operations
	if contains(path, "/api/v1/messages/") {
		if contains(path, "/content") && method == "GET" {
			return audit.ActionMessageContent
		}
		if contains(path, "/notes") && method == "POST" {
			return audit.ActionNoteCreate
		}
		if contains(path, "/notes") && method == "PUT" {
			return audit.ActionNoteUpdate
		}
		if contains(path, "/notes") && method == "DELETE" {
			return audit.ActionNoteDelete
		}
		if contains(path, "/tags") && method == "POST" {
			return audit.ActionTagCreate
		}
		if contains(path, "/tags") && method == "DELETE" {
			return audit.ActionTagDelete
		}
		if method == "GET" {
			return audit.ActionMessageView
		}
	}

	return ""
}

// Helper utility functions

func getUserIDString(user *database.User) string {
	if user == nil {
		return "anonymous"
	}
	return fmt.Sprintf("%d", user.ID)
}

func getClientIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func generateRequestID() string {
	// Simple request ID generation - in production use UUID
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func extractMessageIDFromPath(path string) string {
	// Extract message ID from paths like /api/v1/queue/{id} or /api/v1/messages/{id}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if (part == "queue" || part == "messages") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func getIntParam(param string, defaultValue int) (int, error) {
	if param == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue, err
	}

	return value, nil
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
