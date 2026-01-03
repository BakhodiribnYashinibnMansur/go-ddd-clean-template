package session

import (
	"github.com/gin-gonic/gin"

	"gct/consts"
)

func Route(api *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc) {
	session := api.Group("/sessions")
	session.Use(authMiddleware)
	{
		session.POST("/", c.Create)
		session.GET("/", c.Sessions)
		session.GET("/:"+consts.ParamID, c.Session)
		session.PATCH("/:"+consts.ParamID+"/activity", c.UpdateActivity)
		session.DELETE("/:"+consts.ParamID, c.Delete)
		session.POST("/revoke-all", c.RevokeAll)
		session.DELETE("/device/:device_id", c.RevokeByDevice)
	}
}
