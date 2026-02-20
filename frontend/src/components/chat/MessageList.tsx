import React, { useEffect, useRef, useState, useCallback, useLayoutEffect } from 'react';
import { Spinner, Avatar, Button } from '@heroui/react';
import { useAuthStore } from '../../stores/authStore';
import { useChatStore } from '../../stores/chatStore';
import { apiClient } from '../../services/api';
import type { Message } from '../../types';

// Stable throttle that survives across renders when stored in a ref
function createThrottle<T extends (...args: any[]) => any>(func: T, delay: number) {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;
  let lastRan = 0;
  return (...args: Parameters<T>) => {
    const now = Date.now();
    if (now - lastRan >= delay) {
      func(...args);
      lastRan = now;
    } else {
      if (timeoutId) clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        func(...args);
        lastRan = Date.now();
      }, delay - (now - lastRan));
    }
  };
}

interface MessageListProps {
  className?: string;
}

export const MessageList: React.FC<MessageListProps> = ({ className = '' }) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMoreMessages, setHasMoreMessages] = useState(true);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const lastMessageCountRef = useRef(0);
  // Track cursors already fetched to prevent fetching the same page twice
  const fetchedCursorsRef = useRef<Set<string>>(new Set());
  const isLoadingMoreRef = useRef(false);
  const throttledScrollHandlerRef = useRef<((...args: any[]) => void) | null>(null);
  const prevScrollHeightRef = useRef<number>(0);
  const shouldAdjustScrollRef = useRef(false);

  const user = useAuthStore((state) => state.user);
  const { chatType, chatId, messages: allMessages, prependMessages, setMessages, users } = useChatStore();

  const currentMessages = chatId ? allMessages[chatId] || [] : [];
  const currentMessagesCursor = currentMessages.length > 0 ? currentMessages[0].created_at : undefined;



  const loadMoreMessages = useCallback(async () => {
    if (!chatId || !chatType || isLoadingMoreRef.current || !hasMoreMessages) return;
    // Prevent fetching the same cursor twice
    const cursor = currentMessagesCursor ?? '__start__';
    if (fetchedCursorsRef.current.has(cursor)) return;
    fetchedCursorsRef.current.add(cursor);
    isLoadingMoreRef.current = true;
    setIsLoadingMore(true);
    setError(null);
    try {
      if (scrollRef.current) {
        prevScrollHeightRef.current = scrollRef.current.scrollHeight;
        shouldAdjustScrollRef.current = true;
      }
      let newMessages: Message[] = [];
      if (chatType === 'conversation')
        newMessages = await apiClient.getConversationMessages(chatId, currentMessagesCursor, 50);
      else if (chatType === 'group')
        newMessages = await apiClient.getGroupMessages(chatId, currentMessagesCursor, 50);
      if (newMessages.length < 50) setHasMoreMessages(false);
      if (newMessages.length > 0) {
        // Dedup before prepending
        const existingIds = new Set((chatId ? (useChatStore.getState().messages[chatId] || []) : []).map((m: Message) => m.id));
        const unique = newMessages.filter((m) => !existingIds.has(m.id));
        if (unique.length > 0) {
          prependMessages(chatId, unique);
        } else {
          shouldAdjustScrollRef.current = false;
        }
      } else {
        shouldAdjustScrollRef.current = false;
      }
    } catch {
      setError('Failed to load more messages.');
      shouldAdjustScrollRef.current = false;
      // Remove cursor from fetched set so user can retry
      fetchedCursorsRef.current.delete(cursor);
    } finally {
      isLoadingMoreRef.current = false;
      setIsLoadingMore(false);
    }
  }, [chatId, chatType, currentMessagesCursor, hasMoreMessages, prependMessages]);

  const scrollToBottom = useCallback((smooth = false) => {
    if (scrollRef.current) {
      scrollRef.current.scrollTo({
        top: scrollRef.current.scrollHeight,
        behavior: smooth ? 'smooth' : 'auto',
      });
    }
  }, []);

  // Build the throttled scroll handler
  useEffect(() => {
    throttledScrollHandlerRef.current = createThrottle(() => {
      if (!scrollRef.current || isLoadingMoreRef.current || !hasMoreMessages) return;
      if (scrollRef.current.scrollTop < 250) {
        loadMoreMessages();
      }
    }, 200);
  }, [hasMoreMessages, loadMoreMessages]);

  const handleScroll = useCallback(() => {
    throttledScrollHandlerRef.current?.();
  }, []);

  // Handle scroll adjustment after prepending messages
  useLayoutEffect(() => {
    if (shouldAdjustScrollRef.current && scrollRef.current && prevScrollHeightRef.current > 0) {
      const heightDiff = scrollRef.current.scrollHeight - prevScrollHeightRef.current;
      if (heightDiff > 0) {
        scrollRef.current.scrollTop += heightDiff;
      }
      shouldAdjustScrollRef.current = false;
      prevScrollHeightRef.current = 0;
    }
  }, [currentMessages]);

  // Only run when chatId/chatType changes — NOT on message count change
  useEffect(() => {
    if (!chatId || !chatType) return;
    let isAborted = false;
    // Reset per-chat state
    setIsInitialLoad(true);
    setHasMoreMessages(true);
    setError(null);
    lastMessageCountRef.current = 0;
    fetchedCursorsRef.current = new Set();
    isLoadingMoreRef.current = false;
    shouldAdjustScrollRef.current = false;
    prevScrollHeightRef.current = 0;

    const existingMessages = useChatStore.getState().messages[chatId];
    if (!existingMessages || existingMessages.length === 0) {
      const loadInitial = async () => {
        try {
          let messages: Message[] = [];
          if (chatType === 'conversation')
            messages = await apiClient.getConversationMessages(chatId, undefined, 50);
          else if (chatType === 'group')
            messages = await apiClient.getGroupMessages(chatId, undefined, 50);
          if (!isAborted) {
            if (messages.length < 50) setHasMoreMessages(false);
            // Mark this initial cursor as fetched
            fetchedCursorsRef.current.add('__start__');
            setMessages(chatId, messages);
          }
        } catch {
          if (!isAborted) setError('Failed to load messages.');
        } finally {
          if (!isAborted) setIsInitialLoad(false);
        }
      };
      loadInitial();
    } else {
      setIsInitialLoad(false);
    }

    return () => {
      isAborted = true;
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chatId, chatType, setMessages]);

  useEffect(() => {
    if (isInitialLoad) return;
    const count = currentMessages.length;
    const prev = lastMessageCountRef.current;
    if (count > 0) {
      if (prev === 0) {
        scrollToBottom(false);
      } else if (count > prev) {
        const lastMsg = currentMessages[count - 1];
        const isMine = user && lastMsg.sender_id === user.id;
        if (isMine || (scrollRef.current &&
          scrollRef.current.scrollHeight - scrollRef.current.scrollTop - scrollRef.current.clientHeight < 200)) {
          scrollToBottom(true);
        }
      }
    }
    lastMessageCountRef.current = count;
  }, [currentMessages, isInitialLoad, user, scrollToBottom]);

  const formatTime = (ts: string) =>
    new Date(ts).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });

  const formatDate = (ts: string) => {
    const date = new Date(ts);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    const d = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
    if (d(date) === d(today)) return 'Today';
    if (d(date) === d(yesterday)) return 'Yesterday';
    return date.toLocaleDateString('en-US', { month: 'long', day: 'numeric' });
  };

  const shouldShowDate = (msg: Message, prev?: Message) => {
    if (!prev) return true;
    const a = new Date(msg.created_at), b = new Date(prev.created_at);
    return a.getDate() !== b.getDate() || a.getMonth() !== b.getMonth() || a.getFullYear() !== b.getFullYear();
  };

  const getInitials = (username: string) => {
    return username.charAt(0).toUpperCase();
  };

  const avatarColors = [
    'from-blue-500 to-cyan-500',
    'from-purple-500 to-pink-500',
    'from-green-500 to-emerald-500',
    'from-orange-500 to-red-500',
    'from-indigo-500 to-blue-500',
  ];
  const getColor = (id: string) => {
    const hash = id.split('').reduce((a, c) => a + c.charCodeAt(0), 0);
    return avatarColors[hash % avatarColors.length];
  };

  if (isInitialLoad) {
    return (
      <div className={`flex items-center justify-center h-full ${className}`}>
        <Spinner size="lg" color="primary" label="Loading messages…" />
      </div>
    );
  }

  if (error && currentMessages.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center h-full ${className}`}>
        <div className="text-center p-8">
          <div className="w-14 h-14 rounded-2xl bg-danger-50 flex items-center justify-center mx-auto mb-4">
            <svg className="w-7 h-7 text-danger" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p className="text-danger text-sm mb-4">{error}</p>
          <Button color="primary" size="sm" onPress={() => { setError(null); setIsInitialLoad(true); }}>
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  if (currentMessages.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center h-full ${className}`}>
        <div className="text-center p-8">
          <div className="w-14 h-14 rounded-2xl bg-default-100 flex items-center justify-center mx-auto mb-4">
            <svg className="w-7 h-7 text-default-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p className="text-default-500 font-medium mb-1">No messages yet</p>
          <p className="text-default-400 text-sm">Send the first message below</p>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={scrollRef}
      onScroll={handleScroll}
      className={`flex flex-col overflow-y-auto px-4 py-4 space-y-1 ${className}`}
      role="log"
      aria-live="polite"
    >
      {error && currentMessages.length > 0 && (
        <div className="flex justify-center py-2">
          <div className="bg-danger-50 text-danger border border-danger-200 rounded-xl px-4 py-2 text-xs">
            {error}
          </div>
        </div>
      )}

      {isLoadingMore && (
        <div className="flex justify-center py-3">
          <Spinner size="sm" color="primary" />
        </div>
      )}

      {currentMessages.map((message, index) => {
        const isMe = user && message.sender_id === user.id;
        const prev = index > 0 ? currentMessages[index - 1] : undefined;
        const showDate = shouldShowDate(message, prev);
        const showAvatar = !isMe && (index === 0 || currentMessages[index - 1]?.sender_id !== message.sender_id);

        return (
          <React.Fragment key={message.id}>
            {showDate && (
              <div className="flex items-center justify-center my-3">
                <div className="bg-default-100 text-default-500 text-xs px-3 py-1 rounded-full">
                  {formatDate(message.created_at)}
                </div>
              </div>
            )}

            <div className={`flex ${isMe ? 'justify-end' : 'justify-start'} items-end gap-2`}>
              {/* Avatar placeholder to keep spacing consistent */}
              {!isMe && (
                <div className="w-7 flex-shrink-0">
                  {showAvatar && (
                    <Avatar
                      size="sm"
                      name={getInitials(users.find(u => u.id === message.sender_id)?.username || message.sender_id)}
                      classNames={{
                        base: `w-7 h-7 bg-gradient-to-br ${getColor(message.sender_id)}`,
                        name: 'text-white text-xs font-semibold',
                      }}
                    />
                  )}
                </div>
              )}

              {/* Bubble */}
              <div className={`max-w-[65%] ${isMe ? 'items-end' : 'items-start'} flex flex-col gap-0.5`}>
                {!isMe && showAvatar && chatType === 'group' && (
                  <span className="text-xs text-default-400 px-1 mb-0.5">
                    User {message.sender_id.substring(0, 8)}
                  </span>
                )}
                <div
                  className={`px-3.5 py-2 rounded-2xl ${
                    isMe
                      ? 'bg-primary text-primary-foreground rounded-br-sm'
                      : 'bg-content2 text-foreground border border-default-100 rounded-bl-sm'
                  }`}
                >
                  <p className="text-sm break-words whitespace-pre-wrap leading-relaxed">{message.content}</p>
                  <p className={`text-xs mt-1 text-right ${isMe ? 'text-primary-foreground/60' : 'text-default-400'}`}>
                    {formatTime(message.created_at)}
                  </p>
                </div>
              </div>
            </div>
          </React.Fragment>
        );
      })}
    </div>
  );
};
