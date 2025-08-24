package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/andreitelteu/exim-pilot/internal/api"
	"github.com/andreitelteu/exim-pilot/internal/auth"
	"github.com/andreitelteu/exim-pilot/internal/database"
	"github.com/andreitelteu/exim-pilot/internal/logprocessor"
	"github.com/andreitelteu/exim-pilot/internal/queue"
)

// TestServer represents a test server instance
type TestServer struct {
	server     *api.Server
	httpServer *httptest.Server
	db         *database.DB
	repository *database.Repository
	cleanup    func()
}

// NewTestServer creates a new test server for integration testing
func NewTestServer(t *testing.T) *TestServer {
	// Create temporary database
	dbPath := fmt.Sprintf("test_%d.db", time.Now().UnixNano())

	config := &database.Config{
		Path:            dbPath,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	}

	db, err := database.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Initialize database schema
	if _, err := db.Exec(database.Schema); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Create repository
	repository := database.NewRepository(db)

	// Create services
	queueService := queue.NewService(queue.DefaultConfig())
	logService := logprocessor.NewService(repository, logprocessor.DefaultServiceConfig())

	// Create API server
	apiConfig := api.NewConfig()
	apiConfig.Port = 0 // Use random port for testing
	server := api.NewServer(apiConfig, queueService, logService, repository, db)

	// Create HTTP test server
	httpServer := httptest.NewServer(server.Router())

	cleanup := func() {
		httpServer.Close()
		db.Close()
		os.Remove(dbPath)
	}

	return &TestServer{
		server:     server,
		httpServer: httpServer,
		db:         db,
		repository: repository,
		cleanup:    cleanup,
	}
}

// Close cleans up the test server
func (ts *TestServer) Close() {
	ts.cleanup()
}

// URL returns the base URL for the test server
func (ts *TestServer) URL() string {
	return ts.httpServer.URL
}

// CreateTestUser creates a test user and returns auth token
func (ts *TestServer) CreateTestUser(t *testing.T) string {
	authService := auth.NewService(ts.db)

	user, err := authService.CreateUser(context.Background(), &auth.CreateUserRequest{
		Username: "testuser",
		Password: "testpass123",
		Email:    "test@example.com",
		FullName: "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	session, err := authService.CreateSession(context.Background(), user.ID, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return session.ID
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL() + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to make health check request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response["success"].(bool) {
		t.Error("Expected success to be true")
	}

	data := response["data"].(map[string]interface{})
	if data["status"] != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %v", data["status"])
	}
}

// TestAuthenticationFlow tests the complete authentication flow
func TestAuthenticationFlow(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Test login with invalid credentials
	loginData := map[string]string{
		"username": "nonexistent",
		"password": "wrongpass",
	}
	loginJSON, _ := json.Marshal(loginData)

	resp, err := http.Post(ts.URL()+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatalf("Failed to make login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid credentials, got %d", resp.StatusCode)
	}

	// Create test user
	authService := auth.NewService(ts.db)
	_, err = authService.CreateUser(context.Background(), &auth.CreateUserRequest{
		Username: "testuser",
		Password: "testpass123",
		Email:    "test@example.com",
		FullName: "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test login with valid credentials
	loginData["username"] = "testuser"
	loginData["password"] = "testpass123"
	loginJSON, _ = json.Marshal(loginData)

	resp, err = http.Post(ts.URL()+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatalf("Failed to make login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for valid credentials, got %d", resp.StatusCode)
	}

	var loginResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if !loginResponse["success"].(bool) {
		t.Error("Expected login success to be true")
	}

	// Extract session token from cookies
	cookies := resp.Cookies()
	var sessionToken string
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionToken = cookie.Value
			break
		}
	}

	if sessionToken == "" {
		t.Fatal("No session token found in response cookies")
	}

	// Test accessing protected endpoint with session token
	req, err := http.NewRequest("GET", ts.URL()+"/api/v1/auth/me", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionToken})

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make authenticated request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for authenticated request, got %d", resp.StatusCode)
	}
}

// TestQueueEndpoints tests queue management endpoints
func TestQueueEndpoints(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	token := ts.CreateTestUser(t)

	// Test queue listing without authentication
	resp, err := http.Get(ts.URL() + "/api/v1/queue")
	if err != nil {
		t.Fatalf("Failed to make queue request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthenticated request, got %d", resp.StatusCode)
	}

	// Test queue listing with authentication
	req, err := http.NewRequest("GET", ts.URL()+"/api/v1/queue", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make authenticated queue request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for authenticated queue request, got %d", resp.StatusCode)
	}

	var queueResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&queueResponse); err != nil {
		t.Fatalf("Failed to decode queue response: %v", err)
	}

	if !queueResponse["success"].(bool) {
		t.Error("Expected queue response success to be true")
	}

	// Test queue search
	searchData := map[string]interface{}{
		"criteria": map[string]interface{}{
			"status": "queued",
		},
		"page":     1,
		"per_page": 25,
	}
	searchJSON, _ := json.Marshal(searchData)

	req, err = http.NewRequest("POST", ts.URL()+"/api/v1/queue/search", bytes.NewBuffer(searchJSON))
	if err != nil {
		t.Fatalf("Failed to create search request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make queue search request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for queue search, got %d", resp.StatusCode)
	}
}

