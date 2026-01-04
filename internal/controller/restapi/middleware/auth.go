package middleware

import (
	"crypto/rsa"
	"errors"
	"net/http"
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

var (
	errUnAuth                = errors.New("unauthorized. token is missing")
	errInvalidToken          = errors.New("unauthorized. token is invalid")
	errExpiredToken          = errors.New("unauthorized. token is expired")
	errRevokedToken          = errors.New("unauthorized. token is revoked")
	errInvalidIssuer         = errors.New("invalid issuer")
	errInvalidType           = errors.New("invalid token type")
	errInvalidSession        = errors.New("invalid session id in token")
	errInvalidRefreshFormat  = errors.New("invalid refresh token format")
	errInvalidRefreshToken   = errors.New("invalid refresh token")
	errInvalidRefreshSession = errors.New("invalid refresh session")
	errUserNotFound          = errors.New("user not found")
	errApiKeyMissing         = errors.New("API key missing")
	errInvalidApiKey         = errors.New("invalid API key")
	errAccessDenied          = errors.New("access denied")
	ApiKeyTypeHeader         = "X-Api-Key-Type"
)

type AuthMiddleware struct {
	userUC    *client.UseCaseI
	sessionUC *session.UseCaseI
	authzUC   *authz.UseCase
	cfg       *config.Config
	l         logger.Log
	pubKey    *rsa.PublicKey
}

func NewAuthMiddleware(u *usecase.UseCase, cfg *config.Config, l logger.Log) *AuthMiddleware {
	pubKey, err := jwt.ParseRSAPublicKey(cfg.JWT.PublicKey)
	if err != nil {
		l.Errorw("AuthMiddleware - NewAuthMiddleware - ParseRSAPublicKey", "error", err)
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

func (m *AuthMiddleware) validateAccessToken(ctx *gin.Context) (*domain.Session, error) {
	tokenStr := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	if tokenStr == "" {
		tokenStr = ExtractBearerToken(ctx)
	}

	if tokenStr == "" {
		return nil, errUnAuth
	}

	metadata, err := m.parseAndValidateMetadata(tokenStr)
	if err != nil {
		return nil, err
	}

	sessionID, err := uuid.Parse(metadata.SessionID)
	if err != nil {
		m.l.Warnw("AuthMiddleware - validateAccessToken - Invalid session ID", "session_id", metadata.SessionID)
		return nil, errInvalidSession
	}

	session, err := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Errorw("AuthMiddleware - validateAccessToken - Get", "error", err)
		return nil, errRevokedToken
	}

	return session, nil
}

func (m *AuthMiddleware) parseAndValidateMetadata(tokenStr string) (*jwt.AccessTokenClaims, error) {
	metadata, err := jwt.ParseAccessToken(tokenStr, m.pubKey, m.cfg.JWT.Issuer, "")
	if err != nil {
		if errors.Is(err, jwt.ErrAccessTokenExpired) {
			return nil, errExpiredToken
		}
		return nil, errInvalidToken
	}

	if metadata.Issuer != m.cfg.JWT.Issuer {
		m.l.Warnw("AuthMiddleware - parseAndValidateMetadata - Invalid issuer", "issuer", metadata.Issuer)
		return nil, errInvalidIssuer
	}

	if metadata.Type != "access" {
		m.l.Warnw("AuthMiddleware - parseAndValidateMetadata - Invalid token type", "type", metadata.Type)
		return nil, errInvalidType
	}

	return metadata, nil
}

func (m *AuthMiddleware) AuthClientAccess(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, err.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 5. Context injection
	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	ctx.Next()
}

func (m *AuthMiddleware) AuthClientRefresh(ctx *gin.Context) {
	token := cookie.GetCookie(ctx, consts.COOKIE_REFRESH_TOKEN)
	if token == "" {
		token = ExtractBearerToken(ctx)
	}

	if token == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Parse the refresh token
	rt, err := jwt.ParseRefreshToken(token)
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - invalid refresh token format", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidRefreshFormat.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Get session from database
	sessionID, err := uuid.Parse(rt.SessionID)
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - invalid session id in token", "session_id", rt.SessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidSession.Error(), nil, false)
		ctx.Abort()
		return
	}

	session, err := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - session not found", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidRefreshSession.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Verify the refresh token
	if !rt.Verify(session.RefreshTokenHash) {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - invalid refresh token hash", "session_id", sessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidRefreshToken.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Check if session is still valid
	if session.Revoked || session.IsExpired() {
		m.l.Errorw("AuthMiddleware - AuthClientRefresh - session revoked or expired", "session_id", sessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errRevokedToken.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Set session ID in context for next handler
	ctx.Set(consts.CtxSessionID, rt.SessionID)
	ctx.Set(consts.CtxRefreshToken, token)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())

	ctx.Next()
}

func (m *AuthMiddleware) AuthAdmin(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, err.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 5. Admin check - ensure user exists and has admin role
	user, err := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthAdmin - User Get", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUserNotFound.Error(), nil, false)
		ctx.Abort()
		return
	}

	if user.RoleID == nil {
		m.l.Warnw("AuthMiddleware - AuthAdmin - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, errAccessDenied.Error(), nil, false)
		ctx.Abort()
		return
	}

	role, err := m.authzUC.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
	if err != nil {
		m.l.Errorw("AuthMiddleware - AuthAdmin - Role Get", "error", err)
		response.ControllerResponse(ctx, http.StatusForbidden, errAccessDenied.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Check if role name contains "admin" (case-insensitive)
	if !strings.Contains(strings.ToLower(role.Name), "admin") {
		m.l.Warnw("AuthMiddleware - AuthAdmin - User is not admin", "user_id", user.ID, "role", role.Name)
		response.ControllerResponse(ctx, http.StatusForbidden, errAccessDenied.Error(), nil, false)
		ctx.Abort()
		return
	}

	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxSession, session)
	ctx.Set(consts.CtxUserID, session.UserID.String())
	ctx.Set(consts.CtxIsAdmin, true)

	ctx.Next()
}

func (m *AuthMiddleware) AuthApiKey(ctx *gin.Context) {
	apiKey := util.GetAPIKey(ctx)
	if apiKey == "" {
		m.l.Warnw("AuthMiddleware - AuthApiKey - API key missing", "ip", util.GetIPAddress(ctx))
		response.ControllerResponse(ctx, http.StatusUnauthorized, errApiKeyMissing.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Simple check against config - in real app, check DB or external service
	if apiKey != m.cfg.APIKeys.XApiKey {
		m.l.Warnw("AuthMiddleware - AuthApiKey - Invalid API key", "ip", util.GetIPAddress(ctx))
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidApiKey.Error(), nil, false)
		ctx.Abort()
		return
	}

	ctx.Set(consts.CtxApiKeyAuth, true)
	ctx.Next()
}

func (m *AuthMiddleware) Authz(ctx *gin.Context) {
	sessionVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		// Should have been set by AuthClientAccess
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
		ctx.Abort()
		return
	}

	session, ok := sessionVal.(*domain.Session)
	if !ok {
		m.l.Errorw("AuthMiddleware - Authz - session type cast error")
		response.ControllerResponse(ctx, http.StatusInternalServerError, "internal server error", nil, false)
		ctx.Abort()
		return
	}

	// 1. Get User to check RBAC role existence
	user, err := (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if err != nil {
		m.l.Errorw("AuthMiddleware - Authz - Get User", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUserNotFound.Error(), nil, false)
		ctx.Abort()
		return
	}

	if user.RoleID == nil {
		m.l.Warnw("AuthMiddleware - Authz - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, errAccessDenied.Error(), nil, false)
		ctx.Abort()
		return
	}

	path := ctx.FullPath()
	if path == "" {
		path = ctx.Request.URL.Path
	}
	method := ctx.Request.Method

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

	allowed, err := m.authzUC.Access.Check(ctx.Request.Context(), session.UserID, session, path, method, env)
	if err != nil {
		m.l.Errorw("AuthMiddleware - Authz - Check", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, "internal server error", nil, false)
		ctx.Abort()
		return
	}

	if !allowed {
		response.ControllerResponse(ctx, http.StatusForbidden, errAccessDenied.Error(), nil, false)
		ctx.Abort()
		return
	}

	ctx.Next()
}

func ExtractBearerToken(ctx *gin.Context) string {
	bearToken := util.GetAuthorization(ctx)
	token := bearToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BearerToken {
		return ""
	}

	return onlyToken[1]
}

func ExtractBasicToken(ctx *gin.Context) string {
	basicToken := util.GetAuthorization(ctx)
	token := basicToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BasicToken {
		return ""
	}

	return onlyToken[1]
}
