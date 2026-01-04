package client

import (
	"gct/consts"
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI, authMiddleware, refreshMiddleware, csrfMiddleware gin.HandlerFunc) {
	users := api.Group("/users")
	{
		users.GET("/csrf-token", c.CsrfToken)
		users.POST("/sign-in", c.SignIn)
		users.POST("/sign-up", c.SignUp)
		users.POST("/refresh", csrfMiddleware, refreshMiddleware, c.RefreshToken)

		// Protected routes
		protected := users.Group("/")
		protected.Use(authMiddleware)
		protected.Use(csrfMiddleware)
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
