package client

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers all user-facing endpoints, categorizing them into public and protected segments.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, refreshMiddleware, csrfMiddleware gin.HandlerFunc) {
	users := api.Group("/users")
	{
		// PUBLIC Domain: endpoints accessible without an active session.
		users.GET("/csrf-token", c.CsrfToken) // Fetches initial anti-CSRF token for forms.
		users.POST("/sign-in", c.SignIn)      // Entry point for credentials verification.
		users.POST("/sign-up", c.SignUp)      // Entry point for new account registration.

		// SEMI-PROTECTED Domain: requires specialized refresh token middleware.
		users.POST("/refresh", csrfMiddleware, refreshMiddleware, c.RefreshToken)

		// PROTECTED Domain: requires a valid access token and CSRF verification.
		protected := users.Group("/")
		protected.Use(authMiddleware)
		protected.Use(csrfMiddleware)
		{
			protected.POST("/sign-out", c.SignOut) // Terminate active session and clear cookies.

			// Account Management
			protected.POST("/", c.Create)                       // Create a user manually (Admin privileged).
			protected.GET("/", c.Users)                         // Fetch multiple users with filtering (Admin privileged).
			protected.GET("/:"+consts.ParamUserID, c.User)      // Retrieve detailed profile for a specific user.
			protected.PATCH("/:"+consts.ParamUserID, c.Update)  // Partially update user metadata or attributes.
			protected.DELETE("/:"+consts.ParamUserID, c.Delete) // Permanently remove or deactivate a user account.
		}
	}
}
