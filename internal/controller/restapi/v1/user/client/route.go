package client

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers all user-facing profiling and account management endpoints.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, authzMiddleware, csrfMiddleware gin.HandlerFunc) {
	users := api.Group("users")
	{
		// PROTECTED Domain: requires a valid access token, fine-grained authz, and CSRF verification.
		users.Use(authMiddleware)
		users.Use(authzMiddleware)
		users.Use(csrfMiddleware)
		{
			// Account Management
			users.POST("", c.Create)                        // Create a user manually (Admin privileged).
			users.GET("", c.Users)                          // Fetch multiple users with filtering (Admin privileged).
			users.GET("/:"+consts.ParamUserID, c.User)      // Retrieve detailed profile for a specific user.
			users.PATCH("/:"+consts.ParamUserID, c.Update)  // Partially update user metadata or attributes.
			users.DELETE("/:"+consts.ParamUserID, c.Delete) // Permanently remove or deactivate a user account.
		}
	}
}
