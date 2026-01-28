package state

import "github.com/gin-gonic/gin"

// Route registers state management endpoints.
func Route(h *gin.RouterGroup, c ControllerI) {
	t := h.Group("/test")
	{
		t.POST("/reset", c.Reset)
		t.POST("/seed", c.Seed)
	}
}
