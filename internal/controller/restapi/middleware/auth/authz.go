package auth

import (
	"net/http"
	"strings"
	"time"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Authz performs fine-grained access control (RBAC/ABAC) via the authorization engine.
// It assumes identity has already been verified and injected into the context.
//
// This middleware implements a policy-based authorization system that evaluates
// permissions based on:
// - User role and permissions
// - Request path and HTTP method
// - Dynamic context (IP address, time, custom parameters)
//
// Super admins bypass all policy checks for operational convenience.
func (m *AuthMiddleware) Authz(ctx *gin.Context) {
	// Retrieve session from context (should be set by a previous auth middleware)
	sessionVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	session, ok := sessionVal.(*domain.Session)
	if !ok {
		m.l.Error("AuthMiddleware - Authz - session type cast error")
		response.ControllerResponse(ctx, http.StatusInternalServerError, httpx.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	// Fetch user details
	user, err := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if err != nil {
		m.l.Errorw("AuthMiddleware - Authz - Get User", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUserNotFound, nil, false)
		ctx.Abort()
		return
	}

	if user.RoleID == nil {
		m.l.Warnw("AuthMiddleware - Authz - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	// Fetch role details
	role, err := m.authzUC.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
	if err != nil {
		m.l.Errorw("AuthMiddleware - Authz - Role Get", "error", err)
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	// Permanent bypass for superadmins
	if strings.ToLower(role.Name) == consts.RoleSuperAdmin {
		m.l.Infow("Super admin access granted", "user_id", user.ID)
		ctx.Next()
		return
	}

	// Determine the resource path and method
	path := ctx.FullPath()
	if path == "" {
		path = ctx.Request.URL.Path
	}
	method := ctx.Request.Method

	// Prepare dynamic environment for policy evaluation
	// This allows policies to make decisions based on runtime context
	env := map[string]any{
		consts.PolicyKeyIP:        httpx.GetIPAddress(ctx),
		consts.PolicyKeyUserAgent: httpx.GetUserAgent(ctx),
		consts.PolicyKeyTime:      time.Now(),
		consts.PolicyKeyUserID:    user.ID,
		consts.PolicyKeyRoleID:    *user.RoleID,
	}

	// Include URL parameters in the environment for path-based policies
	for _, p := range ctx.Params {
		env[p.Key] = p.Value
	}

	// Query Authorization Engine
	// This is where the actual permission check happens against stored policies
	allowed, err := m.authzUC.Access.Check(ctx.Request.Context(), session.UserID, session, path, method, env)
	if err != nil {
		m.l.Errorw("AuthMiddleware - Authz - Check", "error", err)
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
