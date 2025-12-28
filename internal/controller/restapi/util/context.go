package util

import (
	"fmt"

	"github.com/evrone/go-clean-template/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gin Context Helpers (Keys set by Middleware)

func GetUserIDUUID(ctx *gin.Context) (uuid.UUID, error) {
	id, ok := ctx.Get(consts.CtxUserID)
	if !ok {
		return uuid.Nil, ErrUserIdNotFound
	}

	// Check if already UUID
	if u, ok := id.(uuid.UUID); ok {
		return u, nil
	}

	// If string, parse it
	if userID, ok := id.(string); ok {
		return uuid.Parse(userID)
	}

	return uuid.Nil, ErrUserIdNotFound
}

func GetCtxSessionID(ctx *gin.Context) (string, error) {
	sessionID, ok := ctx.Get(consts.CtxSessionID)
	if !ok {
		return "", fmt.Errorf("sessionID not found")
	}
	if s, ok := sessionID.(string); ok {
		return s, nil
	}
	if u, ok := sessionID.(uuid.UUID); ok {
		return u.String(), nil
	}
	return "", fmt.Errorf("sessionID is not a string or UUID")
}

func GetUserRole(ctx *gin.Context) (uuid.UUID, error) {
	role, ok := ctx.Get(consts.CtxRoleID)
	if !ok {
		return uuid.Nil, ErrRoleNotFound
	}
	if roleID, ok := role.(uuid.UUID); ok {
		return roleID, nil
	}
	if roleIDStr, ok := role.(string); ok {
		return uuid.Parse(roleIDStr)
	}
	return uuid.Nil, ErrRoleNotFound
}
