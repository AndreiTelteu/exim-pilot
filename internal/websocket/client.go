package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages from client
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Error unmarshaling client message: %v", err)
		return
	}

	switch msg.Type {
	case "subscribe":
		if msg.Endpoint != "" {
			c.hub.Subscribe(c, msg.Endpoint)
			c.sendResponse("subscribed", map[string]interface{}{
				"endpoint": msg.Endpoint,
				"status":   "success",
			})
		}

	case "unsubscribe":
		if msg.Endpoint != "" {
			c.hub.Unsubscribe(c, msg.Endpoint)
			c.sendResponse("unsubscribed", map[string]interface{}{
				"endpoint": msg.Endpoint,
				"status":   "success",
			})
		}

	case "ping":
		c.sendResponse("pong", map[string]interface{}{
			"timestamp": time.Now().UTC(),
		})

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// sendResponse sends a response message to the client
func (c *Client) sendResponse(messageType string, data interface{}) {
	response := Message{
		Type: messageType,
		Data: data,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	select {
	case c.send <- jsonData:
	default:
		log.Printf("Client send channel full, dropping response")
	}
}

// SendMessage sends a message directly to this client
func (c *Client) SendMessage(messageType string, data interface{}) {
	c.sendResponse(messageType, data)
}

// IsSubscribed checks if the client is subscribed to an endpoint
func (c *Client) IsSubscribed(endpoint string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscriptions[endpoint]
}

// GetSubscriptions returns a copy of the client's subscriptions
func (c *Client) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subscriptions := make([]string, 0, len(c.subscriptions))
	for endpoint := range c.subscriptions {
		subscriptions = append(subscriptions, endpoint)
	}
	return subscriptions
}
