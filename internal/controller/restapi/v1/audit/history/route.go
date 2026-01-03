package history

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI) {
	g := r.Group("/audit")
	{
		g.GET("/history", c.Gets)
	}
}
