package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all SystemError HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/system-errors")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("/:id/resolve", h.Resolve)
}
