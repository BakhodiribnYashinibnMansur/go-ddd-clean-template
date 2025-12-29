package session

import (
	"github.com/gin-gonic/gin"

	"gct/consts"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	session := api.Group("/sessions")
	{
		session.POST("/", c.Create)
		session.GET("/", c.Sessions)
		session.GET("/:"+consts.ParamID, c.Session)
		session.PATCH("/:"+consts.ParamID+"/activity", c.UpdateActivity)
		session.DELETE("/:"+consts.ParamID, c.Delete)
	}
}
