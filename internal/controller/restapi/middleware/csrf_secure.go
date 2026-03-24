package middleware

import (
	"net/http"
	"strings"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/security/csrf"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
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
			m.logger.Warnw("CSRF Middleware - No session ID in context",
				"ip", httpx.GetIPAddress(c),
				"path", c.Request.URL.Path)
			response.ControllerResponse(c, http.StatusUnauthorized, httpx.ErrSessionIDNotFound, nil, false)
			c.Abort()
			return
		}

		// 1. Extract Token from Cookie.
		cookieToken, err := c.Cookie(m.cookieName)
		if err != nil || cookieToken == "" {
			m.logger.Warnw("CSRF Middleware - Missing cookie token",
				"ip", httpx.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID)
			response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFMissing, nil, false)
			c.Abort()
			return
		}

		// 2. Extract Token from Header.
		headerToken := c.GetHeader(m.headerName)
		if headerToken == "" {
			m.logger.Warnw("CSRF Middleware - Missing header token",
				"ip", httpx.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID)
			response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFMissing, nil, false)
			c.Abort()
			return
		}

		// 3. Double-Submit Check: Tokens must match.
		if cookieToken != headerToken {
			m.logger.Warnw("CSRF Middleware - Token mismatch",
				"ip", httpx.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID)
			response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFInvalid, nil, false)
			c.Abort()
			return
		}

		// 4. Retrieve Reference Token (Hash) from Storage.
		storedHash, expiresAt, err := m.store.Get(c.Request.Context(), sessionID.String())
		if err != nil {
			m.logger.Warnw("CSRF Middleware - Token not found in store",
				"ip", httpx.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID,
				"error", err)
			response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFInvalid, nil, false)
			c.Abort()
			return
		}

		// 5. Cryptographic Validation.
		// Verifies the HMAC signature and expiration time.
		if err := m.generator.ValidateToken(cookieToken, storedHash, sessionID.String(), expiresAt); err != nil {
			m.logger.Warnw("CSRF Middleware - Token validation failed",
				"ip", httpx.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID,
				"error", err)
			response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFInvalid, nil, false)
			c.Abort()
			return
		}

		c.Next()
	}
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
