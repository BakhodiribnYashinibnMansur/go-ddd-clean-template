package client

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers all user-facing profiling and account management endpoints.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, csrfMiddleware gin.HandlerFunc) {
	users := api.Group("/users")
	{
		// PROTECTED Domain: requires a valid access token and CSRF verification.
		protected := users.Group("/")
		protected.Use(authMiddleware)
		protected.Use(csrfMiddleware)
		{
			// Account Management
			protected.POST("/", c.Create)                       // Create a user manually (Admin privileged).
			protected.GET("/", c.Users)                         // Fetch multiple users with filtering (Admin privileged).
			protected.GET("/:"+consts.ParamUserID, c.User)      // Retrieve detailed profile for a specific user.
			protected.PATCH("/:"+consts.ParamUserID, c.Update)  // Partially update user metadata or attributes.
			protected.DELETE("/:"+consts.ParamUserID, c.Delete) // Permanently remove or deactivate a user account.
		}
	}
}
