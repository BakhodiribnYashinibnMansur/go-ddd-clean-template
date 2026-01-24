package middleware

import (
	"net/http"

	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"

	"github.com/gin-gonic/gin"
)

// FetchMetadata middleware implements "Fetch Metadata Request Headers" protection.
// It is a defense-in-depth resource isolation policy that protects against Cross-Site Request Forgery (CSRF),
// Cross-Site Script Inclusion (XSSI), and timing attacks by validating the origin of the request.
//
// See: https://web.dev/fetch-metadata/
func FetchMetadata(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Optimization: Only enforce strict metadata checks in production environments where security is paramount.
		if !cfg.App.IsProd() || !cfg.Security.FetchMetadata {
			c.Next()
			return
		}

		// 1. Check for the existence of the Sec-Fetch-Site header.
		// Modern browsers send this header. If missing, it implies a non-browser client (Postman, cURL)
		// or a very old browser. In a strict API production environment, we may choose to block these
		// to enforce browser-only access policies, or allow them if API keys are used (handled by other middlewares).
		// Here, we block requests missing the header to prevent ambiguity.
		site := c.GetHeader(consts.HeaderSecFetchSite)
		if site == "" {
			response.ControllerResponse(c, http.StatusForbidden, util.ErrFetchMetadataSuspicious, nil, false)
			c.Abort()
			return
		}

		// 2. Allow Same-Origin and Same-Site requests.
		// "same-origin": Request comes from the same application (e.g. AJAX call).
		// "same-site": Request comes from a related subdomain (e.g. auth.example.com -> api.example.com).
		if site == consts.HeaderValueSameOrigin || site == consts.HeaderValueSameSite {
			c.Next()
			return
		}

		// 3. Allow Top-Level Navigation.
		// We must allow users to arrive at our application from external sites (e.g. clicking a link on Google).
		// A request is a valid top-level navigation if:
		// - It is a GET request (safe method).
		// - Sec-Fetch-Mode is 'navigate'.
		// - Sec-Fetch-Dest is 'document' (loading a full page, not an image/script).
		mode := c.GetHeader(consts.HeaderSecFetchMode)
		dest := c.GetHeader(consts.HeaderSecFetchDest)
		if mode == consts.HeaderValueNavigate && dest == consts.HeaderValueDocument && c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		// 4. Default Deny.
		// Block all other cross-site interactions. This effectively kills:
		// - <img src="api.example.com/delete_account"> (CSRF via GET)
		// - <script src="api.example.com/sensitive_data.json"> (XSSI)
		// - POST submissions from malicious forms.
		response.ControllerResponse(c, http.StatusForbidden, util.ErrFetchMetadataBlocked, nil, false)
		c.Abort()
	}
}
