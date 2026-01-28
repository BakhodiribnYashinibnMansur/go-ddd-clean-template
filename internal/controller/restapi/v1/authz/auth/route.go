package auth

import (
	"github.com/gin-gonic/gin"
)

// Route registers all authentication-facing endpoints, categorizing them into public and protected segments.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, refreshMiddleware, csrfMiddleware gin.HandlerFunc) {
	// CSRF: separate from auth to match swagger and general utility.
	api.GET("/csrf-token", c.CsrfToken)

	// Auth Domain: endpoints accessible without an active session or for session management.
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/sign-in", c.SignIn)
		authGroup.POST("/sign-up", c.SignUp)
		authGroup.POST("/refresh", refreshMiddleware, c.RefreshToken)
	}

	// PROTECTED Domain: requires a valid access token and CSRF verification.
	protected := api.Group("/auth")
	protected.Use(authMiddleware)
	protected.Use(csrfMiddleware)
	{
		protected.POST("/sign-out", c.SignOut) // Terminate active session and clear cookies.
	}
}
