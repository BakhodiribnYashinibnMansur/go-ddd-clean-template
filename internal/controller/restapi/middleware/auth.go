// Package middleware provides Gin-compatible handlers for cross-cutting concerns
// like authentication, authorization, logging, auditing, and security.
package middleware

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/user/client"
	"gct/internal/usecase/user/session"
	"gct/pkg/jwt"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware manages identity verification and permission enforcement.
// It integrates with user, session, and authorization use cases.
type AuthMiddleware struct {
	userUC    *client.UseCaseI
	sessionUC *session.UseCaseI
	authzUC   *authz.UseCase
	cfg       *config.Config
	l         logger.Log
	pubKey    *rsa.PublicKey
}

// NewAuthMiddleware initializes a new authentication middleware instance.
// It pre-parses the RSA public key once at startup for performance.
func NewAuthMiddleware(u *usecase.UseCase, cfg *config.Config, l logger.Log) *AuthMiddleware {
	pubKey, err := jwt.ParseRSAPublicKey(cfg.JWT.PublicKey)
	if err != nil {
		l.Fatalw("AuthMiddleware - NewAuthMiddleware - ParseRSAPublicKey", "error", err)
	}

	return &AuthMiddleware{
		userUC:    &u.User.Client,
		sessionUC: &u.User.Session,
		authzUC:   u.Authz,
		cfg:       cfg,
		l:         l,
		pubKey:    pubKey,
	}
}

// validateAccessToken is a private helper that extracts, parses, and verifies an access token.
// Returns the corresponding session if valid, or an error otherwise.
func (m *AuthMiddleware) validateAccessToken(ctx *gin.Context) (*domain.Session, error) {
	// Strategy 1: HTTP-Only Cookie (common for Web/Browser clients)
	tokenStr := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	// Strategy 2: Authorization Header (common for Native/Mobile/CLI clients)
	if tokenStr == "" {
		tokenStr = ExtractBearerToken(ctx)
	}

	if tokenStr == "" {
		return nil, util.ErrUnAuth
	}

	// Parsing and cryptographic verification
	metadata, err := m.parseAndValidateMetadata(tokenStr)
	if err != nil {
		return nil, err
	}

	// Logical verification: Is the session active and known?
	sessionID, err := uuid.Parse(metadata.SessionID)
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Warnw("AuthMiddleware - validateAccessToken - Invalid session ID", "session_id", metadata.SessionID)
		return nil, util.ErrInvalidSession
	}

	session, err := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - validateAccessToken - Get", "error", err)
		return nil, util.ErrRevokedToken
	}

	return session, nil
}

// parseAndValidateMetadata parses the raw JWT string and validates core claims like issuer and token type.
func (m *AuthMiddleware) parseAndValidateMetadata(tokenStr string) (*jwt.AccessTokenClaims, error) {
	metadata, err := jwt.ParseAccessToken(tokenStr, m.pubKey, m.cfg.JWT.Issuer, "")
	if err != nil {
		if errors.Is(err, jwt.ErrAccessTokenExpired) {
			return nil, util.ErrExpiredToken
		}
		return nil, util.ErrInvalidToken
	}

	if metadata.Issuer != m.cfg.JWT.Issuer {
		m.l.Warnw("AuthMiddleware - parseAndValidateMetadata - Invalid issuer", "issuer", metadata.Issuer)
		return nil, util.ErrInvalidIssuer
	}

	if metadata.Type != consts.TokenTypeAccess {
		m.l.Warnw("AuthMiddleware - parseAndValidateMetadata - Invalid token type", "type", metadata.Type)
		return nil, util.ErrInvalidType
	}

	return metadata, nil
}

