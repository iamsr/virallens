import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  Input,
  Chip,
  Avatar,
  Card,
  Spinner,
} from '@heroui/react';
import { apiClient } from '../../services/api';
import { useAuthStore } from '../../stores/authStore';
import { useChatStore } from '../../stores/chatStore';
import type { User } from '../../types';

interface CreateGroupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onGroupCreated?: (groupId: string) => void;
}

export const CreateGroupModal: React.FC<CreateGroupModalProps> = ({
  isOpen,
  onClose,
  onGroupCreated,
}) => {
  // State
  const [groupName, setGroupName] = useState('');
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');
  const [selectedUserIds, setSelectedUserIds] = useState<Set<string>>(new Set());
  const [isLoadingUsers, setIsLoadingUsers] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [creationError, setCreationError] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);

  // Get current user and chat store actions
  const currentUser = useAuthStore((state) => state.user);
  const { addGroup, selectGroup } = useChatStore();

  /**
   * Reset form state when modal is closed
   */
  useEffect(() => {
    if (!isOpen) {
      setGroupName('');
      setSearchTerm('');
      setDebouncedSearchTerm('');
      setSelectedUserIds(new Set());
      setLoadError(null);
      setCreationError(null);
      setValidationError(null);
    }
  }, [isOpen]);

  /**
   * Debounce search term - wait 300ms after user stops typing
   */
  useEffect(() => {
    const timerId = setTimeout(() => {
      setDebouncedSearchTerm(searchTerm);
    }, 300);

    return () => {
      clearTimeout(timerId);
    };
  }, [searchTerm]);

  /**
   * Load all users when modal opens
   */
  useEffect(() => {
    if (!isOpen) return;

    const loadUsers = async () => {
      setIsLoadingUsers(true);
      setLoadError(null);

      try {
        const users = await apiClient.listUsers();
        setAllUsers(users);
      } catch (err) {
        setLoadError('Failed to load users. Please try again.');
      } finally {
        setIsLoadingUsers(false);
      }
    };

    loadUsers();
  }, [isOpen]);

  /**
   * Filter users based on debounced search term
   * Exclude current user and already selected users from results
   */
  const filteredUsers = useMemo(() => {
    if (!currentUser) {
      return [];
    }

    // Filter out current user
    let users = allUsers.filter((user) => user.id !== currentUser.id);

    // If search term is empty, show all users except selected ones
    if (!debouncedSearchTerm.trim()) {
      // Show first 10 unselected users when no search
      return users.filter((user) => !selectedUserIds.has(user.id)).slice(0, 10);
    }

    // Filter by search term (case-insensitive search in username and email)
    const searchLower = debouncedSearchTerm.toLowerCase().trim();
    users = users.filter(
      (user) =>
        user.username.toLowerCase().includes(searchLower) ||
        user.email.toLowerCase().includes(searchLower)
    );

    return users;
  }, [allUsers, debouncedSearchTerm, currentUser, selectedUserIds]);

  /**
   * Get selected users full objects
   */
  const selectedUsers = useMemo(() => {
    return allUsers.filter((user) => selectedUserIds.has(user.id));
  }, [allUsers, selectedUserIds]);

  /**
   * Toggle user selection
   */
  const toggleUserSelection = useCallback((userId: string) => {
    setSelectedUserIds((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(userId)) {
        newSet.delete(userId);
      } else {
        newSet.add(userId);
      }
      return newSet;
    });
    setValidationError(null);
  }, []);

  /**
   * Remove user from selection
   */
  const removeUser = useCallback((userId: string) => {
    setSelectedUserIds((prev) => {
      const newSet = new Set(prev);
      newSet.delete(userId);
      return newSet;
    });
  }, []);

  /**
   * Validate form
   */
  const validateForm = useCallback((): boolean => {
    // Check group name
    if (!groupName.trim()) {
      setValidationError('Group name is required');
      return false;
    }

    // Check at least 1 other member (excluding current user)
    if (selectedUserIds.size === 0) {
      setValidationError('Please select at least one member for the group');
      return false;
    }

    setValidationError(null);
    return true;
  }, [groupName, selectedUserIds]);

  /**
   * Handle creating the group
   */
  const handleCreate = useCallback(async () => {
    // Validate form
    if (!validateForm() || !currentUser) {
      return;
    }

    setIsCreating(true);
    setCreationError(null);

    try {
      // Step 1: Create group with current user as initial member
      const group = await apiClient.createGroup(groupName.trim(), [currentUser.id]);

      // Step 2: Add selected members to the group
      const addMemberPromises = Array.from(selectedUserIds).map((userId) =>
        apiClient.addGroupMember(group.id, userId)
      );

      await Promise.all(addMemberPromises);

      // Step 3: Fetch updated group with all members
      const updatedGroup = await apiClient.getGroup(group.id);

      // Step 4: Add group to store
      addGroup(updatedGroup);

      // Step 5: Select the newly created group
      selectGroup(updatedGroup);

      // Step 6: Call optional callback
      onGroupCreated?.(updatedGroup.id);

      // Step 7: Close modal (form will be reset by useEffect)
      onClose();
    } catch (err) {
      setCreationError('Failed to create group. Please try again.');
    } finally {
      setIsCreating(false);
    }
  }, [
    validateForm,
    currentUser,
    groupName,
    selectedUserIds,
    addGroup,
    selectGroup,
    onGroupCreated,
    onClose,
  ]);

  /**
   * Get user initials for avatar
   */
  const getUserInitials = useCallback((username: string): string => {
    const parts = username.split(/[\s_-]+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return username.substring(0, 2).toUpperCase();
  }, []);

  /**
   * Get avatar color based on user ID (consistent coloring)
   */
  const getAvatarColor = useCallback((userId: string): string => {
    const colors = [
      'from-blue-500 to-cyan-500',
      'from-purple-500 to-pink-500',
      'from-green-500 to-emerald-500',
      'from-orange-500 to-red-500',
      'from-indigo-500 to-blue-500',
      'from-pink-500 to-rose-500',
    ];
    // Simple hash to pick consistent color for user
    const hash = userId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
    return colors[hash % colors.length];
  }, []);

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="2xl"
      scrollBehavior="inside"
      classNames={{
        base: 'max-h-[90vh]',
      }}
      isDismissable={!isCreating}
      hideCloseButton={isCreating}
    >
      <ModalContent>
        {() => (
          <>
            <ModalHeader className="flex flex-col gap-1">
              <h2 className="text-xl font-semibold">Create New Group</h2>
              <p className="text-sm text-default-500 font-normal">
                Start a group conversation with multiple people
              </p>
            </ModalHeader>

            <ModalBody>
              {/* Group Name Input */}
              <div className="mb-4">
                <Input
                  type="text"
                  label="Group Name"
                  placeholder="Enter group name..."
                  value={groupName}
                  onChange={(e) => {
                    setGroupName(e.target.value);
                    setValidationError(null);
                  }}
                  isDisabled={isCreating}
                  isRequired
                  size="lg"
                  classNames={{
                    input: 'text-sm',
                  }}
                  aria-label="Group name"
                  autoComplete="off"
                />
              </div>

              {/* Selected Members */}
              {selectedUsers.length > 0 && (
                <div className="mb-4">
                  <p className="text-sm font-medium text-default-700 mb-2">
                    Selected Members ({selectedUsers.length})
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {selectedUsers.map((user) => (
                      <Chip
                        key={user.id}
                        onClose={() => removeUser(user.id)}
                        variant="flat"
                        color="primary"
                        avatar={
                          <Avatar
                            name={getUserInitials(user.username)}
                            size="sm"
                            classNames={{
                              base: `bg-gradient-to-br ${getAvatarColor(user.id)}`,
                              name: 'text-white text-xs font-semibold',
                            }}
                          />
                        }
                        classNames={{
                          base: 'pl-1',
                        }}
                        isDisabled={isCreating}
                      >
                        {user.username}
                      </Chip>
                    ))}
                  </div>
                </div>
              )}

              {/* Current User Info */}
              {currentUser && (
                <div className="mb-4">
                  <p className="text-sm font-medium text-default-700 mb-2">You</p>
                  <Chip
                    variant="flat"
                    color="success"
                    avatar={
                      <Avatar
                        name={getUserInitials(currentUser.username)}
                        size="sm"
                        classNames={{
                          base: `bg-gradient-to-br ${getAvatarColor(currentUser.id)}`,
                          name: 'text-white text-xs font-semibold',
                        }}
                      />
                    }
                    classNames={{
                      base: 'pl-1',
                    }}
                  >
                    {currentUser.username} (You)
                  </Chip>
                </div>
              )}

              {/* User Search */}
              <div className="mb-2">
                <Input
                  type="text"
                  placeholder="Search members to add..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  isDisabled={isCreating}
                  size="md"
                  classNames={{
                    input: 'text-sm',
                    inputWrapper: 'bg-default-100',
                  }}
                  aria-label="Search members"
                  autoComplete="off"
                />
              </div>

              {/* User List */}
              <div className="min-h-[200px] max-h-[300px] overflow-y-auto">
                {/* Loading State */}
                {isLoadingUsers && (
                  <div className="flex flex-col items-center justify-center h-[200px]">
                    <Spinner size="lg" label="Loading users..." />
                  </div>
                )}

                {/* Error State */}
                {loadError && !isLoadingUsers && (
                  <div className="flex flex-col items-center justify-center h-[200px]">
                    <div className="text-center p-4">
                      <p className="text-danger text-sm mb-3">{loadError}</p>
                      <Button
                        color="primary"
                        size="sm"
                        onPress={() => {
                          setLoadError(null);
                          setIsLoadingUsers(true);
                          apiClient
                            .listUsers()
                            .then((users) => setAllUsers(users))
                            .catch(() => setLoadError('Failed to load users. Please try again.'))
                            .finally(() => setIsLoadingUsers(false));
                        }}
                      >
                        Try Again
                      </Button>
                    </div>
                  </div>
                )}

                {/* User List */}
                {!isLoadingUsers && !loadError && filteredUsers.length > 0 && (
                  <div className="space-y-2" role="list" aria-label="Available users">
                    {filteredUsers.map((user) => {
                      const isSelected = selectedUserIds.has(user.id);
                      return (
                        <Card
                          key={user.id}
                          className={`p-3 transition-colors cursor-pointer ${
                            isSelected
                              ? 'bg-primary-50 border-2 border-primary'
                              : 'hover:bg-default-50'
                          }`}
                          shadow="sm"
                          isPressable
                          isDisabled={isCreating}
                          onPress={() => toggleUserSelection(user.id)}
                          role="listitem"
                          aria-label={`${isSelected ? 'Remove' : 'Add'} ${user.username}`}
                        >
                          <div className="flex items-center gap-3">
                            {/* Avatar */}
                            <Avatar
                              name={getUserInitials(user.username)}
                              size="sm"
                              classNames={{
                                base: `bg-gradient-to-br ${getAvatarColor(user.id)} flex-shrink-0`,
                                name: 'text-white text-xs font-semibold',
                              }}
                            />

                            {/* User Details */}
                            <div className="flex flex-col min-w-0 flex-1">
                              <p className="text-sm font-semibold text-default-900 truncate">
                                {user.username}
                              </p>
                              <p className="text-xs text-default-500 truncate">
                                {user.email}
                              </p>
                            </div>

                            {/* Selection Indicator */}
                            {isSelected && (
                              <div className="flex-shrink-0">
                                <svg
                                  className="w-5 h-5 text-primary"
                                  fill="currentColor"
                                  viewBox="0 0 20 20"
                                  xmlns="http://www.w3.org/2000/svg"
                                  aria-hidden="true"
                                >
                                  <path
                                    fillRule="evenodd"
                                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                                    clipRule="evenodd"
                                  />
                                </svg>
                              </div>
                            )}
                          </div>
                        </Card>
                      );
                    })}
                  </div>
                )}

                {/* No Results State */}
                {!isLoadingUsers &&
                  !loadError &&
                  searchTerm.trim() !== '' &&
                  filteredUsers.length === 0 && (
                    <div className="flex flex-col items-center justify-center h-[200px] text-center px-4">
                      <p className="text-default-500 text-sm">
                        No users found matching "{searchTerm}"
                      </p>
                    </div>
                  )}
              </div>

              {/* Validation Error */}
              {validationError && (
                <div className="mt-3">
                  <p className="text-danger text-sm">{validationError}</p>
                </div>
              )}

              {/* Create Error */}
              {creationError && (
                <div className="mt-3">
                  <p className="text-danger text-sm">{creationError}</p>
                </div>
              )}
            </ModalBody>

            <ModalFooter>
              <Button
                color="default"
                variant="light"
                onPress={onClose}
                isDisabled={isCreating}
              >
                Cancel
              </Button>
              <Button
                color="primary"
                onPress={handleCreate}
                isLoading={isCreating}
                isDisabled={isLoadingUsers || !!loadError}
              >
                {isCreating ? 'Creating...' : 'Create Group'}
              </Button>
            </ModalFooter>
          </>
        )}
      </ModalContent>
    </Modal>
  );
};
