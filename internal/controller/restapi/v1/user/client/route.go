package client

import (
	"github.com/gin-gonic/gin"

	"gct/consts"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	user := api.Group("/users")
	{
		user.POST("/sign-in", c.SignIn)
		user.POST("/sign-up", c.SignUp)
		user.POST("/sign-out", c.SignOut)
		user.POST("/", c.Create)
		user.GET("/", c.Users)
		user.GET("/:"+consts.ParamUserID, c.User)
		user.PATCH("/:"+consts.ParamUserID, c.Update)
		user.DELETE("/:"+consts.ParamUserID, c.Delete)
	}
}
