package middleware

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Middleware enforces CSRF protection by comparing a token from a cookie with a token from a header.
func Middleware(l logger.Log, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check for state-changing methods
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			cookieToken, err := c.Cookie(cookieName)
			if err != nil {
				l.Warnw("CSRF Middleware - Missing cookie token", "ip", util.GetIPAddress(c), "path", c.Request.URL.Path)
				response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFMissing.Error(), nil, false)
				c.Abort()
				return
			}

			headerToken := c.GetHeader(consts.HeaderXCSRFToken)
			if headerToken == "" || headerToken != cookieToken {
				l.Warnw("CSRF Middleware - Invalid or mismatched token",
					"ip", util.GetIPAddress(c),
					"path", c.Request.URL.Path,
					"headerEmpty", headerToken == "",
					"mismatch", headerToken != cookieToken)
				response.ControllerResponse(c, http.StatusForbidden, util.ErrCSRFInvalid.Error(), nil, false)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// HybridMiddleware skips CSRF protection if an Authorization header is present (typically mobile/API clients).
// For browser clients (no Authorization header, using cookies), it enforces CSRF protection.
func HybridMiddleware(l logger.Log, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// JWT client (Mobile/Desktop App) -> skip CSRF
		auth := util.GetAuthorization(c)
		if auth != "" {
			c.Next()
			return
		}

		// Cookie client (Web Browser) -> enforce CSRF
		Middleware(l, cookieName)(c)
	}
}
