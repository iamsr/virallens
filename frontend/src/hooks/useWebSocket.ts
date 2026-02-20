import { useEffect, useState, useCallback } from 'react';
import { wsClient } from '../services/websocket';
import { useAuthStore } from '../stores/authStore';
import { useChatStore } from '../stores/chatStore';

export function useWebSocket() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const accessToken = useAuthStore((state) => state.accessToken);
  const { addMessage } = useChatStore();

  // Create reactive state for UI updates
  const [connected, setConnected] = useState(() => wsClient.isConnected());

  useEffect(() => {
    if (!isAuthenticated || !accessToken) {
      wsClient.disconnect();
      setConnected(false);
      return;
    }

    // Subscribe BEFORE connecting so we never miss the onopen event
    const unsubConnect = wsClient.onConnect(() => setConnected(true));
    const unsubDisconnect = wsClient.onDisconnect(() => setConnected(false));

    // Handle incoming messages
    const unsubscribeMessage = wsClient.onMessage((message) => {
      const chatId = message.conversation_id || message.group_id;
      if (chatId) {
        addMessage(chatId, message);
      }
    });

    const unsubscribeError = wsClient.onError((error) => {
      console.error('WebSocket error:', error);
    });

    // Handle presence updates (single user online/offline)
    const unsubPresence = wsClient.onPresence((userId, status) => {
      if (status === 'online') {
        useChatStore.getState().setUserOnline(userId);
      } else {
        useChatStore.getState().setUserOffline(userId);
      }
    });

    // Handle initial list of online users on connect
    const unsubPresenceList = wsClient.onPresenceList((userIds) => {
      useChatStore.getState().setOnlineUsers(userIds);
    });

    // Connect if not already connected
    if (!wsClient.isConnected()) {
      wsClient.connect();
    } else {
      // Already connected — sync state
      setConnected(true);
    }

    // Cleanup handlers on unmount/re-run
    return () => {
      unsubscribeMessage();
      unsubscribeError();
      unsubConnect();
      unsubDisconnect();
      unsubPresence();
      unsubPresenceList();
    };
  }, [isAuthenticated, accessToken, addMessage]);

  const sendMessage = useCallback((msg: object) => {
    wsClient.sendMessage(msg as any);
  }, []);

  const isConnected = useCallback(() => connected, [connected]);

  return { sendMessage, isConnected };
}
