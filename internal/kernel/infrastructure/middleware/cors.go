package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"gct/config"
	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures Cross-Origin Resource Sharing (CORS) policies for the API.
// It sets required headers to allow designated origins to access the resources and
// handles preflight OPTIONS requests by signaling valid methods and headers.
func CORSMiddleware(cfg config.CORS) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Resolve allowed origin
		requestOrigin := ctx.GetHeader("Origin")
		allowOrigin := ""

		for _, o := range cfg.AllowedOrigins {
			// Exact match has priority
			if o == requestOrigin {
				allowOrigin = requestOrigin
				break
			}
			// Wildcard match
			if o == "*" {
				if cfg.AllowCredentials {
					// Browsers require specific origin if credentials are used, even if configuration allows all
					if requestOrigin != "" {
						allowOrigin = requestOrigin
					} else {
						allowOrigin = "*"
					}
				} else {
					allowOrigin = "*"
				}
				break
			}
		}

		// Set standard CORS headers.
		// "Access-Control-Allow-Origin" controls which domains can access resources.
		if allowOrigin != "" {
			ctx.Header(consts.HeaderAccessControlAllowOrigin, allowOrigin)
		}

		// "Access-Control-Allow-Credentials" enables cookies/auth headers in cross-origin requests.
		if cfg.AllowCredentials {
			ctx.Writer.Header().Set(consts.HeaderAccessControlAllowCredentials, "true")
		}

		// Define the whitelist of allowed request headers.
		if len(cfg.AllowedHeaders) > 0 {
			ctx.Writer.Header().Set(consts.HeaderAccessControlAllowHeaders, strings.Join(cfg.AllowedHeaders, ", "))
		}

		// Define the whitelist of exposed request headers.
		if len(cfg.ExposedHeaders) > 0 {
			ctx.Writer.Header().Set(consts.HeaderAccessControlExposeHeaders, strings.Join(cfg.ExposedHeaders, ", "))
		}

		// Define the whitelist of allowed HTTP methods.
		if len(cfg.AllowedMethods) > 0 {
			ctx.Writer.Header().Set(consts.HeaderAccessControlAllowMethods, strings.Join(cfg.AllowedMethods, ", "))
		}

		// Cache the preflight response (OPTIONS) to reduce server load.
		if cfg.MaxAge > 0 {
			ctx.Header(consts.HeaderAccessControlMaxAge, strconv.Itoa(cfg.MaxAge))
		}

		// Handle Preflight Request:
		// If the method is OPTIONS, the browser is querying permission to make the actual request.
		// We respond with 204 No Content and the headers above to grant permission.
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
