package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all Dashboard HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard/stats", h.GetStats)
}
