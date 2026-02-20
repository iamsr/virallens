import React, { useMemo } from 'react';
import { Avatar, Divider, Chip } from '@heroui/react';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';
import { useChatStore } from '../../stores/chatStore';
import { useAuthStore } from '../../stores/authStore';

export const ChatWindow: React.FC = () => {
  const { chatType, chatId, selectedConversation, selectedGroup, users, onlineUserIds } = useChatStore();
  const currentUser = useAuthStore((state) => state.user);

  const chatTitle = useMemo(() => {
    if (!chatId || !chatType) return null;
    if (chatType === 'conversation' && selectedConversation) {
      const otherId = selectedConversation.participants.find((p) => p !== currentUser?.id) || 'Unknown';
      const userObj = users.find((u) => u.id === otherId);
      return userObj?.username || `User ${otherId.substring(0, 8)}`;
    } else if (chatType === 'group' && selectedGroup) {
      return selectedGroup.name;
    }
    return 'Chat';
  }, [chatId, chatType, selectedConversation, selectedGroup, users, currentUser]);

  const chatMetadata = useMemo(() => {
    if (!chatId || !chatType) return null;
    if (chatType === 'conversation' && selectedConversation) {
      const otherId = selectedConversation.participants.find((p) => p !== currentUser?.id);
      const isOnline = otherId ? onlineUserIds.has(otherId) : false;
      return isOnline ? '🟢 Online' : 'Direct Message';
    }
    if (chatType === 'group' && selectedGroup) {
      const count = selectedGroup.members.length;
      return `${count} ${count === 1 ? 'member' : 'members'}`;
    }
    return null;
  }, [chatId, chatType, selectedConversation, selectedGroup, currentUser, onlineUserIds]);

  if (!chatId || !chatType) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-background">
        <div className="w-20 h-20 rounded-3xl bg-default-100 flex items-center justify-center mb-4">
          <svg className="w-10 h-10 text-default-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </div>
        <h2 className="text-lg font-semibold text-foreground mb-1">Select a conversation</h2>
        <p className="text-default-400 text-sm">Choose from the sidebar to start chatting</p>
      </div>
    );
  }

  const avatarName = chatType === 'group'
    ? (selectedGroup?.name || 'G').substring(0, 2).toUpperCase()
    : (chatTitle && chatTitle !== 'Chat' ? chatTitle.substring(0, 2).toUpperCase() : 'DM');

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Header */}
      <div className="flex items-center gap-3 px-5 py-3 border-b border-default-100 bg-content1/50 backdrop-blur-sm">
        <Avatar
          name={avatarName}
          size="sm"
          classNames={{
            base: `bg-gradient-to-br ${chatType === 'group' ? 'from-secondary-500 to-pink-500' : 'from-blue-500 to-cyan-500'}`,
            name: 'text-white text-xs font-semibold',
          }}
        />
        <div className="flex-1">
          <h2 className="text-sm font-semibold text-foreground leading-none">{chatTitle}</h2>
          {chatMetadata && (
            <p className="text-xs text-default-400 mt-0.5">{chatMetadata}</p>
          )}
        </div>
        <Chip
          size="sm"
          variant="flat"
          color={chatType === 'group' ? 'secondary' : 'primary'}
          className="text-xs"
        >
          {chatType === 'group' ? 'Group' : 'DM'}
        </Chip>
      </div>

      <Divider />

      {/* Messages */}
      <div className="flex-1 overflow-hidden">
        <MessageList className="h-full" />
      </div>

      {/* Input */}
      <MessageInput />
    </div>
  );
};
