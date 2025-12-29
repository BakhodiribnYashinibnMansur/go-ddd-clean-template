package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gct/consts"
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

func GetUserIDInt64(ctx *gin.Context) (int64, error) {
	id, ok := ctx.Get(consts.CtxUserID)
	if !ok {
		return 0, ErrUserIdNotFound
	}

	if i, ok := id.(int64); ok {
		return i, nil
	}

	if s, ok := id.(string); ok {
		return strconv.ParseInt(s, 10, 64)
	}

	return 0, ErrUserIdNotFound
}

func GetCtxSessionID(ctx *gin.Context) (string, error) {
	sessionID, ok := ctx.Get(consts.CtxSessionID)
	if !ok {
		return "", ErrSessionIDNotFound
	}
	if s, ok := sessionID.(string); ok {
		return s, nil
	}
	if u, ok := sessionID.(uuid.UUID); ok {
		return u.String(), nil
	}
	return "", ErrInvalidSessionID
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
