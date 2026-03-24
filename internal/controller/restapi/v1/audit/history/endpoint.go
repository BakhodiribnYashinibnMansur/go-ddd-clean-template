package history

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI, authMiddleware, authzMiddleware gin.HandlerFunc) {
	g := r.Group("audit")
	g.Use(authMiddleware)
	g.Use(authzMiddleware)
	{
		g.GET("/history", c.Gets)
	}
}
