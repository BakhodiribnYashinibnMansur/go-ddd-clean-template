package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all Session HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	sessions := rg.Group("/sessions")
	{
		sessions.GET("", h.List)
		sessions.GET("/:id", h.Get)
		sessions.DELETE("/:id", h.Delete)
		sessions.POST("/revoke-all", h.RevokeAll)
	}
}
