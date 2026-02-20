import axios, { AxiosInstance, AxiosError } from 'axios';
import type {
  User,
  Conversation,
  Group,
  Message,
  AuthResponse,
  RegisterRequest,
  LoginRequest,
  ApiError
} from '../types';
import { useAuthStore } from '../stores/authStore';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;
  private refreshTokenPromise: Promise<string> | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = useAuthStore.getState().accessToken;
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor to handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError<ApiError>) => {
        const originalRequest = error.config;

        // If error is 401 and we haven't tried to refresh yet
        if (error.response?.status === 401 && originalRequest && !(originalRequest as any)._retry) {
          (originalRequest as any)._retry = true;

          try {
            // Prevent multiple refresh requests
            if (!this.refreshTokenPromise) {
              this.refreshTokenPromise = this.refreshAccessToken();
            }

            const newToken = await this.refreshTokenPromise;
            this.refreshTokenPromise = null;

            // Retry original request with new token
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            this.refreshTokenPromise = null;
            // Token refresh failed, logout user
            useAuthStore.getState().logout();
            window.location.href = '/login';
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      }
    );
  }

  // Removed getAccessToken and getRefreshToken logic to defer securely to Zustand

  private async refreshAccessToken(): Promise<string> {
    const refreshToken = useAuthStore.getState().refreshToken;
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await axios.post<{ access_token: string }>(
      `${API_BASE_URL}/api/auth/refresh`,
      { refresh_token: refreshToken }
    );

    const newToken = response.data.access_token;

    // Update token in store
    // Update token in store
    useAuthStore.getState().setAccessToken(newToken);

    return newToken;
  }

  // Auth endpoints
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/auth/register', data);
    return response.data;
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/auth/login', data);
    return response.data;
  }

  async logout(): Promise<void> {
    await this.client.post('/api/auth/logout');
    useAuthStore.getState().logout();
  }

  // User endpoints
  async listUsers(): Promise<User[]> {
    const response = await this.client.get<User[]>('/api/users');
    return response.data;
  }

  // Conversation endpoints
  async getConversations(): Promise<Conversation[]> {
    const response = await this.client.get<Conversation[]>('/api/conversations');
    return response.data;
  }

  async createOrGetConversation(otherUserId: string): Promise<Conversation> {
    const response = await this.client.post<Conversation>('/api/conversations', {
      other_user_id: otherUserId,
    });
    return response.data;
  }

  async getConversation(id: string): Promise<Conversation> {
    const response = await this.client.get<Conversation>(`/api/conversations/${id}`);
    return response.data;
  }

  async getConversationMessages(
    id: string,
    cursor?: string,
    limit: number = 50
  ): Promise<Message[]> {
    const params: any = { limit };
    if (cursor) params.cursor = cursor;

    const response = await this.client.get<Message[]>(
      `/api/conversations/${id}/messages`,
      { params }
    );
    // Backend returns descending (newest first), reverse to get chronological
    return response.data.reverse();
  }

  // Group endpoints
  async getGroups(): Promise<Group[]> {
    const response = await this.client.get<Group[]>('/api/groups');
    return response.data;
  }

  async createGroup(name: string, members: string[]): Promise<Group> {
    const response = await this.client.post<Group>('/api/groups', {
      name,
      members,
    });
    return response.data;
  }

  async getGroup(id: string): Promise<Group> {
    const response = await this.client.get<Group>(`/api/groups/${id}`);
    return response.data;
  }

  async addGroupMember(groupId: string, userId: string): Promise<void> {
    await this.client.post(`/api/groups/${groupId}/members`, {
      user_id: userId,
    });
  }

  async removeGroupMember(groupId: string, userId: string): Promise<void> {
    await this.client.delete(`/api/groups/${groupId}/members`, {
      data: { user_id: userId },
    });
  }

  async getGroupMessages(
    id: string,
    cursor?: string,
    limit: number = 50
  ): Promise<Message[]> {
    const params: any = { limit };
    if (cursor) params.cursor = cursor;

    const response = await this.client.get<Message[]>(
      `/api/groups/${id}/messages`,
      { params }
    );
    // Backend returns descending (newest first), reverse to get chronological
    return response.data.reverse();
  }
}

export const apiClient = new ApiClient();
