package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all IPRule HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/ip-rules")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
