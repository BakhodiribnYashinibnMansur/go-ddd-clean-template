package role

import (
	"gct/consts"

	"github.com/gin-gonic/gin"
)

// Route registers administrative endpoints for role management and relationship mapping.
func Route(api *gin.RouterGroup, c ControllerI) {
	// ROLE Management Group
	roles := api.Group("/roles")
	{
		// CRUD operations for Role entities.
		roles.POST("", c.Create)                        // Register a new role.
		roles.GET("", c.Gets)                           // List roles with filtering/pagination.
		roles.GET("/:"+consts.ParamRoleID, c.Get)       // Fetch details for a specific role.
		roles.PUT("/:"+consts.ParamRoleID, c.Update)    // Update an existing role's metadata.
		roles.DELETE("/:"+consts.ParamRoleID, c.Delete) // Remove a role from the system.

		// PERMISSION Linkage: Manage many-to-many relationships between Roles and Permissions.
		// These endpoints grant or revoke specific abilities to all users holding a certain role.
		roles.POST("/:"+consts.ParamRoleID+"/permissions/:"+consts.ParamPermID, c.AddPermission)
		roles.DELETE("/:"+consts.ParamRoleID+"/permissions/:"+consts.ParamPermID, c.RemovePermission)
	}

	// USER-ROLE Relationship Group
	// Defined within the authz scope to facilitate linking users to security roles.
	users := api.Group("/users")
	{
		users.POST("/:"+consts.ParamUserID+"/roles/:"+consts.ParamRoleID, c.Assign)
	}
}