// TestLogEndpoints tests log viewing endpoints
func TestLogEndpoints(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	token := ts.CreateTestUser(t)

	// Insert test log entries
	ctx := context.Background()
	testEntries := []*database.LogEntry{
		{
			Timestamp:  time.Now(),
			MessageID:  stringPtr("test-message-1"),
			LogType:    "main",
			Event:      "arrival",
			Host:       stringPtr("localhost"),
			Sender:     stringPtr("test@example.com"),
			Recipients: []string{"recipient@example.com"},
			Size:       intPtr(1024),
			Status:     stringPtr("received"),
			RawLine:    "2024-01-01 12:00:00 test log line",
		},
		{
			Timestamp:  time.Now().Add(-time.Hour),
			MessageID:  stringPtr("test-message-2"),
			LogType:    "main",
			Event:      "delivery",
			Host:       stringPtr("localhost"),
			Sender:     stringPtr("test2@example.com"),
			Recipients: []string{"recipient2@example.com"},
			Size:       intPtr(2048),
			Status:     stringPtr("delivered"),
			RawLine:    "2024-01-01 11:00:00 test log line 2",
		},
	}

	for _, entry := range testEntries {
		if err := ts.repository.CreateLogEntry(ctx, entry); err != nil {
			t.Fatalf("Failed to create test log entry: %v", err)
		}
	}

	// Test log listing
	req, err := http.NewRequest("GET", ts.URL()+"/api/v1/logs?per_page=10", nil)
	if err != nil {
		t.Fatalf("Failed to create logs request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make logs request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for logs request, got %d", resp.StatusCode)
	}

	var logsResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&logsResponse); err != nil {
		t.Fatalf("Failed to decode logs response: %v", err)
	}

	if !logsResponse["success"].(bool) {
		t.Error("Expected logs response success to be true")
	}

	data := logsResponse["data"].(map[string]interface{})
	entries := data["entries"].([]interface{})

	if len(entries) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(entries))
	}

	// Test log search
	searchData := map[string]interface{}{
		"criteria": map[string]interface{}{
			"log_type": "main",
			"event":    "arrival",
		},
		"page":     1,
		"per_page": 10,
	}
	searchJSON, _ := json.Marshal(searchData)

	req, err = http.NewRequest("POST", ts.URL()+"/api/v1/logs/search", bytes.NewBuffer(searchJSON))
	if err != nil {
		t.Fatalf("Failed to create log search request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make log search request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for log search, got %d", resp.StatusCode)
	}
}

// TestPerformanceEndpoints tests performance monitoring endpoints
func TestPerformanceEndpoints(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	token := ts.CreateTestUser(t)

	// Test performance metrics
	req, err := http.NewRequest("GET", ts.URL()+"/api/v1/performance/metrics", nil)
	if err != nil {
		t.Fatalf("Failed to create performance metrics request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make performance metrics request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for performance metrics, got %d", resp.StatusCode)
	}

	var metricsResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metricsResponse); err != nil {
		t.Fatalf("Failed to decode metrics response: %v", err)
	}

	if !metricsResponse["success"].(bool) {
		t.Error("Expected metrics response success to be true")
	}

	// Test database optimization
	req, err = http.NewRequest("POST", ts.URL()+"/api/v1/performance/database/optimize", nil)
	if err != nil {
		t.Fatalf("Failed to create optimization request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make optimization request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for database optimization, got %d", resp.StatusCode)
	}
}

// TestDashboardEndpoint tests the dashboard endpoint
func TestDashboardEndpoint(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	token := ts.CreateTestUser(t)

	// Test dashboard
	req, err := http.NewRequest("GET", ts.URL()+"/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("Failed to create dashboard request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make dashboard request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for dashboard, got %d", resp.StatusCode)
	}

	var dashboardResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&dashboardResponse); err != nil {
		t.Fatalf("Failed to decode dashboard response: %v", err)
	}

	if !dashboardResponse["success"].(bool) {
		t.Error("Expected dashboard response success to be true")
	}

	data := dashboardResponse["data"].(map[string]interface{})
	if _, exists := data["log_statistics"]; !exists {
		t.Error("Expected dashboard to contain log_statistics")
	}

	if _, exists := data["service_status"]; !exists {
		t.Error("Expected dashboard to contain service_status")
	}
}

// TestErrorHandling tests API error handling
func TestErrorHandling(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	token := ts.CreateTestUser(t)

	// Test invalid JSON
	req, err := http.NewRequest("POST", ts.URL()+"/api/v1/queue/search", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}

	// Test non-existent endpoint
	req, err = http.NewRequest("GET", ts.URL()+"/api/v1/nonexistent", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: token})

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent endpoint, got %d", resp.StatusCode)
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int64) *int64 {
	return &i
}
