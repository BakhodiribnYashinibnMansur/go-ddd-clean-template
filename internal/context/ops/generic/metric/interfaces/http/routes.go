package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all Metric HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/metrics")
	g.POST("", h.Create)
	g.GET("", h.List)
}
