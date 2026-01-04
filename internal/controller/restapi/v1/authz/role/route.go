package role

import (
	"gct/consts"
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c ControllerI) {
	roles := api.Group("/roles")
	{
		roles.POST("", c.Create)
		roles.GET("", c.Gets)
		roles.GET("/:"+consts.ParamRoleID, c.Get)
		roles.PUT("/:"+consts.ParamRoleID, c.Update)
		roles.DELETE("/:"+consts.ParamRoleID, c.Delete)

		// Permissions
		roles.POST("/:"+consts.ParamRoleID+"/permissions/:"+consts.ParamPermID, c.AddPermission)
		roles.DELETE("/:"+consts.ParamRoleID+"/permissions/:"+consts.ParamPermID, c.RemovePermission)
	}

	// This part was for USER assignment.
	// Usually POST /users/:id/roles/:role_id ...
	// If I put it here, the path is /authz/roles/users... which is weird.
	// Maybe /authz/users/:id/roles/:role_id is better.
	// But then I need access to /authz/users group.
	// If AuthzRoute creates /authz, I can pass it here.
	// But 'api' passed to Route IS the /authz group (if I implement it that way).
	// So api.Group("/users") works.

	users := api.Group("/users")
	{
		users.POST("/:"+consts.ParamUserID+"/roles/:"+consts.ParamRoleID, c.Assign)
	}
}
