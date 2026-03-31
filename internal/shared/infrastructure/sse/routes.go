package sse

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all SSE streaming endpoints.
// authMW protects all streams. authzMW restricts audit/monitoring to admins.
func RegisterRoutes(router *gin.Engine, h *Handler, authMW, authzMW gin.HandlerFunc) {
	stream := router.Group("/api/v1/stream")
	stream.Use(authMW)

	// User-specific (any authenticated user)
	stream.GET("/notifications", h.StreamNotifications)
	stream.GET("/jobs/:id", h.StreamJobProgress)

	// Admin-only
	admin := stream.Group("")
	admin.Use(authzMW)
	admin.GET("/audit", h.StreamAudit)
	admin.GET("/monitoring", h.StreamMonitoring)
}
