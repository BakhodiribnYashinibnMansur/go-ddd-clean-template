package auth

import (
	"errors"

	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	"gct/internal/domain"
	"gct/pkg/httpx"
	"gct/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validateAccessToken is a private helper that extracts, parses, and verifies an access token.
// Returns the corresponding session if valid, or an error otherwise.
//
// This method implements a dual-strategy token extraction:
// 1. HTTP-Only Cookie (preferred for web/browser clients for XSS protection)
// 2. Authorization Header (for mobile, CLI, and API clients)
func (m *AuthMiddleware) validateAccessToken(ctx *gin.Context) (*domain.Session, error) {
	// Strategy 1: HTTP-Only Cookie (common for Web/Browser clients)
	tokenStr := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
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

	session, err := m.sessionuc.Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Errorw("AuthMiddleware - validateAccessToken - Get", "error", err)
		return nil, httpx.ErrRevokedToken
	}

	return session, nil
}

// parseAndValidateMetadata parses the raw JWT string and validates core claims like issuer and token type.
//
// This method performs cryptographic signature verification using RSA public key
// and validates business logic claims (issuer, token type, expiration).
func (m *AuthMiddleware) parseAndValidateMetadata(tokenStr string) (*jwt.AccessTokenClaims, error) {
	metadata, err := jwt.ParseAccessToken(tokenStr, m.pubKey, m.cfg.JWT.Issuer, "")
	if err != nil {
		if errors.Is(err, jwt.ErrAccessTokenExpired) {
			return nil, httpx.ErrExpiredToken
		}
		return nil, httpx.ErrInvalidToken
	}

	if metadata.Issuer != m.cfg.JWT.Issuer {
		m.l.Warnw("AuthMiddleware - parseAndValidateMetadata - Invalid issuer", "issuer", metadata.Issuer)
		return nil, httpx.ErrInvalidIssuer
	}

	if metadata.Type != consts.TokenTypeAccess {
		m.l.Warnw("AuthMiddleware - parseAndValidateMetadata - Invalid token type", "type", metadata.Type)
		return nil, httpx.ErrInvalidType
	}

	return metadata, nil
}
