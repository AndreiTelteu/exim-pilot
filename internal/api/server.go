package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/queue"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	router       *mux.Router
	httpServer   *http.Server
	config       *Config
	queueService *queue.Service
	logService   *logprocessor.Service
	repository   *database.Repository
}

// NewServer creates a new API server instance
func NewServer(config *Config, queueService *queue.Service, logService *logprocessor.Service, repository *database.Repository) *Server {
	if config == nil {
		config = NewConfig()
		config.LoadFromEnv()
	}

	s := &Server{
		router:       mux.NewRouter(),
		config:       config,
		queueService: queueService,
		logService:   logService,
		repository:   repository,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures all middleware for the server
func (s *Server) setupMiddleware() {
	// CORS middleware for frontend integration
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins(s.config.AllowedOrigins),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"}),
		handlers.AllowCredentials(),
	)

	// Apply CORS to all routes
	s.router.Use(corsHandler)

	// Request logging middleware (if enabled)
	if s.config.LogRequests {
		s.router.Use(s.loggingMiddleware)
	}

	// Error handling middleware
	s.router.Use(s.errorHandlingMiddleware)

	// Content-Type middleware for API routes
	s.router.PathPrefix("/api/").Handler(
		http.HandlerFunc(s.contentTypeMiddleware),
	).Methods("POST", "PUT", "PATCH")
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Queue management routes (Task 5.2)
	if s.queueService != nil {
		queueHandlers := NewQueueHandlers(s.queueService)

		// Queue listing and search
		api.HandleFunc("/queue", queueHandlers.handleQueueList).Methods("GET")
		api.HandleFunc("/queue/search", queueHandlers.handleQueueSearch).Methods("POST")
		api.HandleFunc("/queue/health", queueHandlers.handleQueueHealth).Methods("GET")
		api.HandleFunc("/queue/statistics", queueHandlers.handleQueueStatistics).Methods("GET")

		// Individual message operations
		api.HandleFunc("/queue/{id}", queueHandlers.handleQueueDetails).Methods("GET")
		api.HandleFunc("/queue/{id}/deliver", queueHandlers.handleQueueDeliver).Methods("POST")
		api.HandleFunc("/queue/{id}/freeze", queueHandlers.handleQueueFreeze).Methods("POST")
		api.HandleFunc("/queue/{id}/thaw", queueHandlers.handleQueueThaw).Methods("POST")
		api.HandleFunc("/queue/{id}", queueHandlers.handleQueueDelete).Methods("DELETE")
		api.HandleFunc("/queue/{id}/history", queueHandlers.handleQueueHistory).Methods("GET")

		// Bulk operations
		api.HandleFunc("/queue/bulk", queueHandlers.handleQueueBulk).Methods("POST")
	}

	// Log and monitoring routes (Task 5.3)
	if s.logService != nil {
		logHandlers := NewLogHandlers(s.logService)

		// Basic log endpoints
		api.HandleFunc("/logs", logHandlers.handleLogsList).Methods("GET")
		api.HandleFunc("/logs/search", logHandlers.handleLogsSearch).Methods("POST")
		api.HandleFunc("/logs/tail", logHandlers.handleLogsTail).Methods("GET")
		api.HandleFunc("/logs/export", logHandlers.handleExportLogs).Methods("GET")
		api.HandleFunc("/logs/statistics", logHandlers.handleLogStatistics).Methods("GET")

		// Message-specific log endpoints
		api.HandleFunc("/logs/messages/{id}/history", logHandlers.handleMessageHistory).Methods("GET")
		api.HandleFunc("/logs/messages/{id}/correlation", logHandlers.handleMessageCorrelation).Methods("GET")
		api.HandleFunc("/logs/messages/{id}/similar", logHandlers.handleSimilarMessages).Methods("GET")

		// Service management endpoints
		api.HandleFunc("/logs/service/status", logHandlers.handleServiceStatus).Methods("GET")
		api.HandleFunc("/logs/correlation/trigger", logHandlers.handleTriggerCorrelation).Methods("POST")

		// Dashboard endpoint
		api.HandleFunc("/dashboard", logHandlers.handleDashboard).Methods("GET")
	}

	// Reporting routes (Task 5.4)
	if s.logService != nil && s.repository != nil {
		reportsHandlers := NewReportsHandlers(s.logService, s.queueService, s.repository)

		// Core reporting endpoints
		api.HandleFunc("/reports/deliverability", reportsHandlers.handleDeliverabilityReport).Methods("GET")
		api.HandleFunc("/reports/volume", reportsHandlers.handleVolumeReport).Methods("GET")
		api.HandleFunc("/reports/failures", reportsHandlers.handleFailureReport).Methods("GET")

		// Message tracing
		api.HandleFunc("/messages/{id}/trace", reportsHandlers.handleMessageTrace).Methods("GET")

		// Additional reporting endpoints
		api.HandleFunc("/reports/top-senders", reportsHandlers.handleTopSenders).Methods("GET")
		api.HandleFunc("/reports/top-recipients", reportsHandlers.handleTopRecipients).Methods("GET")
		api.HandleFunc("/reports/domains", reportsHandlers.handleDomainAnalysis).Methods("GET")
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         s.config.GetAddress(),
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.IdleTimeout) * time.Second,
	}

	log.Printf("Starting API server on %s", s.config.GetAddress())
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping API server...")
	return s.httpServer.Shutdown(ctx)
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		},
	}

	WriteJSONResponse(w, http.StatusOK, response)
}
