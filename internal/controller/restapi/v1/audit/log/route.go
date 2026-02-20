package log

import (
	"github.com/gin-gonic/gin"
)

func Route(r *gin.RouterGroup, c ControllerI, authMiddleware, authzMiddleware gin.HandlerFunc) {
	g := r.Group("audit")
	g.Use(authMiddleware)
	g.Use(authzMiddleware)
	{
		g.GET("/logs", c.Gets)
		g.GET("/logins", c.Logins)
		g.GET("/sessions", c.Sessions)
		g.GET("/actions", c.Actions)
	}
}
