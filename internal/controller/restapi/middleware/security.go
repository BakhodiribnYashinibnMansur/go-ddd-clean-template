package middleware

import (
	"strings"

	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Security middleware adds security headers to the response.
// This is a Go implementation of the Helmet concept.
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: prevents the browser from interpreting files as something else than declared by the content type.
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: prevents clickjacking by not allowing the page to be rendered in a <frame>, <iframe>, <embed> or <object>.
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: enables the Cross-site scripting (XSS) filter built into most recent web browsers.
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: controls how much referrer information (sent via the Referer header) should be included with requests.
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// X-Permitted-Cross-Domain-Policies: restricts the ability of Adobe Flash and Adobe Acrobat to make requests across domains.
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Content-Security-Policy: helps detect and mitigate certain types of attacks, including Cross Site Scripting (XSS) and data injection attacks.
		// Note: 'unsafe-inline' is used for styles because the current root template and admin panel use it.
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com",
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com",
			"font-src 'self' https://fonts.gstatic.com https://fonts.googleapis.com data:",
			"img-src 'self' data: https:",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}
		c.Header("Content-Security-Policy", strings.Join(csp, "; "))

		// Strict-Transport-Security (HSTS): informs browsers that the site should only be accessed using HTTPS.
		// Only added if the request is HTTPS.
		if c.Request.TLS != nil || c.GetHeader(consts.HeaderXForwardedProto) == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}

// SecurityCustom allows providing custom CSP directives.
func SecurityCustom(cspDirectives []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		if len(cspDirectives) > 0 {
			c.Header("Content-Security-Policy", strings.Join(cspDirectives, "; "))
		}

		if c.Request.TLS != nil || c.GetHeader(consts.HeaderXForwardedProto) == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}
