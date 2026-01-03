package metric

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc) {
	g := r.Group("/metrics", authMiddleware)
	{
		g.GET("/functions", c.Gets)
	}
}
