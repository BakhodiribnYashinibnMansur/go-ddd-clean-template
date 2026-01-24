package session

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers administrative and user-facing endpoints for session auditing and revocation.
// All session-related endpoints require a valid access token and CSRF protection.
func Route(api *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc, csrfMiddleware gin.HandlerFunc) {
	session := api.Group("/sessions")

	// Apply global security filters for session management.
	session.Use(authMiddleware)
	session.Use(csrfMiddleware)
	{
		session.POST("/", c.Create)                                      // Manually issue a new session token.
		session.GET("/", c.Sessions)                                     // List all active sessions for the authenticated user.
		session.GET("/:"+consts.ParamID, c.Session)                      // Retrieve detailed metadata for a specific session ID.
		session.PATCH("/:"+consts.ParamID+"/activity", c.UpdateActivity) // Mark a session as active (refresh last_seen).
		session.DELETE("/:"+consts.ParamID, c.Delete)                    // Explicitly invalidate a single session.
		session.POST("/revoke-all", c.RevokeAll)                         // Revoke all sessions except the current one.
		session.DELETE("/device/:device_id", c.RevokeByDevice)           // Targeted logout for a specific device.
	}
}
