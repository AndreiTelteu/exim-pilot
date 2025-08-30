type WebSocketEventHandler = (data: any) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 1000;
  private eventHandlers: Map<string, WebSocketEventHandler[]> = new Map();
  private subscriptions: Map<string, WebSocketEventHandler[]> = new Map();

  constructor(url: string|null = null) {
    if (url === null) {
      const currentUrl = new URL(location.href);
      currentUrl.protocol = currentUrl.protocol === 'https:' ? 'wss:' : 'ws:';
      currentUrl.pathname = '/ws';
      url = currentUrl.toString();
    }
    this.url = url;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected:', event.code, event.reason);
          this.handleReconnect();
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  private handleMessage(message: any) {
    const { type, data, endpoint } = message;
    
    // Handle subscription-based messages
    if (endpoint) {
      const handlers = this.subscriptions.get(endpoint) || [];
      handlers.forEach(handler => handler(message));
    } else {
      // Handle event-based messages
      const handlers = this.eventHandlers.get(type) || [];
      handlers.forEach(handler => handler(data));
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      setTimeout(() => {
        this.connect().catch(error => {
          console.error('Reconnection failed:', error);
        });
      }, this.reconnectInterval * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  on(eventType: string, handler: WebSocketEventHandler) {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    this.eventHandlers.get(eventType)!.push(handler);
  }

  off(eventType: string, handler: WebSocketEventHandler) {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  send(message: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // Subscription-based methods for endpoints like log tailing
  subscribe(endpoint: string, handler: WebSocketEventHandler) {
    if (!this.subscriptions.has(endpoint)) {
      this.subscriptions.set(endpoint, []);
    }
    this.subscriptions.get(endpoint)!.push(handler);

    // Send subscription message to server
    this.send({
      type: 'subscribe',
      endpoint: endpoint
    });
  }

  unsubscribe(endpoint: string, handler?: WebSocketEventHandler) {
    if (handler) {
      const handlers = this.subscriptions.get(endpoint);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
        }
        
        // If no more handlers, remove the subscription entirely
        if (handlers.length === 0) {
          this.subscriptions.delete(endpoint);
          this.send({
            type: 'unsubscribe',
            endpoint: endpoint
          });
        }
      }
    } else {
      // Remove all handlers for this endpoint
      this.subscriptions.delete(endpoint);
      this.send({
        type: 'unsubscribe',
        endpoint: endpoint
      });
    }
  }
}

export const webSocketService = new WebSocketService();