package middleware

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/consts"
	"github.com/evrone/go-clean-template/internal/controller/restapi/cookie"
	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/internal/usecase/user/client"
	"github.com/evrone/go-clean-template/internal/usecase/user/session"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	errUnAuth        = errors.New("unauthorized. token is missing")
	errInvalidToken  = errors.New("unauthorized. token is invalid")
	errExpiredToken  = errors.New("unauthorized. token is expired")
	errRevokedToken  = errors.New("unauthorized. token is revoked")
	ApiKeyTypeHeader = "X-Api-Key-Type"
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

func (m *AuthMiddleware) AuthClientAccess(ctx *gin.Context) {
	tokenStr := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	if tokenStr == "" {
		tokenStr = ExtractBearerToken(ctx)
	}

	if tokenStr == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 1. Signature & Basic Validation (RS256)
	// Use empty audience since it's not configured
	metadata, err := jwt.ParseAccessToken(tokenStr, m.pubKey, m.cfg.JWT.Issuer, "")
	if err != nil {
		if errors.Is(err, jwt.ErrAccessTokenExpired) {
			response.ControllerResponse(ctx, http.StatusUnauthorized, errExpiredToken.Error(), nil, false)
		} else {
			response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidToken.Error(), nil, false)
		}
		ctx.Abort()
		return
	}

	// 2. Issuer check
	if metadata.Issuer != m.cfg.JWT.Issuer {
		m.l.Warn("AuthMiddleware - AuthClientAccess - Invalid issuer", "issuer", metadata.Issuer)
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid issuer", nil, false)
		ctx.Abort()
		return
	}

	// 3. Type check (must be access)
	if metadata.Type != "access" {
		m.l.Warn("AuthMiddleware - AuthClientAccess - Invalid token type", "type", metadata.Type)
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid token type", nil, false)
		ctx.Abort()
		return
	}

	// 4. Session & Revocation check (Stateful)
	sessionID, err := uuid.Parse(metadata.SessionID)
	if err != nil {
		m.l.Warn("AuthMiddleware - AuthClientAccess - Invalid session ID", "session_id", metadata.SessionID)
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid session id in token", nil, false)
		ctx.Abort()
		return
	}

	session, err := (*m.sessionUC).GetByID(ctx, &domain.SessionFilter{ID: sessionID})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Error("AuthMiddleware - AuthClientAccess - GetByID", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errRevokedToken.Error(), nil, false)
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
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid refresh token", nil, false)
		ctx.Abort()
		return
	}

	// Get session from database
	session, err := (*m.sessionUC).GetByID(ctx, &domain.SessionFilter{ID: uuid.MustParse(rt.ID)})
	if err != nil {
		m.l.Error("AuthMiddleware - AuthClientRefresh - session not found", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid refresh session", nil, false)
		ctx.Abort()
		return
	}

	// Verify the refresh token
	if !rt.Verify(session.RefreshTokenHash) {
		m.l.Error("AuthMiddleware - AuthClientRefresh - invalid refresh token hash")
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid refresh token", nil, false)
		ctx.Abort()
		return
	}

	// Check if session is still valid
	if session.Revoked || session.IsExpired() {
		m.l.Error("AuthMiddleware - AuthClientRefresh - session revoked or expired")
		response.ControllerResponse(ctx, http.StatusUnauthorized, "session expired or revoked", nil, false)
		ctx.Abort()
		return
	}

	// Set the session ID in context for the next handler
	ctx.Set(consts.CtxSessionID, rt.ID)
	ctx.Set(consts.CtxUserID, session.UserID)

	ctx.Next()
}

func (m *AuthMiddleware) AuthAdmin(ctx *gin.Context) {
	tokenStr := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	if tokenStr == "" {
		tokenStr = ExtractBearerToken(ctx)
	}

	if tokenStr == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 1. Signature & Basic Validation (RS256)
	// Use empty audience since it's not configured
	metadata, err := jwt.ParseAccessToken(tokenStr, m.pubKey, m.cfg.JWT.Issuer, "")
	if err != nil {
		m.l.Error("AuthMiddleware - AuthAdmin - ParseAccessToken", err)
		if errors.Is(err, jwt.ErrAccessTokenExpired) {
			response.ControllerResponse(ctx, http.StatusUnauthorized, errExpiredToken.Error(), nil, false)
		} else {
			response.ControllerResponse(ctx, http.StatusUnauthorized, errInvalidToken.Error(), nil, false)
		}
		ctx.Abort()
		return
	}

	// 2. Issuer check
	if metadata.Issuer != m.cfg.JWT.Issuer {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid issuer", nil, false)
		ctx.Abort()
		return
	}

	// 3. Type check (must be access)
	if metadata.Type != "access" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid token type", nil, false)
		ctx.Abort()
		return
	}

	// 4. Session & Revocation check (Stateful)
	sessionID, err := uuid.Parse(metadata.SessionID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid session id in token", nil, false)
		ctx.Abort()
		return
	}

	session, err := (*m.sessionUC).GetByID(ctx, &domain.SessionFilter{ID: sessionID})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Error("AuthMiddleware - AuthAdmin - GetByID", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, errRevokedToken.Error(), nil, false)
		ctx.Abort()
		return
	}

	// 5. Admin check - for now, just ensure user exists
	// TODO: Implement proper role-based access control
	_, err = (*m.userUC).User(ctx, client.UserInput{ID: session.UserID})
	if err != nil {
		m.l.Error("AuthMiddleware - AuthAdmin - User", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, "user not found", nil, false)
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
		response.ControllerResponse(ctx, http.StatusUnauthorized, "API key missing", nil, false)
		ctx.Abort()
		return
	}

	// Simple check against config - in real app, check DB or external service
	if apiKey != m.cfg.APIKeys.XApiKey {
		m.l.Warn("AuthMiddleware - AuthApiKey - Invalid API key", "ip", ctx.ClientIP())
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid API key", nil, false)
		ctx.Abort()
		return
	}

	ctx.Set("api_key_authenticated", true)
	ctx.Next()
}

func ExtractBearerToken(ctx *gin.Context) string {
	bearToken := ctx.GetHeader(consts.AuthorizationHeader)
	token := fmt.Sprintf("%v", bearToken)
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BearerToken {
		return ""
	}

	return onlyToken[1]
}

func ExtractBasicToken(ctx *gin.Context) string {
	basicToken := ctx.GetHeader(consts.AuthorizationHeader)
	token := fmt.Sprintf("%v", basicToken)
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != "Basic" {
		return ""
	}

	return onlyToken[1]
}
