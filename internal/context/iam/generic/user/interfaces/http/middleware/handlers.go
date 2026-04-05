package middleware

import (
	"net/http"

	"gct/internal/context/iam/generic/user/application/query"
	userdomain "gct/internal/context/iam/generic/user/domain"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/cookie"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/security/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthClientAccess ensures the request has a valid access token.
// Injects session and user identity into the request context.
//
// This is the primary authentication middleware for API endpoints.
// It validates the JWT access token and populates the Gin context
// with *shared.AuthSession data.
func (m *AuthMiddleware) AuthClientAccess(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, err.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Inject identity into context for handlers
	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	ctx.Next()
}

// AuthClientRefresh manages the verification of refresh tokens.
// Used primarily at the token regeneration endpoint.
//
// This middleware validates refresh tokens and ensures the associated session
// is still active and not revoked. It uses cryptographic hash verification
// to prevent token forgery.
func (m *AuthMiddleware) AuthClientRefresh(ctx *gin.Context) {
	// Resolve the integration up front so a missing/unknown X-API-Key is a
	// 401 before any refresh-token work is done.
	apiKey := httpx.GetAPIKey(ctx)
	if apiKey == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}
	resolved, err := m.resolver.Resolve(ctx.Request.Context(), apiKey)
	if err != nil || resolved == nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	token := extractRefreshToken(ctx)
	if token == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	// Parse the refresh token format
	rt, err := jwt.ParseRefreshToken(token)
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - invalid refresh token format", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrInvalidRefreshFormat, nil, false)
		ctx.Abort()
		return
	}

	// Verify session existence
	sessionID, err := uuid.Parse(rt.SessionID)
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - invalid session id in token", "session_id", rt.SessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrInvalidSession, nil, false)
		ctx.Abort()
		return
	}

	session, err := m.findSession.Handle(ctx.Request.Context(), query.FindSessionQuery{SessionID: userdomain.SessionID(sessionID)})
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - session not found", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrInvalidRefreshSession, nil, false)
		ctx.Abort()
		return
	}

	// Cross-integration defence: refresh-token hash rotation must never
	// span integrations. If someone replays a refresh token acquired under
	// one audience into a different-audience request, reject.
	if session.IntegrationName != resolved.Name {
		m.l.Warnw("cross_integration_refresh_attempt",
			"session_integration", session.IntegrationName,
			"resolved_integration", resolved.Name,
			"session_id", sessionID,
		)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrInvalidRefreshToken, nil, false)
		ctx.Abort()
		return
	}

	// Cryptographically verify token vs hash in DB (constant-time HMAC compare).
	if !m.refreshHasher.Verify(rt.Secret, rt.ID, session.RefreshTokenHash) {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - invalid refresh token hash", "session_id", sessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrInvalidRefreshToken, nil, false)
		ctx.Abort()
		return
	}

	// Security check for revocation/expiry
	if session.Revoked || session.IsExpired() {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - session revoked or expired", "session_id", sessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrRevokedToken, nil, false)
		ctx.Abort()
		return
	}

	// Inject refresh context
	ctx.Set(consts.CtxSessionID, rt.SessionID)
	ctx.Set(consts.CtxRefreshToken, token)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	ctx.Next()
}

// extractRefreshToken reads the refresh token from the cookie, falling back to
// the Authorization bearer header for native clients.
func extractRefreshToken(ctx *gin.Context) string {
	if token := cookie.GetCookie(ctx, consts.CookieRefreshToken); token != "" {
		return token
	}
	return httpx.ExtractBearerToken(httpx.GetAuthorization(ctx))
}

// AuthAdmin enforces that the authenticated user has an administrative role.
//
// This middleware first validates the access token, then checks if the user
// has an assigned role. The actual role-based access control (RBAC) check for
// specific permissions is delegated to the Authz middleware.
func (m *AuthMiddleware) AuthAdmin(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, err.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Fetch user to verify administrative status
	user, err := m.findUserForAuth.Handle(ctx.Request.Context(), query.FindUserForAuthQuery{UserID: userdomain.UserID(session.UserID)})
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthAdmin - FindUserForAuth", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUserNotFound, nil, false)
		ctx.Abort()
		return
	}

	// Users without an assigned role cannot access admin endpoints
	if user.RoleID == nil {
		m.l.Warnw("AuthMiddleware - AuthAdmin - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())
	ctx.Set(consts.CtxIsAdmin, true)
	ctx.Set(consts.CtxUser, user)

	ctx.Next()
}

// AuthApiKey verifies request authenticity using a static API key.
// Primarily used for server-to-server communication or simple protected resources.
//
// Deprecated: Internal API keys are now handled by SignatureMiddleware via DB.
func (m *AuthMiddleware) AuthApiKey(ctx *gin.Context) {
	apiKey := httpx.GetAPIKey(ctx)
	if apiKey == "" {
		m.l.Warnw("AuthMiddleware - AuthApiKey - API key missing", "ip", httpx.GetIPAddress(ctx))
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrApiKeyMissing, nil, false)
		ctx.Abort()
		return
	}

	// Deprecated: Internal API keys are now handled by SignatureMiddleware via DB.
	response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
	ctx.Abort()
}
