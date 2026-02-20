import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Textarea, Button, Tooltip } from '@heroui/react';
import { useAuthStore } from '../../stores/authStore';
import { useChatStore } from '../../stores/chatStore';
import { useWebSocket } from '../../hooks/useWebSocket';
import type { OutgoingMessage } from '../../types';

export const MessageInput: React.FC = () => {
  const [message, setMessage] = useState('');
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const user = useAuthStore((state) => state.user);
  const { chatType, chatId } = useChatStore();
  const { sendMessage, isConnected } = useWebSocket();

  const isDisabled = !chatId || !chatType || !user;
  const isWebSocketConnected = isConnected();

  useEffect(() => {
    if (chatId && textareaRef.current) textareaRef.current.focus();
  }, [chatId]);

  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => setError(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [error]);

  const isMessageValid = useCallback((content: string) => content.trim().length > 0, []);

  const handleSendMessage = useCallback(() => {
    if (!chatId || !chatType || !user) {
      setError('No chat selected');
      return;
    }
    if (!isWebSocketConnected) {
      setError('Not connected to server');
      return;
    }
    if (!isMessageValid(message)) return;

    const outgoingMessage: OutgoingMessage = { type: 'message', content: message.trim() };
    if (chatType === 'conversation') outgoingMessage.conversation_id = chatId;
    else if (chatType === 'group') outgoingMessage.group_id = chatId;

    sendMessage(outgoingMessage);
    setMessage('');
    if (textareaRef.current) textareaRef.current.focus();
  }, [chatId, chatType, user, message, isWebSocketConnected, sendMessage, isMessageValid]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSendMessage();
      }
    },
    [handleSendMessage]
  );

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setMessage(e.target.value);
    setError((prev) => (prev ? null : prev));
  }, []);

  return (
    <div className="px-4 py-3 border-t border-default-100 bg-content1/50 backdrop-blur-sm">
      {/* Error / warning banners */}
      {error && (
        <div className="mb-2 flex items-center gap-2 px-3 py-2 rounded-xl bg-danger-50 border border-danger-200 text-danger text-xs" role="alert">
          <svg className="w-3.5 h-3.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
          {error}
        </div>
      )}
      {!isWebSocketConnected && chatId && (
        <div className="mb-2 flex items-center gap-2 px-3 py-2 rounded-xl bg-warning-50 border border-warning-200 text-warning-600 text-xs" role="alert">
          <svg className="w-3.5 h-3.5 flex-shrink-0 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Reconnecting to server…
        </div>
      )}

      {/* Input row */}
      <div className="flex gap-2 items-end">
        <Textarea
          ref={textareaRef}
          value={message}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={
            isDisabled
              ? 'Select a conversation to send messages'
              : 'Message… (Enter to send)'
          }
          disabled={isDisabled || !isWebSocketConnected}
          minRows={1}
          maxRows={5}
          className="flex-1"
          variant="bordered"
          classNames={{
            inputWrapper: 'border-default-200 hover:border-default-400 focus-within:!border-primary bg-default-50',
            input: 'resize-none text-sm',
          }}
          aria-label="Message input"
        />

        <Tooltip content="Send (Enter)" placement="top">
          <Button
            isIconOnly
            color="primary"
            onPress={handleSendMessage}
            isDisabled={isDisabled || !isWebSocketConnected || !isMessageValid(message)}
            className="h-10 w-10 min-w-10 flex-shrink-0"
            aria-label="Send message"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </Button>
        </Tooltip>
      </div>

      {!isDisabled && (
        <p className="text-xs text-default-400 mt-1.5 px-1">
          Shift+Enter for new line
        </p>
      )}
    </div>
  );
};
