import { io, Socket } from 'socket.io-client';
import { toast } from 'react-hot-toast';

interface WebSocketConfig {
  url: string;
  reconnection: boolean;
  reconnectionAttempts: number;
  reconnectionDelay: number;
  timeout: number;
}

interface EventHandler {
  event: string;
  callback: (data: any) => void;
}

class WebSocketClient {
  private socket: Socket | null = null;
  private config: WebSocketConfig;
  private eventHandlers: Map<string, Set<(data: any) => void>> = new Map();
  private connectionPromise: Promise<void> | null = null;
  private isConnecting = false;
  private heartbeatInterval: NodeJS.Timeout | null = null;

  constructor() {
    this.config = {
      url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080',
      reconnection: true,
      reconnectionAttempts: 5,
      reconnectionDelay: 3000,
      timeout: 20000,
    };
  }

  async connect(token?: string): Promise<void> {
    if (this.socket?.connected) {
      return Promise.resolve();
    }

    if (this.isConnecting && this.connectionPromise) {
      return this.connectionPromise;
    }

    this.isConnecting = true;
    
    this.connectionPromise = new Promise((resolve, reject) => {
      try {
        this.socket = io(this.config.url, {
          reconnection: this.config.reconnection,
          reconnectionAttempts: this.config.reconnectionAttempts,
          reconnectionDelay: this.config.reconnectionDelay,
          timeout: this.config.timeout,
          auth: token ? { token } : undefined,
          transports: ['websocket', 'polling'],
        });

        this.setupEventListeners(resolve, reject);
        this.startHeartbeat();
      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });

    return this.connectionPromise;
  }

  private setupEventListeners(resolve: () => void, reject: (error: any) => void): void {
    if (!this.socket) return;

    // Connection events
    this.socket.on('connect', () => {
      console.log('WebSocket connected');
      this.isConnecting = false;
      resolve();
      this.resubscribeToEvents();
    });

    this.socket.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error);
      this.isConnecting = false;
      reject(error);
    });

    this.socket.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason);
      if (reason === 'io server disconnect') {
        // Server initiated disconnect, attempt to reconnect
        this.socket?.connect();
      }
    });

    this.socket.on('reconnect', (attemptNumber) => {
      console.log(`WebSocket reconnected after ${attemptNumber} attempts`);
      toast.success('Connection restored');
      this.resubscribeToEvents();
    });

    this.socket.on('reconnect_error', (error) => {
      console.error('WebSocket reconnection error:', error);
    });

    this.socket.on('reconnect_failed', () => {
      console.error('WebSocket reconnection failed');
      toast.error('Connection lost. Please refresh the page.');
    });

    // Custom events
    this.socket.on('error', (error) => {
      console.error('WebSocket error:', error);
      toast.error(`Connection error: ${error.message || 'Unknown error'}`);
    });

    this.socket.on('notification', (data) => {
      this.handleNotification(data);
    });
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatInterval = setInterval(() => {
      if (this.socket?.connected) {
        this.socket.emit('ping');
      }
    }, 30000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private handleNotification(data: any): void {
    const { type, title, message } = data;
    
    switch (type) {
      case 'success':
        toast.success(message, { duration: 4000 });
        break;
      case 'error':
        toast.error(message, { duration: 5000 });
        break;
      case 'warning':
        toast(message, { icon: '⚠️', duration: 4000 });
        break;
      case 'info':
      default:
        toast(message, { icon: 'ℹ️', duration: 3000 });
    }

    // Trigger browser notification if permitted
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification(title || 'ChengetoPay', {
        body: message,
        icon: '/favicon.ico',
        badge: '/icon-192x192.png',
      });
    }
  }

  private resubscribeToEvents(): void {
    this.eventHandlers.forEach((callbacks, event) => {
      callbacks.forEach(callback => {
        this.socket?.on(event, callback);
      });
    });
  }

  on(event: string, callback: (data: any) => void): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    
    this.eventHandlers.get(event)?.add(callback);
    this.socket?.on(event, callback);
  }

  off(event: string, callback?: (data: any) => void): void {
    if (callback) {
      this.eventHandlers.get(event)?.delete(callback);
      this.socket?.off(event, callback);
    } else {
      this.eventHandlers.delete(event);
      this.socket?.off(event);
    }
  }

  emit(event: string, data?: any): void {
    if (!this.socket?.connected) {
      console.warn('WebSocket not connected. Queuing event:', event);
      this.connect().then(() => {
        this.socket?.emit(event, data);
      });
      return;
    }
    
    this.socket.emit(event, data);
  }

  // Typed event emitters for specific services
  subscribeToPayments(callback: (data: any) => void): void {
    this.on('payment:update', callback);
  }

  subscribeToTransactions(callback: (data: any) => void): void {
    this.on('transaction:update', callback);
  }

  subscribeToEscrow(escrowId: string, callback: (data: any) => void): void {
    this.emit('escrow:subscribe', { escrowId });
    this.on(`escrow:${escrowId}:update`, callback);
  }

  unsubscribeFromEscrow(escrowId: string): void {
    this.emit('escrow:unsubscribe', { escrowId });
    this.off(`escrow:${escrowId}:update`);
  }

  subscribeToMerchantUpdates(callback: (data: any) => void): void {
    this.on('merchant:update', callback);
  }

  subscribeToSystemAlerts(callback: (data: any) => void): void {
    this.on('system:alert', callback);
  }

  // Room management
  joinRoom(room: string): void {
    this.emit('join:room', { room });
  }

  leaveRoom(room: string): void {
    this.emit('leave:room', { room });
  }

  sendToRoom(room: string, event: string, data: any): void {
    this.emit('room:message', { room, event, data });
  }

  disconnect(): void {
    this.stopHeartbeat();
    this.eventHandlers.clear();
    this.socket?.disconnect();
    this.socket = null;
    this.connectionPromise = null;
    this.isConnecting = false;
  }

  isConnected(): boolean {
    return this.socket?.connected || false;
  }

  getSocketId(): string | undefined {
    return this.socket?.id;
  }
}

// Export singleton instance
const wsClient = new WebSocketClient();
export default wsClient;
