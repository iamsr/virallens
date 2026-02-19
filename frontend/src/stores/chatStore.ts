import { create } from 'zustand';
import type { Conversation, Group, Message } from '../types';

interface ChatState {
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
  
  // Actions - UI
  openChat: (type: 'conversation' | 'group', id: string) => void;
  closeChat: () => void;
  
  // Actions - Reset
  reset: () => void;
}

export const useChatStore = create<ChatState>((set) => ({
  // Initial state
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
      conversations: [conversation, ...state.conversations],
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
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [...(state.messages[chatId] || []), message],
      },
    })),

  prependMessages: (chatId, messages) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [...messages, ...(state.messages[chatId] || [])],
      },
    })),

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
