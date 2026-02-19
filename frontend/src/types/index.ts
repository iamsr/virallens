// User types
export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

// Conversation types
export interface Conversation {
  id: string;
  participants: string[];
  created_at: string;
  updated_at: string;
}

// Group types
export interface Group {
  id: string;
  name: string;
  creator_id: string;
  members: string[];
  created_at: string;
  updated_at: string;
}

// Message types
export type MessageType = 'conversation' | 'group';

export interface Message {
  id: string;
  sender_id: string;
  conversation_id?: string;
  group_id?: string;
  content: string;
  type: MessageType;
  created_at: string;
}

// Auth types
export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

// WebSocket message types
export interface WSMessage {
  type: 'message' | 'error' | 'typing';
  data?: any;
  message?: string;
}

export interface OutgoingMessage {
  type: 'message';
  conversation_id?: string;
  group_id?: string;
  content: string;
}

// API Error
export interface ApiError {
  error: string;
}
