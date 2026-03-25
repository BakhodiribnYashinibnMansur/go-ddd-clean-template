package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all Authz HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	{
		roles.POST("", h.CreateRole)
		roles.GET("", h.ListRoles)
		roles.GET("/:id", h.GetRole)
		roles.PATCH("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
		roles.POST("/:id/permissions", h.AssignPermission)
	}

	permissions := rg.Group("/permissions")
	{
		permissions.POST("", h.CreatePermission)
		permissions.GET("", h.ListPermissions)
		permissions.DELETE("/:id", h.DeletePermission)
		permissions.POST("/:id/scopes", h.AssignScope)
	}

	policies := rg.Group("/policies")
	{
		policies.POST("", h.CreatePolicy)
		policies.GET("", h.ListPolicies)
		policies.PATCH("/:id", h.UpdatePolicy)
		policies.DELETE("/:id", h.DeletePolicy)
		policies.POST("/:id/toggle", h.TogglePolicy)
	}

	scopes := rg.Group("/scopes")
	{
		scopes.POST("", h.CreateScope)
		scopes.GET("", h.ListScopes)
		scopes.DELETE("", h.DeleteScope)
	}
}
