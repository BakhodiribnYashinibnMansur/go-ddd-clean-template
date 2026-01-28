package log

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc) {
	g := r.Group("audit")
	g.Use(authMiddleware)
	{
		g.GET("/logs", c.Gets)
		g.GET("/logins", c.Logins)
		g.GET("/sessions", c.Sessions)
		g.GET("/actions", c.Actions)
	}
}
