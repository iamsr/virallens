package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrGroupNotFound         = errors.New("group not found")
	ErrMessageNotFound       = errors.New("message not found")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidToken          = errors.New("invalid token")
	ErrTokenExpired          = errors.New("token expired")
	ErrNotGroupMember        = errors.New("not a group member")
	ErrNotConversationMember = errors.New("not a conversation participant")
)
