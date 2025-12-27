package session

import (
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	session := api.Group("/sessions")
	{
		session.POST("/", c.Create)
		session.GET("/:id", c.Get)
		session.PATCH("/:id/activity", c.UpdateActivity)
		session.DELETE("/:id", c.Delete)
	}
}
