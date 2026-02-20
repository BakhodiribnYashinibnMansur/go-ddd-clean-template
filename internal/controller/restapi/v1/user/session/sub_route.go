package session

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers administrative and user-facing endpoints for session auditing and revocation.
// It follows a dual pattern:
// 1. Resource Ownership: /users/{user_id}/sessions for collection management.
// 2. Direct Access: /sessions/{session_id} for individual session operations.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, authzMiddleware, csrfMiddleware gin.HandlerFunc) {
	// Pattern 1: User-Owned Sessions (Collection management)
	userSessions := api.Group("/users/:" + consts.ParamUserID + "/sessions")
	userSessions.Use(authMiddleware)
	userSessions.Use(authzMiddleware)
	userSessions.Use(csrfMiddleware)
	{
		userSessions.GET("", c.Sessions)     // List all active sessions for the specified user.
		userSessions.POST("", c.Create)      // Manually issue a new session token for the specified user.
		userSessions.DELETE("", c.RevokeAll) // Force logout from all devices for this user.
	}

	// Pattern 2: Global/Direct Session Management
	sessions := api.Group("/sessions")
	sessions.Use(authMiddleware)
	sessions.Use(authzMiddleware)
	sessions.Use(csrfMiddleware)
	{
		// Current user's sessions
		sessions.GET("", c.Sessions)
		sessions.POST("", c.Create)
		sessions.POST("/revoke-all", c.RevokeAll)

		// Current session management
		sessions.DELETE("/current", c.RevokeCurrent) // Revoke current session.

		// Individual resource operations
		sessions.GET("/:"+consts.ParamID, c.Session)                    // Retrieve detailed metadata for a specific session.
		sessions.DELETE("/:"+consts.ParamID, c.Delete)                  // Explicitly invalidate a single session by ID.
		sessions.PUT("/:"+consts.ParamID+"/activity", c.UpdateActivity) // Mark a session as active (refresh last_seen).
		sessions.DELETE("/device/:device_id", c.RevokeByDevice)         // Targeted logout for a specific device.
	}
}
