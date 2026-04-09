package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all ActivityLog HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/activity-logs", h.ListActivityLogs)
}
