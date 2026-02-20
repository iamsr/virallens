package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/common/utils"
	"github.com/yourusername/virallens/backend/modules/chat/dto"
)

type GroupController struct {
	groupService   GroupService
	messageService MessageService
}

func NewGroupController(gs GroupService, ms MessageService) *GroupController {
	return &GroupController{
		groupService:   gs,
		messageService: ms,
	}
}

func (gc *GroupController) Create(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := gc.groupService.Create(req.Name, userID, req.Members)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.MapGroupToResponse(group))
}

func (gc *GroupController) List(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groups, err := gc.groupService.ListUserGroups(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch groups"})
		return
	}

	ctx.JSON(http.StatusOK, dto.MapGroupsToResponse(groups))
}

func (gc *GroupController) Get(ctx *gin.Context) {
	_, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	group, err := gc.groupService.GetByID(groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	ctx.JSON(http.StatusOK, dto.MapGroupToResponse(group))
}

func (gc *GroupController) AddMember(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req dto.AddMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := gc.groupService.AddMember(userID, groupID, req.UserID); err != nil {
		if err == ErrUnauthorized {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

func (gc *GroupController) RemoveMember(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req dto.RemoveMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := gc.groupService.RemoveMember(userID, groupID, req.UserID); err != nil {
		if err == ErrUnauthorized {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

func (gc *GroupController) GetMessages(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var query dto.GetMessagesQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messages, err := gc.messageService.GetGroupMessages(userID, groupID, query.Cursor, query.Limit)
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

func (gc *GroupController) SendMessage(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req dto.SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := gc.messageService.SendGroupMessage(userID, groupID, req.Content)
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
