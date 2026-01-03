package client

import (
	"github.com/gin-gonic/gin"

	"gct/consts"
)

func Route(api *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc) {
	users := api.Group("/users")
	{
		users.POST("/sign-in", c.SignIn)
		users.POST("/sign-up", c.SignUp)

		// Protected routes
		protected := users.Group("/")
		protected.Use(authMiddleware)
		{
			protected.POST("/sign-out", c.SignOut)
			protected.POST("/", c.Create)
			protected.GET("/", c.Users)
			protected.GET("/:"+consts.ParamUserID, c.User)
			protected.PATCH("/:"+consts.ParamUserID, c.Update)
			protected.DELETE("/:"+consts.ParamUserID, c.Delete)
		}
	}
}
