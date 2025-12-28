package client

import (
	"github.com/gin-gonic/gin"
)

const (
	userID = "user_id"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	user := api.Group("/users")
	{
		user.POST("/sign-in", c.SignIn)
		user.POST("/sign-up", c.SignUp)
		user.POST("/sign-out", c.SignOut)
		user.POST("/", c.Create)
		user.GET("/", c.Users)
		user.GET("/:"+userID, c.User)
		user.PATCH("/:"+userID, c.Update)
		user.DELETE("/:"+userID, c.Delete)
	}
}
