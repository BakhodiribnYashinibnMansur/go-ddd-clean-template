package middleware

import (
	"errors"

	"gct/internal/context/iam/generic/user/application/query"
	userdomain "gct/internal/context/iam/generic/user/domain"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/cookie"
	"gct/internal/kernel/infrastructure/security/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validateAccessToken extracts, parses, and verifies an access token.
// Returns the corresponding AuthSession if valid, or an error otherwise.
//
// Token extraction follows a dual strategy:
// 1. HTTP-Only Cookie (preferred for web/browser clients for XSS protection)
// 2. Authorization Header (for mobile, CLI, and API clients)
func (m *AuthMiddleware) validateAccessToken(ctx *gin.Context) (*shared.AuthSession, error) {
	// Strategy 1: HTTP-Only Cookie (common for Web/Browser clients)
	tokenStr := cookie.GetCookie(ctx, consts.CookieAccessToken)
	// Strategy 2: Authorization Header (common for Native/Mobile/CLI clients)
	if tokenStr == "" {
		authHeader := httpx.GetAuthorization(ctx)
		tokenStr = httpx.ExtractBearerToken(authHeader)
	}

	if tokenStr == "" {
		return nil, httpx.ErrUnAuth
	}

	// Parsing and cryptographic verification
	metadata, err := m.parseAndValidateMetadata(tokenStr)
	if err != nil {
		return nil, err
	}

	// Logical verification: Is the session active and known?
	sessionID, err := uuid.Parse(metadata.SessionID)
	if err != nil {
		m.l.Warnw("AuthMiddleware - validateAccessToken - Invalid session ID", "session_id", metadata.SessionID)
		return nil, httpx.ErrInvalidSession
	}

	session, err := m.findSession.Handle(ctx.Request.Context(), query.FindSessionQuery{SessionID: userdomain.SessionID(sessionID)})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Errorw("AuthMiddleware - validateAccessToken - FindSession", "error", err)
		return nil, httpx.ErrRevokedToken
	}

	return session, nil
}

// parseAndValidateMetadata performs full cryptographic and claim validation
// of an access token. Issuer, audience, expiry, and issued-at are enforced by
// the v5 parser options; we only need to additionally validate the custom
// "typ" claim (done inside jwt.ParseAccessToken).
func (m *AuthMiddleware) parseAndValidateMetadata(tokenStr string) (*jwt.AccessTokenClaims, error) {
	metadata, err := jwt.ParseAccessToken(tokenStr, m.pubKey, m.cfg.JWT.Issuer, m.audience, m.leeway)
	if err != nil {
		if errors.Is(err, jwt.ErrAccessTokenExpired) {
			return nil, httpx.ErrExpiredToken
		}
		// ParseAccessToken already enforces typ == TokenTypeAccess. A wrong
		// "typ" surfaces here as ErrAccessTokenInvalid wrapping our own error.
		return nil, httpx.ErrInvalidToken
	}
	return metadata, nil
}
