package util

import (
	"fmt"

	"gct/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gin Context Helpers (Keys set by Middleware)

func GetUserID(ctx *gin.Context) (uuid.UUID, error) {
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
		parsed, err := uuid.Parse(userID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
		}
		return parsed, nil
	}

	return uuid.Nil, ErrUserIdNotFound
}

func GetCtxSessionID(ctx *gin.Context) (uuid.UUID, error) {
	sessionID, ok := ctx.Get(consts.CtxSessionID)
	if !ok {
		return uuid.Nil, ErrSessionIDNotFound
	}
	if s, ok := sessionID.(string); ok {
		parsed, err := uuid.Parse(s)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to parse session ID: %w", err)
		}
		return parsed, nil
	}
	if u, ok := sessionID.(uuid.UUID); ok {
		return u, nil
	}
	return uuid.Nil, ErrInvalidSessionID
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
		parsed, err := uuid.Parse(roleIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to parse role ID: %w", err)
		}
		return parsed, nil
	}
	return uuid.Nil, ErrRoleNotFound
}
