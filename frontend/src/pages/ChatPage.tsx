import { useEffect, useState } from 'react';
import {
  Button,
  Avatar,
  ScrollShadow,
  Badge,
} from '@heroui/react';
import { apiClient } from '../services/api';
import { useChatStore } from '../stores/chatStore';
import { useAuthStore } from '../stores/authStore';
import { useWebSocket } from '../hooks/useWebSocket';
import { ChatWindow, UserSearch, CreateGroupModal } from '../components/chat';

export function ChatPage() {
  const {
    conversations,
    groups,
    chatId,
    chatType,
    setConversations,
    setGroups,
    selectConversation,
    selectGroup,
    users,
    onlineUserIds,
  } = useChatStore();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);
  const { isConnected } = useWebSocket();

  const [isUserSearchOpen, setIsUserSearchOpen] = useState(false);
  const [isCreateGroupOpen, setIsCreateGroupOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<'chats' | 'groups'>('chats');

  useEffect(() => {
    const loadData = async () => {
      try {
        const [convs, grps, usersData] = await Promise.all([
          apiClient.getConversations(),
          apiClient.getGroups(),
          apiClient.listUsers(),
        ]);
        setConversations(convs || []);
        setGroups(grps || []);
        useChatStore.setState({ users: usersData || [] });
      } catch {}
    };
    loadData();
  }, [setConversations, setGroups]);

  const getUserInitials = (name: string) =>
    name ? name.substring(0, 2).toUpperCase() : '??';

  const connected = isConnected();

  return (
    <div className="flex h-screen bg-background overflow-hidden">
      {/* Sidebar */}
      <div className="w-72 flex flex-col border-r border-default-100 bg-content1 flex-shrink-0">
        {/* Header */}
        <div className="px-4 py-4 border-b border-default-100">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
                <svg className="w-4 h-4 text-primary-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <span className="font-bold text-foreground text-lg">ViralLens</span>
            </div>
          </div>

          {/* User info */}
          <div className="flex items-center gap-3 p-2 rounded-xl bg-default-50">
            <Badge
              content=""
              color={connected ? 'success' : 'default'}
              size="sm"
              placement="bottom-right"
              shape="circle"
            >
              <Avatar
                name={getUserInitials(user?.username || '')}
                size="sm"
                classNames={{
                  base: 'bg-gradient-to-br from-primary-500 to-secondary-500',
                  name: 'text-white text-xs font-semibold',
                }}
              />
            </Badge>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-foreground truncate">{user?.username}</p>
              <p className="text-xs text-default-400 truncate">{connected ? 'Online' : 'Connecting...'}</p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="px-4 py-3 flex gap-2">
          <Button
            onPress={() => { setIsUserSearchOpen(true); setIsCreateGroupOpen(false); }}
            size="sm"
            color="primary"
            variant="flat"
            className="flex-1 font-medium"
            startContent={
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            }
          >
            New Chat
          </Button>
          <Button
            onPress={() => { setIsCreateGroupOpen(true); setIsUserSearchOpen(false); }}
            size="sm"
            color="secondary"
            variant="flat"
            className="flex-1 font-medium"
            startContent={
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            }
          >
            New Group
          </Button>
        </div>

        {/* Tabs */}
        <div className="px-4 pb-2">
          <div className="flex gap-1 p-1 rounded-xl bg-default-100">
            <button
              onClick={() => setActiveTab('chats')}
              className={`flex-1 text-xs font-medium py-1.5 rounded-lg transition-all ${
                activeTab === 'chats'
                  ? 'bg-content1 text-foreground shadow-sm'
                  : 'text-default-500 hover:text-foreground'
              }`}
            >
              Chats {(conversations || []).length > 0 && `(${(conversations || []).length})`}
            </button>
            <button
              onClick={() => setActiveTab('groups')}
              className={`flex-1 text-xs font-medium py-1.5 rounded-lg transition-all ${
                activeTab === 'groups'
                  ? 'bg-content1 text-foreground shadow-sm'
                  : 'text-default-500 hover:text-foreground'
              }`}
            >
              Groups {(groups || []).length > 0 && `(${(groups || []).length})`}
            </button>
          </div>
        </div>

        {/* List */}
        <ScrollShadow className="flex-1 px-2 pb-4">
          {activeTab === 'chats' ? (
            (conversations || []).length === 0 ? (
              <div className="flex flex-col items-center justify-center h-40 text-center px-4">
                <svg className="w-10 h-10 text-default-200 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <p className="text-default-400 text-xs">No chats yet. Start one!</p>
              </div>
            ) : (
              <div className="space-y-0.5 mt-1">
                {(conversations || []).map((conv) => {
                  const otherUserId = conv.participants.find((p) => p !== user?.id);
                  const otherUser = users.find((u) => u.id === otherUserId);
                  const chatName = otherUser?.username || 'Unknown User';

                    const isOtherOnline = otherUserId ? onlineUserIds.has(otherUserId) : false;

                  return (
                    <button
                      key={conv.id}
                      onClick={() => selectConversation(conv)}
                      className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all text-left ${
                        chatType === 'conversation' && chatId === conv.id
                          ? 'bg-primary-100 text-primary-700'
                          : 'hover:bg-default-100 text-foreground'
                      }`}
                    >
                      <Badge
                        content=""
                        color={isOtherOnline ? 'success' : 'default'}
                        size="sm"
                        placement="bottom-right"
                        shape="circle"
                        isInvisible={!isOtherOnline}
                      >
                        <Avatar
                          name={chatName.substring(0, 2).toUpperCase()}
                          size="sm"
                          classNames={{
                            base: 'bg-gradient-to-br from-blue-500 to-cyan-500 flex-shrink-0',
                            name: 'text-white text-xs font-semibold',
                          }}
                        />
                      </Badge>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1.5">
                          <p className="text-sm font-medium truncate">{chatName}</p>
                        </div>
                        <p className="text-xs text-default-400 truncate">
                          {new Date(conv.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </button>
                  );
                })}
              </div>
            )
          ) : (
            (groups || []).length === 0 ? (
              <div className="flex flex-col items-center justify-center h-40 text-center px-4">
                <svg className="w-10 h-10 text-default-200 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <p className="text-default-400 text-xs">No groups yet. Create one!</p>
              </div>
            ) : (
              <div className="space-y-0.5 mt-1">
                {(groups || []).map((group) => (
                  <button
                    key={group.id}
                    onClick={() => selectGroup(group)}
                    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all text-left ${
                      chatType === 'group' && chatId === group.id
                        ? 'bg-secondary-100 text-secondary-700'
                        : 'hover:bg-default-100 text-foreground'
                    }`}
                  >
                    <Avatar
                      name={group.name.substring(0, 2).toUpperCase()}
                      size="sm"
                      classNames={{
                        base: 'bg-gradient-to-br from-secondary-500 to-pink-500 flex-shrink-0',
                        name: 'text-white text-xs font-semibold',
                      }}
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{group.name}</p>
                      <p className="text-xs text-default-400 truncate">{group.members.length} members</p>
                    </div>
                  </button>
                ))}
              </div>
            )
          )}
        </ScrollShadow>

        <div className="p-4 border-t border-default-100 mt-auto">
          <Button
            onPress={logout}
            color="danger"
            variant="solid"
            className="w-full font-medium"
          >
            Log out
          </Button>
        </div>
      </div>

      {/* Main area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {isUserSearchOpen ? (
          <div className="h-full overflow-y-auto bg-background">
            <div className="max-w-2xl mx-auto p-6">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-xl font-bold text-foreground">New Conversation</h2>
                  <p className="text-default-400 text-sm mt-0.5">Search for a user to start chatting</p>
                </div>
                <Button
                  onPress={() => setIsUserSearchOpen(false)}
                  variant="flat"
                  size="sm"
                  color="default"
                  startContent={
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  }
                >
                  Cancel
                </Button>
              </div>
              <UserSearch onConversationCreated={() => { setIsUserSearchOpen(false); setActiveTab('chats'); }} />
            </div>
          </div>
        ) : chatId && chatType ? (
          <ChatWindow />
        ) : (
          <div className="flex items-center justify-center h-full bg-background">
            <div className="text-center max-w-xs">
              <div className="w-20 h-20 rounded-3xl bg-default-100 flex items-center justify-center mx-auto mb-4">
                <svg className="w-10 h-10 text-default-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-foreground mb-2">Your messages</h3>
              <p className="text-default-400 text-sm">Select a conversation from the sidebar or start a new one</p>
            </div>
          </div>
        )}
      </div>

      {/* Modals */}
      <CreateGroupModal
        isOpen={isCreateGroupOpen}
        onClose={() => setIsCreateGroupOpen(false)}
      />
    </div>
  );
}
