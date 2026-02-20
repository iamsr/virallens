package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamsr/virallens/backend/common/utils"
	"github.com/iamsr/virallens/backend/modules/user/dto"
)

type Controller struct {
	userService Service
}

func NewController(userService Service) *Controller {
	return &Controller{userService: userService}
}

func (c *Controller) ListUsers(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	users, err := c.userService.ListUsers(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	response := dto.MapDomainUsersToResponse(users)
	ctx.JSON(http.StatusOK, response)
}
