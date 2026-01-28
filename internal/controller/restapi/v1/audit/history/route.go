package history

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc) {
	g := r.Group("audit")
	g.Use(authMiddleware)
	{
		g.GET("/history", c.Gets)
	}
}
