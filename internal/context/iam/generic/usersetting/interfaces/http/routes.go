package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all UserSetting HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/user-settings")
	g.POST("", h.Upsert)
	g.GET("", h.List)
	g.DELETE("/:id", h.Delete)
}
