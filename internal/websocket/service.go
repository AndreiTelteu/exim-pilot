package websocket

import (
	"context"
	"log"
	"sync"
	"time"
)

// Service manages WebSocket connections and real-time updates
type Service struct {
	hub     *Hub
	running bool
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewService creates a new WebSocket service
func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		hub:    NewHub(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the WebSocket service
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	go s.hub.Run()

	log.Println("WebSocket service started")
	return nil
}

// Stop stops the WebSocket service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.cancel()
	s.running = false

	log.Println("WebSocket service stopped")
	return nil
}

// GetHub returns the WebSocket hub
func (s *Service) GetHub() *Hub {
	return s.hub
}

// BroadcastQueueUpdate broadcasts queue update to all connected clients
func (s *Service) BroadcastQueueUpdate(data interface{}) {
	s.hub.BroadcastToAll("queue_update", data)
}

// BroadcastLogEntry broadcasts new log entry to subscribers
func (s *Service) BroadcastLogEntry(entry interface{}) {
	s.hub.BroadcastToSubscribers("/api/v1/logs/tail", entry)
}

// BroadcastDashboardUpdate broadcasts dashboard metrics update
func (s *Service) BroadcastDashboardUpdate(metrics interface{}) {
	s.hub.BroadcastToAll("dashboard_update", metrics)
}

// BroadcastMessageUpdate broadcasts message-specific updates
func (s *Service) BroadcastMessageUpdate(messageID string, data interface{}) {
	endpoint := "/api/v1/messages/" + messageID + "/updates"
	s.hub.BroadcastToSubscribers(endpoint, data)
}

// BroadcastSystemAlert broadcasts system alerts
func (s *Service) BroadcastSystemAlert(alert interface{}) {
	s.hub.BroadcastToAll("system_alert", alert)
}

// GetStats returns WebSocket service statistics
func (s *Service) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"running":       s.running,
		"total_clients": s.hub.GetClientCount(),
		"subscriptions": s.getSubscriptionStats(),
		"uptime":        time.Since(time.Now()).String(), // This would need proper tracking
	}
}

// getSubscriptionStats returns subscription statistics
func (s *Service) getSubscriptionStats() map[string]int {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	stats := make(map[string]int)
	for endpoint, subscribers := range s.hub.subscriptions {
		stats[endpoint] = len(subscribers)
	}
	return stats
}

// IsRunning returns whether the service is running
func (s *Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
