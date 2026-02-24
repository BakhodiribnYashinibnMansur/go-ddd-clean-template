package dashboard

import "github.com/gin-gonic/gin"

func Route(api *gin.RouterGroup, ctrl ControllerI, auth, authz gin.HandlerFunc) {
	g := api.Group("/dashboard", auth, authz)
	{
		g.GET("/stats", ctrl.Get)
	}
}
