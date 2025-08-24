package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
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
		// Set default content-type for API responses
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
