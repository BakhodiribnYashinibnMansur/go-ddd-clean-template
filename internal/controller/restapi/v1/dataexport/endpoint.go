package dataexport

import "github.com/gin-gonic/gin"

func Route(api *gin.RouterGroup, ctrl ControllerI, auth, authz, csrf gin.HandlerFunc) {
	g := api.Group("/data-export", auth, authz)
	{
		g.POST("", csrf, ctrl.Create)
		g.GET("", ctrl.List)
		g.DELETE("/:id", csrf, ctrl.Delete)
	}
}
