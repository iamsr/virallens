import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Input, Card, CardBody, Avatar, Button, Spinner } from '@heroui/react';
import { apiClient } from '../../services/api';
import { useAuthStore } from '../../stores/authStore';
import { useChatStore } from '../../stores/chatStore';
import type { User } from '../../types';

interface UserSearchProps {
  className?: string;
  onConversationCreated?: (conversationId: string) => void;
}

export const UserSearch: React.FC<UserSearchProps> = ({ className = '', onConversationCreated }) => {
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const currentUser = useAuthStore((state) => state.user);
  const { selectConversation, addConversation } = useChatStore();

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(searchTerm), 300);
    return () => clearTimeout(t);
  }, [searchTerm]);

  useEffect(() => {
    apiClient.listUsers()
      .then((u) => setAllUsers(u))
      .catch(() => setError('Failed to load users.'))
      .finally(() => setIsLoading(false));
  }, []);

  const filteredUsers = useMemo(() => {
    if (!currentUser) return [];
    let users = allUsers.filter((u) => u.id !== currentUser.id);
    if (!debouncedSearch.trim()) return [];
    const q = debouncedSearch.toLowerCase().trim();
    return users.filter((u) => u.username.toLowerCase().includes(q) || u.email.toLowerCase().includes(q));
  }, [allUsers, debouncedSearch, currentUser]);

  const handleStartConversation = useCallback(async (userId: string) => {
    setIsCreating(userId);
    setError(null);
    try {
      const conversation = await apiClient.createOrGetConversation(userId);
      addConversation(conversation);
      selectConversation(conversation);
      onConversationCreated?.(conversation.id);
    } catch {
      setError('Failed to start conversation.');
    } finally {
      setIsCreating(null);
    }
  }, [addConversation, selectConversation, onConversationCreated]);

  const getInitials = (name: string) => {
    const parts = name.split(/[\s_-]+/);
    return parts.length >= 2 ? (parts[0][0] + parts[1][0]).toUpperCase() : name.substring(0, 2).toUpperCase();
  };

  const avatarColors = ['from-blue-500 to-cyan-500', 'from-purple-500 to-pink-500', 'from-green-500 to-emerald-500', 'from-orange-500 to-red-500', 'from-indigo-500 to-blue-500'];
  const getColor = (id: string) => avatarColors[id.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % avatarColors.length];

  return (
    <div className={`flex flex-col gap-4 ${className}`}>
      <Input
        placeholder="Search by username or email…"
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        size="lg"
        variant="bordered"
        classNames={{
          inputWrapper: 'border-default-200 hover:border-default-400 focus-within:!border-primary bg-default-50',
        }}
        startContent={
          <svg className="w-4 h-4 text-default-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        }
        autoComplete="off"
        autoFocus
      />

      {isLoading && (
        <div className="flex justify-center py-12">
          <Spinner color="primary" label="Loading users…" />
        </div>
      )}

      {error && !isLoading && (
        <div className="flex flex-col items-center py-12 gap-3">
          <p className="text-danger text-sm">{error}</p>
          <Button size="sm" color="primary" onPress={() => { setError(null); setIsLoading(true); apiClient.listUsers().then(setAllUsers).catch(() => setError('Failed.')).finally(() => setIsLoading(false)); }}>
            Retry
          </Button>
        </div>
      )}

      {!isLoading && !error && searchTerm.trim() === '' && (
        <div className="flex flex-col items-center py-12 text-center">
          <div className="w-14 h-14 rounded-2xl bg-default-100 flex items-center justify-center mb-3">
            <svg className="w-7 h-7 text-default-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <p className="text-default-500 font-medium text-sm">Find someone to chat with</p>
          <p className="text-default-400 text-xs mt-1">Type a username or email above</p>
        </div>
      )}

      {!isLoading && !error && debouncedSearch.trim() !== '' && filteredUsers.length === 0 && (
        <div className="flex flex-col items-center py-12 text-center">
          <div className="w-14 h-14 rounded-2xl bg-default-100 flex items-center justify-center mb-3">
            <svg className="w-7 h-7 text-default-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          </div>
          <p className="text-default-500 font-medium text-sm">No users found</p>
          <p className="text-default-400 text-xs mt-1">Try a different search term</p>
        </div>
      )}

      {!isLoading && filteredUsers.length > 0 && (
        <div className="flex flex-col gap-2">
          {filteredUsers.map((u) => (
              <Card
                key={u.id}
                className="border border-default-100 hover:border-primary/30 transition-all cursor-pointer"
                classNames={{ base: 'bg-content1 hover:bg-content2' }}
                shadow="none"
                isPressable
                onPress={() => handleStartConversation(u.id)}
              >
                <CardBody className="flex flex-row items-center gap-3 p-3">
                <Avatar
                  name={getInitials(u.username)}
                  size="md"
                  classNames={{
                    base: `bg-gradient-to-br ${getColor(u.id)} flex-shrink-0`,
                    name: 'text-white font-semibold',
                  }}
                />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-foreground">{u.username}</p>
                  <p className="text-xs text-default-400 truncate">{u.email}</p>
                </div>
                <Button
                  size="sm"
                  color="primary"
                  variant="flat"
                  isLoading={isCreating === u.id}
                  onPress={() => handleStartConversation(u.id)}
                  className="flex-shrink-0 font-medium"
                >
                  {isCreating === u.id ? 'Opening…' : 'Chat'}
                </Button>
              </CardBody>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};
