package sitesetting

import (
	"github.com/gin-gonic/gin"
)

func SiteSettingRoute(api *gin.RouterGroup, ctrl ControllerI, authMiddleware, authzMiddleware gin.HandlerFunc) {
	g := api.Group("/site-settings", authMiddleware, authzMiddleware)
	{
		g.GET("", ctrl.Gets)
		g.GET("/:key", ctrl.GetByKey)
		g.PUT("/:key", ctrl.UpdateByKey)
	}
}
