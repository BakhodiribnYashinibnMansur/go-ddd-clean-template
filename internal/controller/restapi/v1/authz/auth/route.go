package auth

import (
	"github.com/gin-gonic/gin"
)

// Route registers all authentication-facing endpoints, categorizing them into public and protected segments.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, refreshMiddleware, csrfMiddleware gin.HandlerFunc) {
	// Auth Domain: endpoints accessible without an active session or for session management.
	api.GET("/csrf-token", c.CsrfToken) // Fetches initial anti-CSRF token for forms.
	api.POST("/sign-in", c.SignIn)      // Entry point for credentials verification.
	api.POST("/sign-up", c.SignUp)      // Entry point for new account registration.

	// SEMI-PROTECTED Domain: requires specialized refresh token middleware.
	api.POST("/refresh", csrfMiddleware, refreshMiddleware, c.RefreshToken)

	// PROTECTED Domain: requires a valid access token and CSRF verification.
	protected := api.Group("/")
	protected.Use(authMiddleware)
	protected.Use(csrfMiddleware)
	{
		protected.POST("/sign-out", c.SignOut) // Terminate active session and clear cookies.
	}
}