// AuthClientAccess ensures the request has a valid access token.
// Injects session and user identity into the request context.
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
func (m *AuthMiddleware) AuthClientRefresh(ctx *gin.Context) {
	token := cookie.GetCookie(ctx, consts.COOKIE_REFRESH_TOKEN)
	if token == "" {
		token = ExtractBearerToken(ctx)
	}

	if token == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	// Parse the refresh token format
	rt, err := jwt.ParseRefreshToken(token)
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthClientRefresh - invalid refresh token format", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrInvalidRefreshFormat, nil, false)
		ctx.Abort()
		return
	}

	// Verify session existence
	sessionID, err := uuid.Parse(rt.SessionID)
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthClientRefresh - invalid session id in token", "session_id", rt.SessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrInvalidSession, nil, false)
		ctx.Abort()
		return
	}

	session, err := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthClientRefresh - session not found", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrInvalidRefreshSession, nil, false)
		ctx.Abort()
		return
	}

	// Cryptographically verify token vs hash in DB
	if !rt.Verify(session.RefreshTokenHash) {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthClientRefresh - invalid refresh token hash", "session_id", sessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrInvalidRefreshToken, nil, false)
		ctx.Abort()
		return
	}

	// Security check for revocation/expiry
	if session.Revoked || session.IsExpired() {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthClientRefresh - session revoked or expired", "session_id", sessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrRevokedToken, nil, false)
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

// AuthAdmin enforces that the authenticated user has an administrative role.
func (m *AuthMiddleware) AuthAdmin(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, err.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Verify administrative status
	user, err := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthAdmin - User Get", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrUserNotFound, nil, false)
		ctx.Abort()
		return
	}

	if user.RoleID == nil {
		m.l.WithContext(ctx.Request.Context()).Warnw("AuthMiddleware - AuthAdmin - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, util.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	role, err := m.authzUC.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - AuthAdmin - Role Get", "error", err)
		response.ControllerResponse(ctx, http.StatusForbidden, util.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	// Simple case-insensitive match for admin role
	if !strings.Contains(strings.ToLower(role.Name), consts.RoleAdmin) {
		m.l.WithContext(ctx.Request.Context()).Warnw("AuthMiddleware - AuthAdmin - User is not admin", "user_id", user.ID, "role", role.Name)
		response.ControllerResponse(ctx, http.StatusForbidden, util.ErrAccessDenied, nil, false)
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
func (m *AuthMiddleware) AuthApiKey(ctx *gin.Context) {
	apiKey := util.GetAPIKey(ctx)
	if apiKey == "" {
		m.l.WithContext(ctx.Request.Context()).Warnw("AuthMiddleware - AuthApiKey - API key missing", "ip", util.GetIPAddress(ctx))
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrApiKeyMissing, nil, false)
		ctx.Abort()
		return
	}

	if apiKey != m.cfg.APIKeys.XApiKey {
		m.l.WithContext(ctx.Request.Context()).Warnw("AuthMiddleware - AuthApiKey - Invalid API key", "ip", util.GetIPAddress(ctx))
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrInvalidApiKey, nil, false)
		ctx.Abort()
		return
	}

	ctx.Set(consts.CtxApiKeyAuth, true)
	ctx.Next()
}

// AuthWeb wraps authentication for web/admin panel, supporting interactive features like auto-refreshing expired tokens.
func (m *AuthMiddleware) AuthWeb(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		// Attempt logical auto-refresh if token is simply expired but refresh token exists
		if errors.Is(err, util.ErrExpiredToken) {
			refreshToken := cookie.GetCookie(ctx, consts.COOKIE_REFRESH_TOKEN)
			if refreshToken != "" {
				rt, pErr := jwt.ParseRefreshToken(refreshToken)

				if pErr == nil {
					sID, _ := uuid.Parse(rt.SessionID)
					sess, sErr := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sID})

					if sErr == nil && !sess.Revoked && !sess.IsExpired() && rt.Verify(sess.RefreshTokenHash) {
						// Perform silent Token Rotation
						res, rErr := (*m.userUC).RotateSession(ctx.Request.Context(), &domain.RefreshIn{SessionID: sID})
						if rErr == nil {
							isSecure := ctx.Request.TLS != nil || ctx.Request.Header.Get("X-Forwarded-Proto") == "https"

							ctx.SetCookie(consts.COOKIE_ACCESS_TOKEN, res.AccessToken, int(m.cfg.JWT.AccessTTL.Seconds()), "/", "", isSecure, true)
							ctx.SetCookie(consts.COOKIE_REFRESH_TOKEN, res.RefreshToken, int(m.cfg.JWT.RefreshTTL.Seconds()), "/", "", isSecure, true)

							freshSess, fErr := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sID})
							if fErr != nil {
								m.l.Errorw("AuthWeb - Failed to fetch fresh session", "error", fErr)
								ctx.Redirect(http.StatusFound, "/admin/login")
								ctx.Abort()
								return
							}

							ctx.Set(consts.CtxSessionID, freshSess.ID)
							ctx.Set(consts.CtxSession, freshSess)
							ctx.Set(consts.CtxUserID, freshSess.UserID.String())

							u, uErr := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &freshSess.UserID})
							if uErr == nil {
								ctx.Set(consts.CtxUser, u)
							}
							ctx.Next()
							return
						} else {
							m.l.Warnw("AuthWeb - Auto-refresh rotation failed", "error", rErr)
						}
					}
				}
			}
		}

		// Fallback to login redirect for web interaction
		ctx.Redirect(http.StatusFound, "/admin/login?return_url="+url.QueryEscape(ctx.Request.RequestURI))
		ctx.Abort()
		return
	}

	// standard identity injection
	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	user, uErr := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if uErr == nil {
		ctx.Set(consts.CtxUser, user)

		if user.RoleID != nil {
			role, rErr := m.authzUC.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
			if rErr == nil {
				ctx.Set(consts.CtxRoleTitle, role.Name)
			} else {
				m.l.Warnw("AuthWeb - Failed to fetch role", "role_id", user.RoleID, "error", rErr)
			}
		}
	} else {
		m.l.Warnw("AuthWeb - Failed to fetch user for context", "user_id", session.UserID, "error", uErr)
	}

	ctx.Next()
}

