package admin

import (
	"github.com/gin-gonic/gin"
)

// Register adds admin-specific routes to the provided router group.
func (c *Controller) Register(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	g := r.Group("admin")
	g.Use(authMiddleware)
	{
		g.POST("/linter/run", c.RunLinter)
	}
}
