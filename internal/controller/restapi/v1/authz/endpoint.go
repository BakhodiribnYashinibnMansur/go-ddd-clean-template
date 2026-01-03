package authz

import (
	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/controller/restapi/v1/authz/scope"
)

func AuthzRoute(api *gin.RouterGroup, controller *Controller, authFn, authzFn gin.HandlerFunc) {
	authz := api.Group("/authz")
	authz.Use(authFn)
	authz.Use(authzFn)

	role.Route(authz, controller.RoleI)
	permission.Route(authz, controller.PermissionI)
	scope.Route(authz, controller.ScopeI)
}
