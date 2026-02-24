package webhook

import "github.com/gin-gonic/gin"

func Route(api *gin.RouterGroup, ctrl ControllerI, auth, authz, csrf gin.HandlerFunc) {
	g := api.Group("/webhooks", auth, authz)
	{
		g.POST("", csrf, ctrl.Create)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.Get)
		g.PUT("/:id", csrf, ctrl.Update)
		g.DELETE("/:id", csrf, ctrl.Delete)
		g.POST("/:id/test", csrf, ctrl.Test)
	}
}
