package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/pkg/csrf"
	"github.com/gin-gonic/gin"
)

// CsrfToken godoc
// @Summary     Get CSRF token
// @Description Generates and sets a CSRF token with HMAC protection. Token is stored server-side and sent to client in both cookie and response body.
// @Tags        auth
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /csrf-token [get]
func (c *Controller) CsrfToken(ctx *gin.Context) {
	// Get or create session ID for CSRF token binding
	// For guest users, we create a temporary session
	sessionID := util.GetDeviceID(ctx)
	if sessionID == "" {
		sessionID = util.GenerateToken()
	}

	// Generate cryptographically secure CSRF token with HMAC
	csrfGen := c.csrfGenerator
	if csrfGen == nil {
		c.l.Errorw("CSRF generator not initialized")
		response.ControllerResponse(ctx, http.StatusInternalServerError, "CSRF service unavailable", nil, false)
		return
	}

	token, err := csrfGen.GenerateToken(sessionID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - auth - csrf - generate")
		response.ControllerResponse(ctx, http.StatusInternalServerError, "Failed to generate CSRF token", nil, false)
		return
	}

	// Store token hash server-side
	if c.csrfStore != nil {
		if err := c.csrfStore.Set(ctx.Request.Context(), sessionID, token.Hash, csrf.DefaultExpiration); err != nil {
			util.LogError(c.l, err, "http - v1 - auth - csrf - store")
			response.ControllerResponse(ctx, http.StatusInternalServerError, "Failed to store CSRF token", nil, false)
			return
		}
	}

	// Set cookie with CSRF token (plain value for double-submit pattern)
	// HttpOnly=false because client JS needs to read it to send in header
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     consts.COOKIE_CSRF_TOKEN,
		Value:    token.Value,
		Path:     consts.CookiePath,
		HttpOnly: false,                   // Must be readable by JS for header submission
		Secure:   c.cfg.Cookie.IsSecure(), // HTTPS only in production
		SameSite: http.SameSiteLaxMode,    // Lax provides good CSRF protection
		MaxAge:   int(csrf.DefaultExpiration.Seconds()),
	})

	c.l.Infow("CSRF token generated",
		"session_id", sessionID,
		"expires_at", token.ExpiresAt,
		"ip", util.GetIPAddress(ctx))

	response.ControllerResponse(ctx, http.StatusOK, nil, gin.H{
		"csrf_token": token.Value,
		"expires_at": token.ExpiresAt.Unix(),
	}, true)
}
