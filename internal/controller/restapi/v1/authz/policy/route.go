package policy

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers administrative endpoints for policy management.
func Route(api *gin.RouterGroup, c ControllerI) {
	policies := api.Group("/policies")
	{
		policies.POST("", c.Create)                          // Create new policy
		policies.GET("", c.Gets)                             // List policies
		policies.GET("/:"+consts.ParamPolicyID, c.Get)       // Get policy details
		policies.PUT("/:"+consts.ParamPolicyID, c.Update)    // Update policy
		policies.DELETE("/:"+consts.ParamPolicyID, c.Delete) // Delete policy
	}
}
