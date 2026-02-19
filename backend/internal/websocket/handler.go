package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, check against allowed origins
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub                 *Hub
	messageService      *service.MessageService
	conversationService service.ConversationService
	groupService        service.GroupService
	jwtService          service.JWTService
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	hub *Hub,
	messageService *service.MessageService,
	conversationService service.ConversationService,
	groupService service.GroupService,
	jwtService service.JWTService,
) *Handler {
	return &Handler{
		hub:                 hub,
		messageService:      messageService,
		conversationService: conversationService,
		groupService:        groupService,
		jwtService:          jwtService,
	}
}

// HandleWebSocket handles WebSocket connection requests
func (h *Handler) HandleWebSocket(c echo.Context) error {
	// Get token from query parameter
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token required"})
	}

	// Validate token and get user ID
	claims, err := h.jwtService.ValidateAccessToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}

	userID := claims.UserID

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return err
	}

	// Create client
	client := &Client{
		ID:     uuid.New(),
		UserID: userID,
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register client with hub
	h.hub.RegisterClient(client)

	// Start pumps
	client.StartPumps(h.handleMessage)

	return nil
}

// handleMessage processes incoming WebSocket messages
func (h *Handler) handleMessage(client *Client, data []byte) error {
	var msg OutgoingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return errors.New("invalid message format")
	}

	// Validate message type
	if msg.Type != "message" {
		return errors.New("invalid message type")
	}

	// Validate content
	if msg.Content == "" {
		return errors.New("message content cannot be empty")
	}

	// Handle conversation message
	if msg.ConversationID != nil {
		return h.handleConversationMessage(client, *msg.ConversationID, msg.Content)
	}

	// Handle group message
	if msg.GroupID != nil {
		return h.handleGroupMessage(client, *msg.GroupID, msg.Content)
	}

	return errors.New("either conversation_id or group_id must be provided")
}

// handleConversationMessage handles sending a message to a conversation
func (h *Handler) handleConversationMessage(client *Client, conversationID uuid.UUID, content string) error {
	// Send message via service
	message, err := h.messageService.SendConversationMessage(client.UserID, conversationID, content)
	if err != nil {
		return err
	}

	// Get conversation to find participants
	conversation, err := h.conversationService.GetByID(client.UserID, conversationID)
	if err != nil {
		return err
	}

	// Broadcast to all participants
	if err := h.hub.BroadcastMessage(message, conversation.Participants); err != nil {
		log.Printf("Failed to broadcast message: %v", err)
	}

	return nil
}

// handleGroupMessage handles sending a message to a group
func (h *Handler) handleGroupMessage(client *Client, groupID uuid.UUID, content string) error {
	// Send message via service
	message, err := h.messageService.SendGroupMessage(client.UserID, groupID, content)
	if err != nil {
		return err
	}

	// Get group to find members
	group, err := h.groupService.GetByID(client.UserID, groupID)
	if err != nil {
		return err
	}

	// Broadcast to all members
	if err := h.hub.BroadcastMessage(message, group.Members); err != nil {
		log.Printf("Failed to broadcast message: %v", err)
	}

	return nil
}

// GetHub returns the hub instance
func (h *Handler) GetHub() *Hub {
	return h.hub
}

// SendMessageToUsers is a helper to send a custom message to specific users
func (h *Handler) SendMessageToUsers(userIDs []uuid.UUID, msgType string, data interface{}) error {
	wsMsg := WSMessage{
		Type: msgType,
		Data: data,
	}

	msgData, err := json.Marshal(wsMsg)
	if err != nil {
		return err
	}

	h.hub.BroadcastToUsers(userIDs, msgData)
	return nil
}

// SendErrorToUser sends an error message to a specific user
func (h *Handler) SendErrorToUser(userID uuid.UUID, message string) error {
	return h.SendMessageToUsers([]uuid.UUID{userID}, "error", map[string]string{
		"message": message,
	})
}

// NotifyTyping notifies users that someone is typing
func (h *Handler) NotifyTyping(userID uuid.UUID, targetID uuid.UUID, isTyping bool) error {
	return h.SendMessageToUsers([]uuid.UUID{targetID}, "typing", map[string]interface{}{
		"user_id":   userID,
		"is_typing": isTyping,
	})
}

// GetOnlineStatus returns online status for a list of user IDs
func (h *Handler) GetOnlineStatus(userIDs []uuid.UUID) map[uuid.UUID]bool {
	status := make(map[uuid.UUID]bool)
	for _, userID := range userIDs {
		status[userID] = h.hub.IsUserOnline(userID)
	}
	return status
}
