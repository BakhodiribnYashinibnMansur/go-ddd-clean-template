package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all Audit HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/audit-logs", h.ListAuditLogs)
	rg.GET("/endpoint-history", h.ListEndpointHistory)
}
