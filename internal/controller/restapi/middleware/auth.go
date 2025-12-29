package middleware

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/usecase/user/client"
	"gct/internal/usecase/user/session"
	"gct/pkg/jwt"
	"gct/pkg/logger"
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
	ApiKeyTypeHeader         = "X-Api-Key-Type"
)

type AuthMiddleware struct {
	userUC    *client.UseCaseI
	sessionUC *session.UseCaseI
	cfg       *config.Config
	l         logger.Log
	pubKey    *rsa.PublicKey
}

func NewAuthMiddleware(u *usecase.UseCase, cfg *config.Config, l logger.Log) *AuthMiddleware {
	pubKey, err := jwt.ParseRSAPublicKey(cfg.JWT.PublicKey)
	if err != nil {
		l.Error("AuthMiddleware - NewAuthMiddleware - ParseRSAPublicKey", err)
	}

	return &AuthMiddleware{
		userUC:    &u.User.Client,
		sessionUC: &u.User.Session,
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
		m.l.Warn("AuthMiddleware - validateAccessToken - Invalid session ID", "session_id", metadata.SessionID)
		return nil, errInvalidSession
	}

	session, err := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Error("AuthMiddleware - validateAccessToken - Get", err)
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
		m.l.Warn("AuthMiddleware - parseAndValidateMetadata - Invalid issuer", "issuer", metadata.Issuer)
		return nil, errInvalidIssuer
	}

	if metadata.Type != "access" {
		m.l.Warn("AuthMiddleware - parseAndValidateMetadata - Invalid token type", "type", metadata.Type)
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
	ctx.Set(consts.CtxUserID, session.UserID)

	ctx.Next()
}

func (m *AuthMiddleware) AuthClientRefresh(ctx *gin.Context) {
	token := cookie.GetCookie(ctx, consts.COOKIE_REFRESH_TOKEN)
	header := ctx.GetHeader(consts.AuthorizationHeader)

	if header == "" && token == "" {
		token = ExtractBearerToken(ctx)
		if token == "" {
			response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
			ctx.Abort()
			return
		}
	}

	// Parse the refresh token
	rt, err := jwt.ParseRefreshToken(token)
	if err != nil {
		m.l.Error("AuthMiddleware - AuthClientRefresh - invalid refresh token format", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidRefreshFormat.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Get session from database
	sessionID := uuid.MustParse(rt.ID)
	session, err := (*m.sessionUC).Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		m.l.Error("AuthMiddleware - AuthClientRefresh - session not found", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidRefreshSession.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Verify the refresh token
	if !rt.Verify(session.RefreshTokenHash) {
		m.l.Error("AuthMiddleware - AuthClientRefresh - invalid refresh token hash")
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidRefreshToken.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Check if session is still valid
	if session.Revoked || session.IsExpired() {
		m.l.Error("AuthMiddleware - AuthClientRefresh - session revoked or expired")
		response.ControllerResponse(ctx, http.StatusUnauthorized, errRevokedToken.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Set the session ID in context for the next handler
	ctx.Set(consts.CtxSessionID, rt.ID)
	ctx.Set(consts.CtxUserID, session.UserID)

	ctx.Next()
}

func (m *AuthMiddleware) AuthAdmin(ctx *gin.Context) {
	session, err := m.validateAccessToken(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, err.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 5. Admin check - for now, just ensure user exists
	// TODO: Implement proper role-based access control
	_, err = (*m.userUC).Get(ctx, &domain.UserFilter{ID: &session.UserID})
	if err != nil {
		m.l.Error("AuthMiddleware - AuthAdmin - User", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUserNotFound.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 6. Context injection
	ctx.Set(consts.CtxSessionID, session.ID)
	ctx.Set(consts.CtxUserID, session.UserID)
	ctx.Set("is_admin", true)

	ctx.Next()
}

func (m *AuthMiddleware) AuthApiKey(ctx *gin.Context) {
	apiKey := ctx.GetHeader("X-API-KEY")
	if apiKey == "" {
		m.l.Warn("AuthMiddleware - AuthApiKey - API key missing", "ip", ctx.ClientIP())
		response.ControllerResponse(ctx, http.StatusUnauthorized, errApiKeyMissing.Error(), nil, false)
		ctx.Abort()
		return
	}

	// Simple check against config - in real app, check DB or external service
	if apiKey != m.cfg.APIKeys.XApiKey {
		m.l.Warn("AuthMiddleware - AuthApiKey - Invalid API key", "ip", ctx.ClientIP())
		response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidApiKey.Error(), nil, false)
		ctx.Abort()
		return
	}

	ctx.Set("api_key_authenticated", true)
	ctx.Next()
}

func ExtractBearerToken(ctx *gin.Context) string {
	bearToken := ctx.GetHeader(consts.AuthorizationHeader)
	token := bearToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BearerToken {
		return ""
	}

	return onlyToken[1]
}

func ExtractBasicToken(ctx *gin.Context) string {
	basicToken := ctx.GetHeader(consts.AuthorizationHeader)
	token := basicToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != "Basic" {
		return ""
	}

	return onlyToken[1]
}
