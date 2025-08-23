package api

import (
	"log"
	"net/http"
	"time"
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

// contentTypeMiddleware ensures proper Content-Type for API requests
func (s *Server) contentTypeMiddleware(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "" {
		response := APIResponse{
			Success: false,
			Error:   "Content-Type must be application/json",
		}
		WriteJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Continue to the next handler
	s.router.ServeHTTP(w, r)
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
