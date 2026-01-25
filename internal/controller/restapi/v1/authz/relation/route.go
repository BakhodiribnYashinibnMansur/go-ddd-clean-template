package relation

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers administrative endpoints for relation management.
func Route(api *gin.RouterGroup, c ControllerI) {
	relations := api.Group("/relations")
	{
		relations.POST("", c.Create)                            // Create new relation
		relations.GET("", c.Gets)                               // List relations
		relations.GET("/:"+consts.ParamRelationID, c.Get)       // Get relation details
		relations.PUT("/:"+consts.ParamRelationID, c.Update)    // Update relation
		relations.DELETE("/:"+consts.ParamRelationID, c.Delete) // Delete relation

		// User linkage
		relations.POST("/:"+consts.ParamRelationID+"/users/:"+consts.ParamUserID, c.AddUser)
		relations.DELETE("/:"+consts.ParamRelationID+"/users/:"+consts.ParamUserID, c.RemoveUser)
	}
}
