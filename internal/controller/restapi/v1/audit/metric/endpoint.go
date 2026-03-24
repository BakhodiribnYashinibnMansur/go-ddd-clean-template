package metric

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI, authMiddleware, authzMiddleware gin.HandlerFunc) {
	g := r.Group("/metrics", authMiddleware, authzMiddleware)
	{
		g.GET("/functions", c.Gets)
	}
}
