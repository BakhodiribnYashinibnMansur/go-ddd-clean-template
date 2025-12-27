package client

import (
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	user := api.Group("/users")
	{
		user.POST("/sign-in", c.SignIn)
		user.POST("/sign-up", c.SignUp)
		user.POST("/sign-out", c.SignOut)
		user.POST("/", c.Create)
		user.GET("/:id", c.Get)
		user.PATCH("/:id", c.Update)
		user.DELETE("/:id", c.Delete)
	}
}
