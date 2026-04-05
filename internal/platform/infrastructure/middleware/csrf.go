package middleware

import (
	"net/http"

	"gct/internal/platform/domain/consts"
	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/platform/infrastructure/httpx"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Middleware enforces CSRF protection by comparing a token from a cookie with a token from a header.
// This implements the "Double Submit Cookie" pattern to prevent Cross-Site Request Forgery attacks.
func Middleware(l logger.Log, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Filter: Apply checks only to state-changing HTTP methods (POST, PUT, PATCH, DELETE).
		// Safe methods (GET, HEAD, OPTIONS) are read-only and bypass this check.
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			// 1. Retrieve the CSRF token from the browser cookie.
			cookieToken, err := c.Cookie(cookieName)
			if err != nil {
				l.Warnw("CSRF Middleware - Missing cookie token", "ip", httpx.GetIPAddress(c), "path", c.Request.URL.Path)
				response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFMissing, nil, false)
				c.Abort()
				return
			}

			// 2. Retrieve the CSRF token from the custom request header.
			headerToken := c.GetHeader(consts.HeaderXCSRFToken)

			// 3. Validation: Both tokens must exist and strictly match.
			if headerToken == "" || headerToken != cookieToken {
				l.Warnw("CSRF Middleware - Invalid or mismatched token",
					"ip", httpx.GetIPAddress(c),
					"path", c.Request.URL.Path,
					"headerEmpty", headerToken == "",
					"mismatch", headerToken != cookieToken)
				response.ControllerResponse(c, http.StatusForbidden, httpx.ErrCSRFInvalid, nil, false)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// HybridMiddleware intelligently applies CSRF protection based on the client type.
// - Native Clients (Mobile/Desktop) with "Authorization" headers are exempt (as they don't use cookies).
// - Browser Clients (Cookie-based auth) are strictly enforced to have CSRF protection.
func HybridMiddleware(l logger.Log, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check for Bearer token authorization (JWT flow).
		// Presence implies a non-browser client where CSRF is not a vector.
		auth := httpx.GetAuthorization(c)
		if auth != "" {
			c.Next()
			return
		}

		// 2. Fallback to standard Cookie-based flow (Browser).
		// Enforce CSRF checks.
		Middleware(l, cookieName)(c)
	}
}
