package scope

import (
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	api.GET("/scope", c.Get)
	scopes := api.Group("/scopes")
	{
		scopes.POST("", c.Create)
		scopes.GET("", c.Gets)
		scopes.DELETE("", c.Delete)
	}
}
