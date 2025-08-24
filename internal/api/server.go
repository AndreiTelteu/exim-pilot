package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/auth"
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
	authService  *auth.Service
}

// NewServer creates a new API server instance
func NewServer(config *Config, queueService *queue.Service, logService *logprocessor.Service, repository *database.Repository, db *database.DB) *Server {
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
		authService:  auth.NewService(db),
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

	// Security headers middleware (first for all responses)
	s.router.Use(s.securityHeadersMiddleware)

	// Request logging middleware (if enabled)
	if s.config.LogRequests {
		s.router.Use(s.loggingMiddleware)
	}

	// Error handling middleware
	s.router.Use(s.errorHandlingMiddleware)

	// Content-type middleware
	s.router.Use(s.contentTypeMiddleware)

	// Input validation middleware
	s.router.Use(s.validationMiddleware)

	// Audit logging middleware (after validation, before auth)
	s.router.Use(s.auditMiddleware)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint (no auth required)
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Authentication routes (no auth required for login)
	authHandlers := NewAuthHandlers(s.authService)
	api.HandleFunc("/auth/login", authHandlers.handleLogin).Methods("POST")

	// Protected routes - apply auth middleware
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.authMiddleware)

	// Auth routes that require authentication
	protected.HandleFunc("/auth/logout", authHandlers.handleLogout).Methods("POST")
	protected.HandleFunc("/auth/me", authHandlers.handleMe).Methods("GET")

	// Queue management routes (Task 5.2) - Protected
	if s.queueService != nil {
		queueHandlers := NewQueueHandlers(s.queueService)

		// Queue listing and search
		protected.HandleFunc("/queue", queueHandlers.handleQueueList).Methods("GET")
		protected.HandleFunc("/queue/search", queueHandlers.handleQueueSearch).Methods("POST")
		protected.HandleFunc("/queue/health", queueHandlers.handleQueueHealth).Methods("GET")
		protected.HandleFunc("/queue/statistics", queueHandlers.handleQueueStatistics).Methods("GET")

		// Individual message operations
		protected.HandleFunc("/queue/{id}", queueHandlers.handleQueueDetails).Methods("GET")
		protected.HandleFunc("/queue/{id}/deliver", queueHandlers.handleQueueDeliver).Methods("POST")
		protected.HandleFunc("/queue/{id}/freeze", queueHandlers.handleQueueFreeze).Methods("POST")
		protected.HandleFunc("/queue/{id}/thaw", queueHandlers.handleQueueThaw).Methods("POST")
		protected.HandleFunc("/queue/{id}", queueHandlers.handleQueueDelete).Methods("DELETE")
		protected.HandleFunc("/queue/{id}/history", queueHandlers.handleQueueHistory).Methods("GET")

		// Bulk operations
		protected.HandleFunc("/queue/bulk", queueHandlers.handleQueueBulk).Methods("POST")
	}

	// Log and monitoring routes (Task 5.3) - Protected
	if s.logService != nil {
		logHandlers := NewLogHandlers(s.logService)

		// Basic log endpoints
		protected.HandleFunc("/logs", logHandlers.handleLogsList).Methods("GET")
		protected.HandleFunc("/logs/search", logHandlers.handleLogsSearch).Methods("POST")
		protected.HandleFunc("/logs/tail", logHandlers.handleLogsTail).Methods("GET")
		protected.HandleFunc("/logs/export", logHandlers.handleExportLogs).Methods("GET")
		protected.HandleFunc("/logs/statistics", logHandlers.handleLogStatistics).Methods("GET")

		// Message-specific log endpoints
		protected.HandleFunc("/logs/messages/{id}/history", logHandlers.handleMessageHistory).Methods("GET")
		protected.HandleFunc("/logs/messages/{id}/correlation", logHandlers.handleMessageCorrelation).Methods("GET")
		protected.HandleFunc("/logs/messages/{id}/similar", logHandlers.handleSimilarMessages).Methods("GET")

		// Service management endpoints
		protected.HandleFunc("/logs/service/status", logHandlers.handleServiceStatus).Methods("GET")
		protected.HandleFunc("/logs/correlation/trigger", logHandlers.handleTriggerCorrelation).Methods("POST")

		// Dashboard endpoint
		protected.HandleFunc("/dashboard", logHandlers.handleDashboard).Methods("GET")
	}

	// Reporting routes (Task 5.4) - Protected
	if s.logService != nil && s.repository != nil {
		reportsHandlers := NewReportsHandlers(s.logService, s.queueService, s.repository)

		// Core reporting endpoints
		protected.HandleFunc("/reports/deliverability", reportsHandlers.handleDeliverabilityReport).Methods("GET")
		protected.HandleFunc("/reports/volume", reportsHandlers.handleVolumeReport).Methods("GET")
		protected.HandleFunc("/reports/failures", reportsHandlers.handleFailureReport).Methods("GET")

		// Message tracing (legacy endpoint)
		protected.HandleFunc("/messages/{id}/trace", reportsHandlers.handleMessageTrace).Methods("GET")

		// Additional reporting endpoints
		protected.HandleFunc("/reports/top-senders", reportsHandlers.handleTopSenders).Methods("GET")
		protected.HandleFunc("/reports/top-recipients", reportsHandlers.handleTopRecipients).Methods("GET")
		protected.HandleFunc("/reports/domains", reportsHandlers.handleDomainAnalysis).Methods("GET")
	}

	// Enhanced Message Tracing routes (Task 11.1) - Protected
	if s.repository != nil {
		messageTraceHandlers := NewMessageTraceHandlers(s.repository, s.queueService, s.logService)

		// Enhanced message delivery tracing (Task 11.1)
		protected.HandleFunc("/messages/{id}/delivery-trace", messageTraceHandlers.handleMessageDeliveryTrace).Methods("GET")
		protected.HandleFunc("/messages/{id}/recipients/{recipient}/history", messageTraceHandlers.handleRecipientDeliveryHistory).Methods("GET")
		protected.HandleFunc("/messages/{id}/timeline", messageTraceHandlers.handleDeliveryTimeline).Methods("GET")
		protected.HandleFunc("/messages/{id}/retry-schedule", messageTraceHandlers.handleRetrySchedule).Methods("GET")
		protected.HandleFunc("/messages/{id}/delivery-stats", messageTraceHandlers.handleMessageDeliveryStats).Methods("GET")

		// Delivery attempt details
		protected.HandleFunc("/delivery-attempts/{id}", messageTraceHandlers.handleDeliveryAttemptDetails).Methods("GET")

		// Troubleshooting and notes functionality (Task 11.2)
		protected.HandleFunc("/messages/{id}/threaded-timeline", messageTraceHandlers.handleThreadedTimeline).Methods("GET")
		protected.HandleFunc("/messages/{id}/content", messageTraceHandlers.handleMessageContent).Methods("GET")

		// Message notes
		protected.HandleFunc("/messages/{id}/notes", messageTraceHandlers.handleMessageNotes).Methods("GET")
		protected.HandleFunc("/messages/{id}/notes", messageTraceHandlers.handleCreateMessageNote).Methods("POST")
		protected.HandleFunc("/messages/{id}/notes/{noteId}", messageTraceHandlers.handleUpdateMessageNote).Methods("PUT")
		protected.HandleFunc("/messages/{id}/notes/{noteId}", messageTraceHandlers.handleDeleteMessageNote).Methods("DELETE")

		// Message tags
		protected.HandleFunc("/messages/{id}/tags", messageTraceHandlers.handleMessageTags).Methods("GET")
		protected.HandleFunc("/messages/{id}/tags", messageTraceHandlers.handleCreateMessageTag).Methods("POST")
		protected.HandleFunc("/messages/{id}/tags/{tagId}", messageTraceHandlers.handleDeleteMessageTag).Methods("DELETE")

		// Popular tags
		protected.HandleFunc("/tags/popular", messageTraceHandlers.handlePopularTags).Methods("GET")
	}

	// Performance and Optimization routes (Task 13.1) - Protected
	performanceHandlers := NewPerformanceHandlers(s.repository.GetDB())

	// Database performance endpoints
	protected.HandleFunc("/performance/database/stats", performanceHandlers.handleDatabaseStats).Methods("GET")
	protected.HandleFunc("/performance/database/optimize", performanceHandlers.handleOptimizeDatabase).Methods("POST")
	protected.HandleFunc("/performance/database/query-hints", performanceHandlers.handleQueryOptimizationHints).Methods("GET")

	// Data retention endpoints
	protected.HandleFunc("/performance/retention/status", performanceHandlers.handleRetentionStatus).Methods("GET")
	protected.HandleFunc("/performance/retention/cleanup", performanceHandlers.handleCleanupExpiredData).Methods("POST")

	// General performance endpoints
	protected.HandleFunc("/performance/metrics", performanceHandlers.handlePerformanceMetrics).Methods("GET")
	protected.HandleFunc("/performance/cache/stats", performanceHandlers.handleCacheStats).Methods("GET")
	protected.HandleFunc("/performance/memory/stats", performanceHandlers.handleMemoryStats).Methods("GET")

	// Batch optimization
	protected.HandleFunc("/performance/batch/optimize", performanceHandlers.handleBatchOptimization).Methods("POST")

	// Performance configuration
	protected.HandleFunc("/performance/config", performanceHandlers.handlePerformanceConfig).Methods("GET", "POST")

	// Performance testing
	protected.HandleFunc("/performance/test", performanceHandlers.handlePerformanceTest).Methods("POST")
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
