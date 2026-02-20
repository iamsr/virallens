package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/api/dto"
	"github.com/yourusername/virallens/backend/internal/service"
)

type ConversationController struct {
	conversationService service.ConversationService
	messageService      *service.MessageService
}

func NewConversationController(conversationService service.ConversationService, messageService *service.MessageService) *ConversationController {
	return &ConversationController{
		conversationService: conversationService,
		messageService:      messageService,
	}
}

// CreateOrGet creates a new conversation or returns existing one
func (cc *ConversationController) CreateOrGet(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req dto.CreateOrGetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	otherUserID, err := parseUUID(req.OtherUserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid other_user_id"})
	}

	conversation, err := cc.conversationService.CreateOrGet(userID, otherUserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, conversation)
}

// GetByID retrieves a conversation by ID
func (cc *ConversationController) GetByID(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	conversationID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid conversation ID"})
	}

	conversation, err := cc.conversationService.GetByID(userID, conversationID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, conversation)
}

// List retrieves all conversations for the authenticated user
func (cc *ConversationController) List(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	conversations, err := cc.conversationService.ListByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, conversations)
}

// GetMessages retrieves messages for a conversation
func (cc *ConversationController) GetMessages(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	conversationID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid conversation ID"})
	}

	var query dto.GetMessagesQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	// Set default limit
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}

	// Parse cursor if provided
	var cursor *time.Time
	if query.Cursor != "" {
		t, err := time.Parse(time.RFC3339, query.Cursor)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid cursor format"})
		}
		cursor = &t
	}

	messages, err := cc.messageService.GetConversationMessages(userID, conversationID, cursor, query.Limit)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, messages)
}
