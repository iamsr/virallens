import type { Message, WSMessage, OutgoingMessage } from '../types';
import { useAuthStore } from '../stores/authStore';

const WS_BASE_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';

type MessageHandler = (message: Message) => void;
type ErrorHandler = (error: string) => void;
type ConnectionHandler = () => void;
type PresenceHandler = (userId: string, status: 'online' | 'offline') => void;
type PresenceListHandler = (userIds: string[]) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = 1000;
  private messageHandlers: MessageHandler[] = [];
  private errorHandlers: ErrorHandler[] = [];
  private connectHandlers: ConnectionHandler[] = [];
  private disconnectHandlers: ConnectionHandler[] = [];
  private presenceHandlers: PresenceHandler[] = [];
  private presenceListHandlers: PresenceListHandler[] = [];
  private isIntentionalClose = false;

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    const token = useAuthStore.getState().accessToken;
    if (!token) return;

    this.isIntentionalClose = false;
    const url = `${WS_BASE_URL}/ws?token=${token}`;

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        console.log('[WS] Connected successfully');
        this.reconnectAttempts = 0;
        if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }
        this.connectHandlers.forEach((handler) => handler());
      };

      this.ws.onmessage = (event) => {
        try {
          const messages = typeof event.data === 'string' ? event.data.split('\n') : [event.data];
          for (const msgStr of messages) {
            if (!msgStr.trim()) continue;
            const data: WSMessage = JSON.parse(msgStr);
            this.handleMessage(data);
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error, event.data);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.errorHandlers.forEach((handler) =>
          handler('WebSocket connection error')
        );
      };

      this.ws.onclose = (event) => {
        console.log('[WS] Disconnected. code:', event.code, 'reason:', event.reason || 'none');
        this.disconnectHandlers.forEach((handler) => handler());

        // Attempt to reconnect if not intentional close
        if (!this.isIntentionalClose) {
          this.attemptReconnect();
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.attemptReconnect();
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] Max reconnection attempts reached. Giving up.');
      this.errorHandlers.forEach((handler) =>
        handler('Failed to reconnect to server')
      );
      return;
    }

    this.reconnectAttempts++;
    // Cap delay at 30s
    const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), 30000);

    console.log(`[WS] Reconnecting (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${delay}ms...`);

    this.reconnectTimer = setTimeout(() => {
      // Just reconnect with whatever token is currently in the store.
      // The Zustand auth store handles silently refreshing the access token
      // via the HTTP interceptor when any other API call happens.
      this.connect();
    }, delay);
  }

  private handleMessage(data: WSMessage): void {
    switch (data.type) {
      case 'message':
        if (data.data) {
          this.messageHandlers.forEach((handler) => handler(data.data));
        }
        break;

      case 'error':
        if (data.message) {
          this.errorHandlers.forEach((handler) => handler(data.message!));
        }
        break;

      case 'typing':
        // Handle typing indicators if needed
        break;

      case 'presence':
        if (data.data && data.data.user_id && data.data.status) {
          this.presenceHandlers.forEach((h) => h(data.data.user_id, data.data.status));
        }
        break;

      case 'presence_list':
        if (Array.isArray(data.data)) {
          this.presenceListHandlers.forEach((h) => h(data.data));
        }
        break;

      default:
        console.warn('Unknown message type:', data.type);
    }
  }

  sendMessage(message: OutgoingMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const payload = JSON.stringify(message);
      console.log('[WS] Sending message:', payload);
      this.ws.send(payload);
    } else {
      console.error('[WS] Cannot send — readyState:', this.ws?.readyState ?? 'null (no socket)');
      this.errorHandlers.forEach((handler) =>
        handler('Cannot send message: not connected')
      );
    }
  }

  disconnect(): void {
    this.isIntentionalClose = true;
    if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.push(handler);
    // Return unsubscribe function
    return () => {
      this.messageHandlers = this.messageHandlers.filter((h) => h !== handler);
    };
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.push(handler);
    return () => {
      this.errorHandlers = this.errorHandlers.filter((h) => h !== handler);
    };
  }

  onConnect(handler: ConnectionHandler): () => void {
    this.connectHandlers.push(handler);
    return () => {
      this.connectHandlers = this.connectHandlers.filter((h) => h !== handler);
    };
  }

  onDisconnect(handler: ConnectionHandler): () => void {
    this.disconnectHandlers.push(handler);
    return () => {
      this.disconnectHandlers = this.disconnectHandlers.filter(
        (h) => h !== handler
      );
    };
  }

  onPresence(handler: PresenceHandler): () => void {
    this.presenceHandlers.push(handler);
    return () => {
      this.presenceHandlers = this.presenceHandlers.filter((h) => h !== handler);
    };
  }

  onPresenceList(handler: PresenceListHandler): () => void {
    this.presenceListHandlers.push(handler);
    return () => {
      this.presenceListHandlers = this.presenceListHandlers.filter((h) => h !== handler);
    };
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  getReadyState(): number | null {
    return this.ws?.readyState ?? null;
  }
}

// Export singleton instance
export const wsClient = new WebSocketClient();
