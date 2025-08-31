type WebSocketEventHandler = (data: any) => void;
type ConnectionStatusCallback = (status: 'connected' | 'disconnected' | 'connecting') => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 1000;
  private eventHandlers: Map<string, WebSocketEventHandler[]> = new Map();
  private subscriptions: Map<string, WebSocketEventHandler[]> = new Map();
  private connectionStatusCallback: ConnectionStatusCallback | null = null;
  private isConnecting = false;
  private reconnectTimeout: NodeJS.Timeout | null = null;

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
    console.log('WebSocket connect() called - Current state:', {
      isConnecting: this.isConnecting,
      wsState: this.ws?.readyState,
      wsStateText: this.ws?.readyState === 0 ? 'CONNECTING' : 
                   this.ws?.readyState === 1 ? 'OPEN' : 
                   this.ws?.readyState === 2 ? 'CLOSING' : 
                   this.ws?.readyState === 3 ? 'CLOSED' : 'NONE'
    });

    // Prevent multiple simultaneous connections
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.CONNECTING)) {
      console.log('WebSocket connect() - Already connecting, skipping');
      return Promise.resolve();
    }

    // If already connected, return immediately
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket connect() - Already connected, skipping');
      return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
      try {
        this.isConnecting = true;
        this.connectionStatusCallback?.('connecting');
        
        // Close any existing connection
        if (this.ws) {
          this.ws.close();
          this.ws = null;
        }

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.connectionStatusCallback?.('connected');
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
          this.isConnecting = false;
          this.connectionStatusCallback?.('disconnected');
          // Only attempt reconnect if it wasn't a manual disconnect
          if (event.code !== 1000 && event.code !== 1001) {
            this.handleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.isConnecting = false;
          this.connectionStatusCallback?.('disconnected');
          reject(error);
        };
      } catch (error) {
        this.isConnecting = false;
        this.connectionStatusCallback?.('disconnected');
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
    // Clear any existing reconnect timeout
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    // Don't reconnect if already connecting or connected
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      this.reconnectTimeout = setTimeout(() => {
        this.connect().catch(error => {
          console.error('Reconnection failed:', error);
        });
      }, this.reconnectInterval * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
      this.connectionStatusCallback?.('disconnected');
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
    // Clear any reconnection timeout
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    // Reset reconnect attempts
    this.reconnectAttempts = 0;
    this.isConnecting = false;

    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect'); // Code 1000 = normal closure
      this.ws = null;
    }
    this.connectionStatusCallback?.('disconnected');
  }

  setConnectionStatusCallback(callback: ConnectionStatusCallback | null) {
    this.connectionStatusCallback = callback;
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

// Singleton pattern to ensure only one instance
let webSocketServiceInstance: WebSocketService | null = null;

export const webSocketService = (() => {
  if (!webSocketServiceInstance) {
    webSocketServiceInstance = new WebSocketService();
    console.log('WebSocket service instance created');
  }
  return webSocketServiceInstance;
})();