// Authz performs fine-grained access control (RBAC/ABAC) via the authorization engine.
// It assumes identity has already been verified and injected into the context.
func (m *AuthMiddleware) Authz(ctx *gin.Context) {
	sessionVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	session, ok := sessionVal.(*domain.Session)
	if !ok {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - Authz - session type cast error")
		response.ControllerResponse(ctx, http.StatusInternalServerError, util.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	user, err := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - Authz - Get User", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, util.ErrUserNotFound, nil, false)
		ctx.Abort()
		return
	}

	if user.RoleID == nil {
		m.l.WithContext(ctx.Request.Context()).Warnw("AuthMiddleware - Authz - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, util.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	role, err := m.authzUC.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - Authz - Role Get", "error", err)
		response.ControllerResponse(ctx, http.StatusForbidden, util.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	// Permanent bypass for superadmins
	if strings.ToLower(role.Name) == consts.RoleSuperAdmin {
		m.l.WithContext(ctx.Request.Context()).Infow("Super admin access granted", "user_id", user.ID)
		ctx.Next()
		return
	}

	path := ctx.FullPath()
	if path == "" {
		path = ctx.Request.URL.Path
	}
	method := ctx.Request.Method

	// Prepare dynamic environment for policy evaluation
	env := map[string]any{
		consts.PolicyKeyIP:        util.GetIPAddress(ctx),
		consts.PolicyKeyUserAgent: util.GetUserAgent(ctx),
		consts.PolicyKeyTime:      time.Now(),
		consts.PolicyKeyUserID:    user.ID,
		consts.PolicyKeyRoleID:    *user.RoleID,
	}

	for _, p := range ctx.Params {
		env[p.Key] = p.Value
	}

	// Query Authz Engine
	allowed, err := m.authzUC.Access.Check(ctx.Request.Context(), session.UserID, session, path, method, env)
	if err != nil {
		m.l.WithContext(ctx.Request.Context()).Errorw("AuthMiddleware - Authz - Check", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, util.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	if !allowed {
		response.ControllerResponse(ctx, http.StatusForbidden, util.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	ctx.Next()
}

// ExtractBearerToken parses the "Bearer <token>" string from authorization sources.
func ExtractBearerToken(ctx *gin.Context) string {
	bearToken := util.GetAuthorization(ctx)
	token := bearToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BearerToken {
		return ""
	}

	return onlyToken[1]
}

// ExtractBasicToken parses the "Basic <token>" string from authorization sources.
func ExtractBasicToken(ctx *gin.Context) string {
	basicToken := util.GetAuthorization(ctx)
	token := basicToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BasicToken {
		return ""
	}

	return onlyToken[1]
}
