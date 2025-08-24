package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/auth"
	"github.com/andreitelteu/exim-pilot/internal/database"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService *auth.Service
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(authService *auth.Service) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// handleLogin handles user login
func (h *AuthHandlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq database.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		response := APIResponse{
			Success: false,
			Error:   "Invalid request body",
		}
		WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if loginReq.Username == "" || loginReq.Password == "" {
		response := APIResponse{
			Success: false,
			Error:   "Username and password are required",
		}
		WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Get client IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Attempt login
	loginResp, err := h.authService.Login(loginReq.Username, loginReq.Password, ipAddress, userAgent)
	if err != nil {
		response := APIResponse{
			Success: false,
			Error:   "Invalid credentials",
		}
		WriteJSONResponse(w, http.StatusUnauthorized, response)
		return
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    loginResp.SessionID,
		Expires:  loginResp.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Only secure if HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	// Return user info (without password hash)
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user":       loginResp.User,
			"expires_at": loginResp.ExpiresAt,
		},
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

// handleLogout handles user logout
func (h *AuthHandlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session ID from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		response := APIResponse{
			Success: false,
			Error:   "No active session",
		}
		WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Get client IP
	ipAddress := getClientIP(r)

	// Logout
	if err := h.authService.Logout(cookie.Value, ipAddress); err != nil {
		response := APIResponse{
			Success: false,
			Error:   "Failed to logout",
		}
		WriteJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	// Clear session cookie
	clearCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, clearCookie)

	response := APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "Logged out successfully"},
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

// handleMe returns current user information
func (h *AuthHandlers) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		response := APIResponse{
			Success: false,
			Error:   "User not found in context",
		}
		WriteJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    user,
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
