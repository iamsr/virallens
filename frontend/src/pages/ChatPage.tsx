import { useEffect } from 'react';
import { apiClient } from '../services/api';
import { useChatStore } from '../stores/chatStore';
import { useWebSocket } from '../hooks/useWebSocket';

export function ChatPage() {
  const { conversations, groups, setConversations, setGroups } = useChatStore();
  useWebSocket(); // Initialize WebSocket connection

  useEffect(() => {
    // Load conversations and groups
    const loadData = async () => {
      try {
        const [convs, grps] = await Promise.all([
          apiClient.getConversations(),
          apiClient.getGroups(),
        ]);
        setConversations(convs);
        setGroups(grps);
      } catch (error) {
        console.error('Failed to load chat data:', error);
      }
    };

    loadData();
  }, [setConversations, setGroups]);

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <div className="w-80 bg-white border-r border-gray-200">
        <div className="p-4 border-b border-gray-200">
          <h1 className="text-xl font-bold">ViralLens Chat</h1>
        </div>

        {/* Conversations */}
        <div className="p-4">
          <h2 className="text-sm font-semibold text-gray-500 mb-2">
            Conversations
          </h2>
          {conversations.length === 0 ? (
            <p className="text-sm text-gray-400">No conversations yet</p>
          ) : (
            <div className="space-y-2">
              {conversations.map((conv) => (
                <div
                  key={conv.id}
                  className="p-3 rounded-lg hover:bg-gray-50 cursor-pointer"
                >
                  <p className="font-medium">Conversation</p>
                  <p className="text-sm text-gray-500">
                    {new Date(conv.created_at).toLocaleDateString()}
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Groups */}
        <div className="p-4">
          <h2 className="text-sm font-semibold text-gray-500 mb-2">Groups</h2>
          {groups.length === 0 ? (
            <p className="text-sm text-gray-400">No groups yet</p>
          ) : (
            <div className="space-y-2">
              {groups.map((group) => (
                <div
                  key={group.id}
                  className="p-3 rounded-lg hover:bg-gray-50 cursor-pointer"
                >
                  <p className="font-medium">{group.name}</p>
                  <p className="text-sm text-gray-500">
                    {group.members.length} members
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center text-gray-400">
          <p className="text-lg">Select a conversation to start chatting</p>
        </div>
      </div>
    </div>
  );
}
