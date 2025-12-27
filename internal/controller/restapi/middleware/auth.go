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
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
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
	metadata, err := jwt.ParseToken(tokenStr, m.pubKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
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

	session, err := (*m.sessionUC).GetByID(ctx, sessionID)
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
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
		ctx.Abort()
		return
	}

	if token == "" {
		token = ExtractBearerToken(ctx)
	}

	// Refresh logic using sessionUC
	sessionID, err := uuid.Parse(token)
	if err == nil {
		_, err := (*m.sessionUC).GetByID(ctx, sessionID)
		if err != nil {
			response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid refresh session", nil, false)
			ctx.Abort()
			return
		}
	}

	ctx.Next()
}

func (m *AuthMiddleware) AuthAdmin(ctx *gin.Context) {
	token := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	header := ctx.GetHeader(consts.AuthorizationHeader)

	if header == "" && token == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, errUnAuth.Error(), nil, false)
		ctx.Abort()
		return
	}

	if token == "" {
		token = ExtractBearerToken(ctx)
	}

	// Admin check logic
	// For now, it just continues if token exists.
	// Real implementation would check role in JWT or DB.

	ctx.Next()
}

func (m *AuthMiddleware) AuthApiKey(ctx *gin.Context) {
	apiKey := ctx.GetHeader("X-API-KEY")
	if apiKey == "" {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "API key missing", nil, false)
		ctx.Abort()
		return
	}

	// Simple check against config - in real app, check DB or external service
	if apiKey != m.cfg.APIKeys.XApiKey {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid API key", nil, false)
		ctx.Abort()
		return
	}

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
