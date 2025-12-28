package session

import (
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	session := api.Group("/sessions")
	{
		session.POST("/", c.Create)
		session.GET("/", c.Sessions)
		session.GET("/:id", c.Session)
		session.PATCH("/:id/activity", c.UpdateActivity)
		session.DELETE("/:id", c.Delete)
	}
}
