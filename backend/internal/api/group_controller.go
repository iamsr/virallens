package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/api/dto"
	"github.com/yourusername/virallens/backend/internal/service"
)

type GroupController struct {
	groupService   service.GroupService
	messageService *service.MessageService
}

func NewGroupController(groupService service.GroupService, messageService *service.MessageService) *GroupController {
	return &GroupController{
		groupService:   groupService,
		messageService: messageService,
	}
}

// Create creates a new group
func (gc *GroupController) Create(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req dto.CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Parse member UUIDs
	memberIDs := make([]uuid.UUID, 0, len(req.Members))
	for _, memberStr := range req.Members {
		memberID, err := parseUUID(memberStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid member ID: " + memberStr})
		}
		memberIDs = append(memberIDs, memberID)
	}

	group, err := gc.groupService.Create(req.Name, userID, memberIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, group)
}

// GetByID retrieves a group by ID
func (gc *GroupController) GetByID(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	groupID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid group ID"})
	}

	group, err := gc.groupService.GetByID(userID, groupID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, group)
}

// List retrieves all groups for the authenticated user
func (gc *GroupController) List(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	groups, err := gc.groupService.ListByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, groups)
}

// AddMember adds a member to a group
func (gc *GroupController) AddMember(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	groupID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid group ID"})
	}

	var req dto.AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	newMemberID, err := parseUUID(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
	}

	if err := gc.groupService.AddMember(userID, groupID, newMemberID); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "member added successfully"})
}

// RemoveMember removes a member from a group
func (gc *GroupController) RemoveMember(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	groupID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid group ID"})
	}

	var req dto.RemoveMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	memberToRemoveID, err := parseUUID(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
	}

	if err := gc.groupService.RemoveMember(userID, groupID, memberToRemoveID); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "member removed successfully"})
}

// GetMessages retrieves messages for a group
func (gc *GroupController) GetMessages(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	groupID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid group ID"})
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

	messages, err := gc.messageService.GetGroupMessages(userID, groupID, cursor, query.Limit)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, messages)
}
