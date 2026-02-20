package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/yourusername/virallens/backend/modules/auth"
	"github.com/yourusername/virallens/backend/modules/chat"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub                 *Hub
	messageService      chat.MessageService
	conversationService chat.ConversationService
	groupService        chat.GroupService
	jwtService          auth.JWTService
}

func NewHandler(
	hub *Hub,
	messageService chat.MessageService,
	conversationService chat.ConversationService,
	groupService chat.GroupService,
	jwtService auth.JWTService,
) *Handler {
	return &Handler{
		hub:                 hub,
		messageService:      messageService,
		conversationService: conversationService,
		groupService:        groupService,
		jwtService:          jwtService,
	}
}

// HandleWebSocket uses gin.Context instead of echo.Context
func (h *Handler) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	userIDStr, err := h.jwtService.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		ID:     uuid.New(),
		UserID: userID,
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.hub.RegisterClient(client)
	// Send the connecting client the current list of online users
	onlineIDs := h.hub.GetOnlineUsers()
	onlineStrings := make([]string, 0, len(onlineIDs))
	for _, id := range onlineIDs {
		onlineStrings = append(onlineStrings, id.String())
	}
	if presenceList, err := json.Marshal(WSMessage{
		Type: "presence_list",
		Data: onlineStrings,
	}); err == nil {
		select {
		case client.Send <- presenceList:
		default:
		}
	}
	client.StartPumps(h.handleMessage)
}

func (h *Handler) handleMessage(client *Client, data []byte) error {
	var msg OutgoingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return errors.New("invalid message format")
	}

	if msg.Type != "message" {
		return errors.New("invalid message type")
	}

	if msg.Content == "" {
		return errors.New("message content cannot be empty")
	}

	if msg.ConversationID != nil {
		conversationID, err := uuid.Parse(*msg.ConversationID)
		if err != nil {
			return errors.New("invalid conversation_id format")
		}
		return h.handleConversationMessage(client, conversationID, msg.Content)
	}

	if msg.GroupID != nil {
		groupID, err := uuid.Parse(*msg.GroupID)
		if err != nil {
			return errors.New("invalid group_id format")
		}
		return h.handleGroupMessage(client, groupID, msg.Content)
	}

	return errors.New("either conversation_id or group_id must be provided")
}

func (h *Handler) handleConversationMessage(client *Client, conversationID uuid.UUID, content string) error {
	message, err := h.messageService.SendConversationMessage(client.UserID, conversationID, content)
	if err != nil {
		return err
	}

	conversation, err := h.conversationService.GetByID(conversationID)
	if err != nil {
		return err
	}

	participants := []uuid.UUID{conversation.Participant1, conversation.Participant2}
	if err := h.hub.BroadcastMessage(message, participants); err != nil {
		log.Printf("Failed to broadcast message: %v", err)
	}
	return nil
}

func (h *Handler) handleGroupMessage(client *Client, groupID uuid.UUID, content string) error {
	message, err := h.messageService.SendGroupMessage(client.UserID, groupID, content)
	if err != nil {
		return err
	}

	group, err := h.groupService.GetByID(groupID)
	if err != nil {
		return err
	}

	participants := make([]uuid.UUID, 0, len(group.Members))
	for _, m := range group.Members {
		participants = append(participants, m.ID)
	}

	if err := h.hub.BroadcastMessage(message, participants); err != nil {
		log.Printf("Failed to broadcast message: %v", err)
	}

	return nil
}

func (h *Handler) GetHub() *Hub {
	return h.hub
}
