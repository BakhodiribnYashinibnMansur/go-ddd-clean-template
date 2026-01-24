package middleware

import (
	"errors"
	"strings"

	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validateAccessToken is a private helper that extracts, parses, and verifies an access token.
// Returns the corresponding session if valid, or an error otherwise.
//
// This method handles both cookie-based (browser) and header-based (mobile/API) authentication.
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
//
// This method performs cryptographic signature verification using RSA public key
// and validates JWT standard claims (issuer, type, expiration).
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

// ExtractBearerToken parses the "Bearer <token>" string from authorization headers.
//
// Returns empty string if the authorization header is missing or malformed.
// Expected format: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
func ExtractBearerToken(ctx *gin.Context) string {
	bearToken := util.GetAuthorization(ctx)
	token := bearToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BearerToken {
		return ""
	}

	return onlyToken[1]
}

// ExtractBasicToken parses the "Basic <token>" string from authorization headers.
//
// Returns empty string if the authorization header is missing or malformed.
// Expected format: "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
func ExtractBasicToken(ctx *gin.Context) string {
	basicToken := util.GetAuthorization(ctx)
	token := basicToken
	onlyToken := strings.Split(token, " ")

	if len(onlyToken) != 2 || onlyToken[0] != consts.BasicToken {
		return ""
	}

	return onlyToken[1]
}
