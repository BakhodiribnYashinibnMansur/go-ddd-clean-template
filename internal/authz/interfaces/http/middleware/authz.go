// Package middleware contains Gin handlers for authorization cross-cutting concerns.
package middleware

import (
	"net/http"
	"strings"
	"time"

	access "gct/internal/authz/application/query"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/httpx/response"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/application/query"

	"github.com/gin-gonic/gin"
)

// AuthzMiddleware performs fine-grained access control (RBAC) via the authorization engine.
// It assumes identity has already been verified and injected into the context by a prior auth middleware.
type AuthzMiddleware struct {
	checkAccess     *access.CheckAccessHandler
	findUserForAuth *query.FindUserForAuthHandler
	l               logger.Log
}

// NewAuthzMiddleware creates a new AuthzMiddleware with the required DDD query handlers.
func NewAuthzMiddleware(
	checkAccess *access.CheckAccessHandler,
	findUserForAuth *query.FindUserForAuthHandler,
	l logger.Log,
) *AuthzMiddleware {
	return &AuthzMiddleware{
		checkAccess:     checkAccess,
		findUserForAuth: findUserForAuth,
		l:               l,
	}
}

// Authz is a Gin middleware that evaluates whether the authenticated user has permission
// to access the requested endpoint based on their role.
//
// Flow:
//  1. Retrieve session from context (set by a prior auth middleware).
//  2. Look up the user's role via the User BC read model.
//  3. Super admins bypass all policy checks.
//  4. Delegate to the Authz BC's CheckAccess query for role-based permission evaluation.
func (m *AuthzMiddleware) Authz(ctx *gin.Context) {
	// 1. Retrieve session from context (should be set by a previous auth middleware).
	sessionVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	session, ok := sessionVal.(*shared.AuthSession)
	if !ok {
		m.l.Error("AuthzMiddleware - Authz - session type cast error")
		response.ControllerResponse(ctx, http.StatusInternalServerError, httpx.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	// 2. Find user for auth to get role information.
	user, err := m.findUserForAuth.Handle(ctx.Request.Context(), query.FindUserForAuthQuery{
		UserID: session.UserID,
	})
	if err != nil {
		m.l.Errorw("AuthzMiddleware - Authz - FindUserForAuth", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUserNotFound, nil, false)
		ctx.Abort()
		return
	}

	// 3. Verify user has a role assigned.
	if user.RoleID == nil {
		m.l.Warnw("AuthzMiddleware - Authz - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	// 4. Determine the resource path and HTTP method.
	path := ctx.FullPath()
	if path == "" {
		path = ctx.Request.URL.Path
	}
	method := ctx.Request.Method

	// 5. Build ABAC evaluation context.
	evalCtx := domain.EvaluationContext{
		Attrs: map[string]map[string]any{
			"user":     buildUserAttrs(user),
			"env":      buildEnvAttrs(ctx),
			"resource": {},
			"target":   {},
		},
	}

	// 6. Check access via the Authz BC query handler.
	allowed, err := m.checkAccess.Handle(ctx.Request.Context(), access.CheckAccessQuery{
		RoleID:  *user.RoleID,
		Path:    path,
		Method:  strings.ToUpper(method),
		EvalCtx: evalCtx,
	})
	if err != nil {
		m.l.Errorw("AuthzMiddleware - Authz - CheckAccess", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, httpx.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	if !allowed {
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	ctx.Next()
}

// buildUserAttrs constructs the "user" attribute bag for ABAC evaluation.
// It merges the user's static attributes with their ID and role ID.
func buildUserAttrs(user *shared.AuthUser) map[string]any {
	attrs := make(map[string]any)
	for k, v := range user.Attributes {
		attrs[k] = v
	}
	attrs["id"] = user.ID.String()
	if user.RoleID != nil {
		attrs["role_id"] = user.RoleID.String()
	}
	return attrs
}

// buildEnvAttrs constructs the "env" attribute bag from the current request context.
func buildEnvAttrs(ctx *gin.Context) map[string]any {
	return map[string]any{
		"ip":         ctx.ClientIP(),
		"user_agent": ctx.GetHeader("User-Agent"),
		"time":       time.Now().Format("15:04"),
	}
}
