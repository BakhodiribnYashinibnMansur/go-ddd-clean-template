package middleware

import (
	"net/http"
	"strings"

	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/security/csrf"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CSRFMiddleware provides production-grade CSRF protection using HMAC signatures.
// Unlike simple "Double Submit", this validates the token integrity against a server-side secret.
type CSRFMiddleware struct {
	generator  *csrf.Generator // Cryptographic utility for token signing/verification.
	store      csrf.Store      // Storage interface (e.g. Redis) for tracking issued tokens.
	logger     logger.Log      // Standardized logger.
	cookieName string          // Name of the cookie containing the token.
	headerName string          // Name of the header expected to carry the token.
}

// NewCSRFMiddleware initializes the secure middleware component.
func NewCSRFMiddleware(generator *csrf.Generator, store csrf.Store, l logger.Log) *CSRFMiddleware {
	return &CSRFMiddleware{
		generator:  generator,
		store:      store,
		logger:     l,
		cookieName: consts.CookieCsrfToken,
		headerName: consts.HeaderXCSRFToken,
	}
}

// Protect enforces strict CSRF verification for all state-changing HTTP requests.
// It ensures that the token provided in the header matches the one in the cookie
// AND that the token is cryptographically valid and bound to the current session.
func (m *CSRFMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Filter: Skip check for safe methods (GET, HEAD, OPTIONS).
		if !isStateChangingMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Prerequisite: A valid session ID is required to bind the CSRF token.
		sessionID, err := httpx.GetCtxSessionID(c)
		if err != nil {
			m.rejectCSRF(c, nil, http.StatusUnauthorized, httpx.ErrSessionIDNotFound,
				"CSRF Middleware - No session ID in context")
			return
		}

		// 1. Extract Token from Cookie.
		cookieToken, err := c.Cookie(m.cookieName)
		if err != nil || cookieToken == "" {
			m.rejectCSRF(c, &sessionID, http.StatusForbidden, httpx.ErrCSRFMissing,
				"CSRF Middleware - Missing cookie token")
			return
		}

		// 2. Extract Token from Header.
		headerToken := c.GetHeader(m.headerName)
		if headerToken == "" {
			m.rejectCSRF(c, &sessionID, http.StatusForbidden, httpx.ErrCSRFMissing,
				"CSRF Middleware - Missing header token")
			return
		}

		// 3. Double-Submit Check: Tokens must match.
		if cookieToken != headerToken {
			m.rejectCSRF(c, &sessionID, http.StatusForbidden, httpx.ErrCSRFInvalid,
				"CSRF Middleware - Token mismatch")
			return
		}

		// 4. Retrieve Reference Token (Hash) from Storage.
		storedHash, expiresAt, err := m.store.Get(c.Request.Context(), sessionID.String())
		if err != nil {
			m.rejectCSRF(c, &sessionID, http.StatusForbidden, httpx.ErrCSRFInvalid,
				"CSRF Middleware - Token not found in store", "error", err)
			return
		}

		// 5. Cryptographic Validation (HMAC signature + expiration).
		if err := m.generator.ValidateToken(cookieToken, storedHash, sessionID.String(), expiresAt); err != nil {
			m.rejectCSRF(c, &sessionID, http.StatusForbidden, httpx.ErrCSRFInvalid,
				"CSRF Middleware - Token validation failed", "error", err)
			return
		}

		c.Next()
	}
}

// rejectCSRF logs a CSRF rejection with common context (ip, path, session) and
// aborts the request with the given status/error.
func (m *CSRFMiddleware) rejectCSRF(c *gin.Context, sessionID *uuid.UUID, status int, errMsg error, logMsg string, extra ...any) {
	kv := []any{"ip", httpx.GetIPAddress(c), "path", c.Request.URL.Path}
	if sessionID != nil {
		kv = append(kv, "session_id", *sessionID)
	}
	kv = append(kv, extra...)
	m.logger.Warnw(logMsg, kv...)
	response.ControllerResponse(c, status, errMsg, nil, false)
	c.Abort()
}

// HybridProtect applies logic to skip CSRF for authorized native clients (e.g. Mobile Apps),
// while strictly enforcing it for browser-based sessions.
func (m *CSRFMiddleware) HybridProtect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Native Client Detection: Presence of Authorization Bearer header.
		// These clients typically store tokens securely and aren't subject to browser cookie vulnerabilities.
		auth := httpx.GetAuthorization(c)
		if auth != "" && strings.HasPrefix(auth, consts.AuthBearer) {
			c.Next()
			return
		}

		// Default to Browser protection.
		m.Protect()(c)
	}
}

// isStateChangingMethod returns true if the HTTP verb implies a modification of server state.
func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
