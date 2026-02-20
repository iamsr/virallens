package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/common/utils"
	"github.com/yourusername/virallens/backend/modules/chat/dto"
)

type ConversationController struct {
	conversationService ConversationService
	messageService      MessageService
}

func NewConversationController(cs ConversationService, ms MessageService) *ConversationController {
	return &ConversationController{
		conversationService: cs,
		messageService:      ms,
	}
}

func (cc *ConversationController) CreateOrGet(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateOrGetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := cc.conversationService.CreateOrGet(userID, req.OtherUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.MapConversationToResponse(conversation))
}

func (cc *ConversationController) List(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversations, err := cc.conversationService.ListUserConversations(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch conversations"})
		return
	}

	resp := make([]dto.ConversationResponse, 0, len(conversations))
	for _, c := range conversations {
		resp = append(resp, dto.MapConversationToResponse(c))
	}

	ctx.JSON(http.StatusOK, resp)
}

func (cc *ConversationController) GetMessages(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var query dto.GetMessagesQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messages, err := cc.messageService.GetConversationMessages(userID, conversationID, query.Cursor, query.Limit)
	if err != nil {
		if err == ErrUnauthorized {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}

	ctx.JSON(http.StatusOK, dto.MapMessagesToResponse(messages))
}

func (cc *ConversationController) SendMessage(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var req dto.SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := cc.messageService.SendConversationMessage(userID, conversationID, req.Content)
	if err != nil {
		if err == ErrUnauthorized {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.MapMessageToResponse(message))
}
