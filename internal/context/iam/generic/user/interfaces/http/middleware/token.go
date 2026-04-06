package middleware

import (
	"errors"
	"net/http"

	"gct/internal/context/iam/generic/user/application/query"
	userentity "gct/internal/context/iam/generic/user/domain/entity"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/cookie"
	"gct/internal/kernel/infrastructure/security/apikey"
	"gct/internal/kernel/infrastructure/security/apikeythrottle"
	"gct/internal/kernel/infrastructure/security/audit"
	"gct/internal/kernel/infrastructure/security/clockskew"
	"gct/internal/kernel/infrastructure/security/jwt"
	"gct/internal/kernel/infrastructure/security/tbh"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validateAccessToken extracts, parses, and verifies an access token.
// Returns the corresponding AuthSession if valid, or an error otherwise.
//
// Token extraction follows a dual strategy:
// 1. HTTP-Only Cookie (preferred for web/browser clients for XSS protection)
// 2. Authorization Header (for mobile, CLI, and API clients)
//
// The caller MUST present an X-API-Key header identifying the integration the
// request belongs to. The audience and RSA public key used to verify the JWT
// are pulled from the resolved integration — not from global config — so
// tokens issued by one integration cannot be replayed against another.
func (m *AuthMiddleware) validateAccessToken(ctx *gin.Context) (*shared.AuthSession, error) {
	ip := httpx.GetIPAddress(ctx)
	ua := ctx.Request.UserAgent()

	// API key throttle: reject IPs that are brute-forcing API keys.
	if m.apiKeyThrottle != nil {
		if err := m.apiKeyThrottle.Check(ctx.Request.Context(), ip); err != nil {
			if errors.Is(err, apikeythrottle.ErrThrottled) {
				m.auditLogger.Log(ctx.Request.Context(), audit.Entry{
					Event: audit.EventAPIKeyScraping, IPAddress: ip, UserAgent: ua,
					Metadata: map[string]any{"reason": "ip_blocked"},
				})
				return nil, httpx.ErrTooManyRequests
			}
		}
	}

	apiKey := httpx.GetAPIKey(ctx)
	if apiKey == "" {
		return nil, httpx.ErrUnAuth
	}
	resolved, err := m.resolver.Resolve(ctx.Request.Context(), apiKey)
	if err != nil || resolved == nil {
		// Record API key resolve failure for throttle tracking.
		if m.apiKeyThrottle != nil {
			_ = m.apiKeyThrottle.RecordFail(ctx.Request.Context(), ip)
		}
		m.auditLogger.Log(ctx.Request.Context(), audit.Entry{
			Event: audit.EventAPIKeyMismatch, IPAddress: ip, UserAgent: ua,
			Metadata: map[string]any{"api_key": apikey.Mask(apiKey)},
		})
		return nil, httpx.ErrUnAuth
	}

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
	metadata, err := m.parseAndValidateMetadata(tokenStr, resolved)
	if err != nil {
		return nil, err
	}

	// Clock-skew observation: record the difference between the token's iat
	// and server time so we can detect drifting integration clocks.
	if metadata.IssuedAt != nil {
		clockskew.Observe(resolved.Name, metadata.IssuedAt.Time)
	}

	// Logical verification: Is the session active and known?
	sessionID, err := uuid.Parse(metadata.SessionID)
	if err != nil {
		m.l.Warnw("AuthMiddleware - validateAccessToken - Invalid session ID", "session_id", metadata.SessionID)
		return nil, httpx.ErrInvalidSession
	}

	// Revocation check: fast-path deny via Redis denylist.
	if m.revStore != nil && m.revStore.IsRevoked(ctx.Request.Context(), sessionID.String()) {
		m.l.Warnw("AuthMiddleware - validateAccessToken - session revoked in Redis",
			"session_id", sessionID)
		return nil, httpx.ErrRevokedToken
	}

	session, err := m.findSession.Handle(ctx.Request.Context(), query.FindSessionQuery{SessionID: userentity.SessionID(sessionID)})
	if err != nil || session.Revoked || session.IsExpired() {
		m.l.Errorw("AuthMiddleware - validateAccessToken - FindSession", "error", err)
		return nil, httpx.ErrRevokedToken
	}

	// Cross-integration defence: a token signed for one audience must never
	// authenticate a session that was bound to a different audience.
	if session.IntegrationName != resolved.Name {
		m.l.Warnw("AuthMiddleware - validateAccessToken - integration_mismatch",
			"session_integration", session.IntegrationName,
			"resolved_integration", resolved.Name)
		sid := session.ID
		m.auditLogger.Log(ctx.Request.Context(), audit.Entry{
			Event: audit.EventCrossIntegration, IPAddress: ip, UserAgent: ua,
			SessionID: &sid, UserID: &session.UserID,
			Metadata: map[string]any{
				"session_integration":  session.IntegrationName,
				"resolved_integration": resolved.Name,
			},
		})
		return nil, httpx.ErrInvalidToken
	}

	// TBH (Token-Binding Hash) verification — backward-compatible.
	// Only checked when the claim is present in the token AND the
	// integration's binding mode is not "off".
	if metadata.TBH != "" && len(m.tbhPepper) > 0 && resolved.BindingMode != "off" {
		if !tbh.Verify(m.tbhPepper, ip, ua, metadata.TBH) {
			sid := session.ID
			m.auditLogger.Log(ctx.Request.Context(), audit.Entry{
				Event: audit.EventTBHMismatch, IPAddress: ip, UserAgent: ua,
				SessionID: &sid, UserID: &session.UserID,
				Metadata: map[string]any{"binding_mode": resolved.BindingMode},
			})
			if resolved.BindingMode == "strict" {
				return nil, httpx.ErrInvalidToken
			}
			// "warn" mode: log only, allow through.
			m.l.Warnw("AuthMiddleware - validateAccessToken - TBH mismatch (warn mode)",
				"session_id", sessionID, "binding_mode", resolved.BindingMode)
		}
	}

	return session, nil
}

// errTooManyRequests is used for the 429 sentinel that validateAccessToken
// returns when the API key throttle fires.
var _ = http.StatusTooManyRequests // keep net/http import used

// parseAndValidateMetadata performs full cryptographic and claim validation
// of an access token against the current integration's public key, falling
// back to the previous key if the token carries the previous kid (key
// rotation window).
func (m *AuthMiddleware) parseAndValidateMetadata(tokenStr string, resolved *ResolvedForVerify) (*jwt.AccessTokenClaims, error) {
	metadata, err := jwt.ParseAccessToken(tokenStr, resolved.PublicKey, m.issuer, resolved.Name, m.leeway)
	if err == nil {
		return metadata, nil
	}

	// Expired tokens are never retried — return immediately.
	if errors.Is(err, jwt.ErrAccessTokenExpired) {
		return nil, httpx.ErrExpiredToken
	}

	// Rotation window: if the token's kid matches the previous key, retry.
	if resolved.PreviousPublicKey != nil && resolved.PreviousKeyID != "" {
		if kid := peekKID(tokenStr); kid != "" && kid == resolved.PreviousKeyID {
			if md, retryErr := jwt.ParseAccessToken(tokenStr, resolved.PreviousPublicKey, m.issuer, resolved.Name, m.leeway); retryErr == nil {
				return md, nil
			} else if errors.Is(retryErr, jwt.ErrAccessTokenExpired) {
				return nil, httpx.ErrExpiredToken
			}
		}
	}

	return nil, httpx.ErrInvalidToken
}

// peekKID unverified-parses a JWT purely to read its `kid` header. It does NOT
// validate signature, expiry, or any claim — the caller MUST verify the
// token independently with ParseAccessToken afterwards.
func peekKID(tokenStr string) string {
	parser := jwtv5.NewParser()
	tok, _, err := parser.ParseUnverified(tokenStr, jwtv5.MapClaims{})
	if err != nil || tok == nil {
		return ""
	}
	if v, ok := tok.Header["kid"].(string); ok {
		return v
	}
	return ""
}
