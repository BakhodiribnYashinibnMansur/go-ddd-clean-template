package permission

import (
	"gct/internal/shared/domain/consts"

	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	perms := api.Group("/permissions")
	{
		perms.POST("", c.Create)
		perms.GET("", c.Gets)
		perms.GET("/:"+consts.ParamPermID, c.Get)
		perms.PUT("/:"+consts.ParamPermID, c.Update)
		perms.DELETE("/:"+consts.ParamPermID, c.Delete)

		perms.POST("/:"+consts.ParamPermID+"/scopes", c.AssignScope)
		perms.DELETE("/:"+consts.ParamPermID+"/scopes", c.RemoveScope)
	}
}
