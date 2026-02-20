import { create } from 'zustand';
import type { Conversation, Group, Message, User } from '../types';

interface ChatState {
  // Users Cache
  users: User[];

  // Online presence: set of user IDs that are currently online
  onlineUserIds: Set<string>;

  // Conversations
  conversations: Conversation[];
  selectedConversation: Conversation | null;

  // Groups
  groups: Group[];
  selectedGroup: Group | null;

  // Messages
  messages: Record<string, Message[]>; // key: conversation_id or group_id

  // UI state
  isChatOpen: boolean;
  chatType: 'conversation' | 'group' | null;
  chatId: string | null;

  // Actions - Conversations
  setConversations: (conversations: Conversation[]) => void;
  addConversation: (conversation: Conversation) => void;
  selectConversation: (conversation: Conversation) => void;

  // Actions - Groups
  setGroups: (groups: Group[]) => void;
  addGroup: (group: Group) => void;
  selectGroup: (group: Group) => void;
  updateGroup: (groupId: string, updates: Partial<Group>) => void;

  // Actions - Messages
  setMessages: (chatId: string, messages: Message[]) => void;
  addMessage: (chatId: string, message: Message) => void;
  prependMessages: (chatId: string, messages: Message[]) => void;

  // Actions - Presence
  setOnlineUsers: (userIds: string[]) => void;
  setUserOnline: (userId: string) => void;
  setUserOffline: (userId: string) => void;

  // Actions - UI
  openChat: (type: 'conversation' | 'group', id: string) => void;
  closeChat: () => void;

  // Actions - Reset
  reset: () => void;
}

export const useChatStore = create<ChatState>((set) => ({
  // Initial state
  users: [],
  onlineUserIds: new Set<string>(),
  conversations: [],
  selectedConversation: null,
  groups: [],
  selectedGroup: null,
  messages: {},
  isChatOpen: false,
  chatType: null,
  chatId: null,

  // Conversation actions
  setConversations: (conversations) =>
    set({ conversations }),

  addConversation: (conversation) =>
    set((state) => ({
      // Dedup: move existing to front if already present, otherwise prepend
      conversations: state.conversations.some((c) => c.id === conversation.id)
        ? state.conversations
        : [conversation, ...state.conversations],
    })),

  selectConversation: (conversation) =>
    set({
      selectedConversation: conversation,
      selectedGroup: null,
      isChatOpen: true,
      chatType: 'conversation',
      chatId: conversation.id,
    }),

  // Group actions
  setGroups: (groups) =>
    set({ groups }),

  addGroup: (group) =>
    set((state) => ({
      groups: [group, ...state.groups],
    })),

  selectGroup: (group) =>
    set({
      selectedGroup: group,
      selectedConversation: null,
      isChatOpen: true,
      chatType: 'group',
      chatId: group.id,
    }),

  updateGroup: (groupId, updates) =>
    set((state) => ({
      groups: state.groups.map((g) =>
        g.id === groupId ? { ...g, ...updates } : g
      ),
      selectedGroup:
        state.selectedGroup?.id === groupId
          ? { ...state.selectedGroup, ...updates }
          : state.selectedGroup,
    })),

  // Message actions
  setMessages: (chatId, messages) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: messages,
      },
    })),

  addMessage: (chatId, message) =>
    set((state) => {
      const existing = state.messages[chatId] || [];
      // Deduplicate: don't add if a message with the same ID is already present
      if (existing.some((m) => m.id === message.id)) return state;
      return {
        messages: {
          ...state.messages,
          [chatId]: [...existing, message],
        },
      };
    }),

  prependMessages: (chatId, messages) =>
    set((state) => {
      const existing = state.messages[chatId] || [];
      const existingIds = new Set(existing.map((m) => m.id));
      const unique = messages.filter((m) => !existingIds.has(m.id));
      if (unique.length === 0) return state;
      return {
        messages: {
          ...state.messages,
          [chatId]: [...unique, ...existing],
        },
      };
    }),

  // Presence actions
  setOnlineUsers: (userIds) =>
    set({ onlineUserIds: new Set(userIds) }),

  setUserOnline: (userId) =>
    set((state) => {
      const next = new Set(state.onlineUserIds);
      next.add(userId);
      return { onlineUserIds: next };
    }),

  setUserOffline: (userId) =>
    set((state) => {
      const next = new Set(state.onlineUserIds);
      next.delete(userId);
      return { onlineUserIds: next };
    }),

  // UI actions
  openChat: (type, id) =>
    set({
      isChatOpen: true,
      chatType: type,
      chatId: id,
    }),

  closeChat: () =>
    set({
      isChatOpen: false,
      chatType: null,
      chatId: null,
      selectedConversation: null,
      selectedGroup: null,
    }),

  // Reset
  reset: () =>
    set({
      conversations: [],
      selectedConversation: null,
      groups: [],
      selectedGroup: null,
      messages: {},
      isChatOpen: false,
      chatType: null,
      chatId: null,
    }),
}));
