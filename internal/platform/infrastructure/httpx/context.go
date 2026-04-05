// Package util provides cross-cutting helper functions for the Gin REST API controllers.
package httpx

import (
	"fmt"

	"gct/internal/platform/domain/consts"

	shared "gct/internal/platform/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Gin Context Helpers are used to safely retrieve values injected by the authentication
// and authorization middlewares into the request context.

// GetUserID retrieves the authenticated User ID from the Gin context.
// It supports both raw uuid.UUID types and string representations, attempting to parse the latter.
func GetUserID(ctx *gin.Context) (uuid.UUID, error) {
	id, ok := ctx.Get(consts.CtxUserID)
	if !ok {
		return uuid.Nil, ErrUserIdNotFound
	}

	// Direct type assertion for efficiency if the middleware set it as a UUID object
	if u, ok := id.(uuid.UUID); ok {
		return u, nil
	}

	// Fallback for cases where the ID was stored as a string (e.g. from a JWT claim)
	if userID, ok := id.(string); ok {
		parsed, err := uuid.Parse(userID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
		}
		return parsed, nil
	}

	return uuid.Nil, ErrUserIdNotFound
}

// GetCtxSessionID extracts the active Session ID from the Gin context.
// Ensures the resulting value is a valid UUID, regardless of how it was originally stored.
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

// GetUserRole retrieves the Role ID associated with the current requester from the context.
// Returns an error if the role information is missing or malformed.
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

// GetCtxSession retrieves the AuthSession from the context (set by DDD auth middleware).
func GetCtxSession(ctx *gin.Context) (*shared.AuthSession, error) {
	sessionVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		return nil, ErrSessionNotFound
	}
	session, ok := sessionVal.(*shared.AuthSession)
	if !ok {
		return nil, ErrSessionCastFailed
	}
	return session, nil
}
