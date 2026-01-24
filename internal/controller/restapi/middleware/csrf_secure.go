package middleware

import (
	"net/http"
	"strings"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/pkg/csrf"
	"gct/pkg/logger"
	"github.com/gin-gonic/gin"
)

// CSRFMiddleware provides production-grade CSRF protection
type CSRFMiddleware struct {
	generator  *csrf.Generator
	store      csrf.Store
	logger     logger.Log
	cookieName string
	headerName string
}

// NewCSRFMiddleware creates a new CSRF middleware with HMAC-based token validation
func NewCSRFMiddleware(generator *csrf.Generator, store csrf.Store, l logger.Log) *CSRFMiddleware {
	return &CSRFMiddleware{
		generator:  generator,
		store:      store,
		logger:     l,
		cookieName: consts.COOKIE_CSRF_TOKEN,
		headerName: consts.HeaderXCSRFToken,
	}
}

// Protect enforces CSRF protection for state-changing requests
func (m *CSRFMiddleware) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check for state-changing methods
		if !isStateChangingMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Get session ID from context (set by auth middleware)
		sessionID, err := util.GetCtxSessionID(c)
		if err != nil {
			m.logger.WithContext(c.Request.Context()).Warnw("CSRF Middleware - No session ID in context",
				"ip", util.GetIPAddress(c),
				"path", c.Request.URL.Path)
			response.ControllerResponse(c, http.StatusUnauthorized, "unauthorized", nil, false)
			c.Abort()
			return
		}

		// Get token from cookie
		cookieToken, err := c.Cookie(m.cookieName)
		if err != nil || cookieToken == "" {
			m.logger.WithContext(c.Request.Context()).Warnw("CSRF Middleware - Missing cookie token",
				"ip", util.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID)
			response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFMissing.Error(), nil, false)
			c.Abort()
			return
		}

		// Get token from header
		headerToken := c.GetHeader(m.headerName)
		if headerToken == "" {
			m.logger.WithContext(c.Request.Context()).Warnw("CSRF Middleware - Missing header token",
				"ip", util.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID)
			response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFMissing.Error(), nil, false)
			c.Abort()
			return
		}

		// Double submit check: cookie and header must match
		if cookieToken != headerToken {
			m.logger.WithContext(c.Request.Context()).Warnw("CSRF Middleware - Token mismatch",
				"ip", util.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID)
			response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFInvalid.Error(), nil, false)
			c.Abort()
			return
		}

		// Get stored token hash from store
		storedHash, expiresAt, err := m.store.Get(c.Request.Context(), sessionID.String())
		if err != nil {
			m.logger.WithContext(c.Request.Context()).Warnw("CSRF Middleware - Token not found in store",
				"ip", util.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID,
				"error", err)
			response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFInvalid.Error(), nil, false)
			c.Abort()
			return
		}

		// Validate token using HMAC
		if err := m.generator.ValidateToken(cookieToken, storedHash, sessionID.String(), expiresAt); err != nil {
			m.logger.WithContext(c.Request.Context()).Warnw("CSRF Middleware - Token validation failed",
				"ip", util.GetIPAddress(c),
				"path", c.Request.URL.Path,
				"session_id", sessionID,
				"error", err)
			response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFInvalid.Error(), nil, false)
			c.Abort()
			return
		}

		c.Next()
	}
}

// HybridProtect skips CSRF for JWT clients, enforces for cookie clients
func (m *CSRFMiddleware) HybridProtect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// JWT client (Mobile/Desktop App with Bearer token) -> skip CSRF
		auth := util.GetAuthorization(c)
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			c.Next()
			return
		}

		// Cookie client (Web Browser) -> enforce CSRF
		m.Protect()(c)
	}
}

// isStateChangingMethod checks if the HTTP method modifies state
func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
