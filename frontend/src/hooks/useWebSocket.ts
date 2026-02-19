import { useEffect } from 'react';
import { wsClient } from '../services/websocket';
import { useAuthStore } from '../stores/authStore';
import { useChatStore } from '../stores/chatStore';

export function useWebSocket() {
  const { accessToken, isAuthenticated } = useAuthStore();
  const { addMessage } = useChatStore();

  useEffect(() => {
    if (!isAuthenticated || !accessToken) {
      wsClient.disconnect();
      return;
    }

    // Connect WebSocket
    wsClient.connect(accessToken);

    // Handle incoming messages
    const unsubscribeMessage = wsClient.onMessage((message) => {
      // Add message to the appropriate chat
      const chatId = message.conversation_id || message.group_id;
      if (chatId) {
        addMessage(chatId, message);
      }
    });

    const unsubscribeError = wsClient.onError((error) => {
      console.error('WebSocket error:', error);
    });

    // Cleanup on unmount
    return () => {
      unsubscribeMessage();
      unsubscribeError();
    };
  }, [isAuthenticated, accessToken, addMessage]);

  return {
    sendMessage: wsClient.sendMessage.bind(wsClient),
    isConnected: wsClient.isConnected.bind(wsClient),
  };
}
