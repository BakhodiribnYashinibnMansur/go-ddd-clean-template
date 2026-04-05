package middleware

import (
	"strings"

	"gct/internal/platform/domain/consts"

	"github.com/gin-gonic/gin"
)

// Security returns a Gin middleware that implements essential HTTP security headers (Helmet-style).
// These headers protect the application against common vulnerabilities like XSS, Clickjacking, and MIME-sniffing.
//
// Recommended reading: https://owasp.org/www-project-secure-headers/
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: "nosniff"
		// Prevents the browser from misinterpreting files as a different MIME type (e.g., treating a text file as JavaScript).
		c.Header(consts.HeaderXContentTypeOptions, consts.HeaderValueNoSniff)

		// X-Frame-Options: "DENY"
		// Prevents the page from being displayed in a frame/iframe/object to mitigate Clickjacking attacks.
		c.Header(consts.HeaderXFrameOptions, consts.HeaderValueDeny)

		// X-XSS-Protection: "1; mode=block"
		// Instructs legacy browsers to block the response if a reflected XSS attack is detected.
		c.Header(consts.HeaderXXSSProtection, consts.HeaderValueXSSBlock)

		// Referrer-Policy: "strict-origin"
		// Controls how much referrer information is sent when navigating to other sites.
		// "strict-origin" sends the origin only when the protocol security level mimics (HTTPS->HTTPS).
		c.Header(consts.HeaderReferrerPolicy, consts.HeaderValueStrictOrigin)

		// X-Permitted-Cross-Domain-Policies: "none"
		// Restricts policies for Adobe Flash or PDF documents to prevent cross-domain data leakage.
		c.Header(consts.HeaderXPermittedCP, consts.HeaderValueNone)

		// Content-Security-Policy (CSP):
		// Acts as a whitelist for sources of executable scripts, styles, and other resources.
		// This configuration is relatively permissive for a template but should be tightened in production.
		//
		// - default-src 'self': Only allow resources from same origin by default.
		// - script-src: Allow scripts from self and trusted CDNs.
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com https://unpkg.com https://cdn.jsdelivr.net",
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com https://unpkg.com https://cdn.jsdelivr.net https://boxicons.com",
			"font-src 'self' https://fonts.gstatic.com https://fonts.googleapis.com https://unpkg.com https://boxicons.com data:",
			"img-src 'self' data: https:",
			"connect-src 'self'",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}
		c.Header(consts.HeaderCSP, strings.Join(csp, "; "))

		// Strict-Transport-Security (HSTS):
		// Tells the browser to remember that this site should only be accessed using HTTPS.
		// Triggered only if the current request is already secure (or behind a secure proxy).
		if c.Request.TLS != nil || c.GetHeader(consts.HeaderXForwardedProto) == "https" {
			c.Header(consts.HeaderHSTS, consts.HeaderValueHSTS)
		}

		c.Next()
	}
}

// SecurityCustom provides a flexible version of the security middleware allowing for customized CSP policies.
// This is useful for specific endpoints that might need to render external widgets or frames.
func SecurityCustom(cspDirectives []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header(consts.HeaderXContentTypeOptions, consts.HeaderValueNoSniff)
		c.Header(consts.HeaderXFrameOptions, consts.HeaderValueDeny)
		c.Header(consts.HeaderXXSSProtection, consts.HeaderValueXSSBlock)
		c.Header(consts.HeaderReferrerPolicy, consts.HeaderValueStrictOrigin)
		c.Header(consts.HeaderXPermittedCP, consts.HeaderValueNone)

		if len(cspDirectives) > 0 {
			c.Header(consts.HeaderCSP, strings.Join(cspDirectives, "; "))
		}

		if c.Request.TLS != nil || c.GetHeader(consts.HeaderXForwardedProto) == "https" {
			c.Header(consts.HeaderHSTS, consts.HeaderValueHSTS)
		}

		c.Next()
	}
}
