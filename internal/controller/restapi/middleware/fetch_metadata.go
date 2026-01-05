package middleware

import (
	"net/http"

	"gct/config"

	"github.com/gin-gonic/gin"
)

// FetchMetadata middleware implements Fetch Metadata Request Headers protection.
// It blocks cross-site requests and non-browser requests (like Postman) in production.
// This is mandated by security policy to prevent CSRF and cross-site leaks.
func FetchMetadata(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only enforce in production and if enabled in config
		if !cfg.App.IsProd() || !cfg.Security.FetchMetadata {
			c.Next()
			return
		}

		// 1. Missing Sec-Fetch-Site header indicates a non-browser request (e.g. Postman, curl, or old browser)
		// Requirement: block Postman/cURL in production.
		site := c.GetHeader("Sec-Fetch-Site")
		if site == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Request blocked: suspicious source (Postman/cURL blocked in production)",
				"code":  "SUSPICIOUS_SOURCE",
			})
			return
		}

		// 2. Allow same-origin and same-site requests
		if site == "same-origin" || site == "same-site" {
			c.Next()
			return
		}

		// 3. Allow top-level navigation (e.g. user clicking a link to our site)
		// Mode must be 'navigate', Dest must be 'document', and Method must be 'GET'
		mode := c.GetHeader("Sec-Fetch-Mode")
		dest := c.GetHeader("Sec-Fetch-Dest")
		if mode == "navigate" && dest == "document" && c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		// 4. Block everything else (cross-site sub-resources like images, scripts, or cross-site API calls)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Request blocked by Fetch Metadata policy (cross-site requests blocked for security)",
			"code":  "CROSS_SITE_BLOCK",
		})
	}
}
