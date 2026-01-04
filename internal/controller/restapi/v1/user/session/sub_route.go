package session

import (
	"gct/consts"
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc, csrfMiddleware gin.HandlerFunc) {
	session := api.Group("/sessions")
	session.Use(authMiddleware)
	session.Use(csrfMiddleware)
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
