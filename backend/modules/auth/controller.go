package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamsr/virallens/backend/common/utils"
	"github.com/iamsr/virallens/backend/modules/auth/dto"
	userdto "github.com/iamsr/virallens/backend/modules/user/dto"
)

type Controller struct {
	authService Service
}

func NewController(authService Service) *Controller {
	return &Controller{authService: authService}
}

func (c *Controller) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Register(&req)
	if err != nil {
		if err == ErrUserAlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":          userdto.MapDomainUserToResponse(resp.User),
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func (c *Controller) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Login(&req)
	if err != nil {
		if err == ErrInvalidCredentials {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":          userdto.MapDomainUserToResponse(resp.User),
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func (c *Controller) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		if err != ErrTokenExpired && err != ErrInvalidToken {
			status = http.StatusInternalServerError
		}
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":          userdto.MapDomainUserToResponse(resp.User),
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func (c *Controller) Logout(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := c.authService.Logout(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
