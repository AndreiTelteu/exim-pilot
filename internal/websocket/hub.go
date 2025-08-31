package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Subscription management
	subscriptions map[string]map[*Client]bool
	mu            sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client subscriptions
	subscriptions map[string]bool
	mu            sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data,omitempty"`
	Endpoint string      `json:"endpoint,omitempty"`
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin properly
		return true
	},
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		subscriptions: make(map[string]map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove client from all subscriptions
				h.mu.Lock()
				for endpoint, subscribers := range h.subscriptions {
					delete(subscribers, client)
					if len(subscribers) == 0 {
						delete(h.subscriptions, endpoint)
					}
				}
				h.mu.Unlock()

				log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// ServeWS handles websocket requests from the peer
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:           h,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(messageType string, data interface{}) {
	message := Message{
		Type: messageType,
		Data: data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	select {
	case h.broadcast <- jsonData:
	default:
		log.Printf("Broadcast channel full, dropping message")
	}
}

// BroadcastToSubscribers sends a message to clients subscribed to a specific endpoint
func (h *Hub) BroadcastToSubscribers(endpoint string, data interface{}) {
	h.mu.RLock()
	subscribers, exists := h.subscriptions[endpoint]
	if !exists || len(subscribers) == 0 {
		h.mu.RUnlock()
		return
	}

	// Create a copy of subscribers to avoid holding the lock during broadcast
	clientList := make([]*Client, 0, len(subscribers))
	for client := range subscribers {
		clientList = append(clientList, client)
	}
	h.mu.RUnlock()

	message := Message{
		Type:     "subscription_update",
		Data:     data,
		Endpoint: endpoint,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling subscription message: %v", err)
		return
	}

	for _, client := range clientList {
		select {
		case client.send <- jsonData:
		default:
			// Client's send channel is full, skip this client
			log.Printf("Client send channel full, skipping message")
		}
	}
}

// Subscribe adds a client to an endpoint subscription
func (h *Hub) Subscribe(client *Client, endpoint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.subscriptions[endpoint] == nil {
		h.subscriptions[endpoint] = make(map[*Client]bool)
	}
	h.subscriptions[endpoint][client] = true

	client.mu.Lock()
	client.subscriptions[endpoint] = true
	client.mu.Unlock()

	log.Printf("Client subscribed to endpoint: %s. Total subscribers: %d", endpoint, len(h.subscriptions[endpoint]))
}

// Unsubscribe removes a client from an endpoint subscription
func (h *Hub) Unsubscribe(client *Client, endpoint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, exists := h.subscriptions[endpoint]; exists {
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.subscriptions, endpoint)
		}
	}

	client.mu.Lock()
	delete(client.subscriptions, endpoint)
	client.mu.Unlock()

	log.Printf("Client unsubscribed from endpoint: %s", endpoint)
}

// GetSubscriberCount returns the number of subscribers for an endpoint
func (h *Hub) GetSubscriberCount(endpoint string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subscribers, exists := h.subscriptions[endpoint]; exists {
		return len(subscribers)
	}
	return 0
}

// GetClientCount returns the total number of connected clients
func (h *Hub) GetClientCount() int {
	return len(h.clients)
}
