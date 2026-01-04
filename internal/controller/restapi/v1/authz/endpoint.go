package authz

import (
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/controller/restapi/v1/authz/scope"
	"github.com/gin-gonic/gin"
)

func AuthzRoute(api *gin.RouterGroup, controller *Controller, authFn, authzFn, csrfFn gin.HandlerFunc) {
	authz := api.Group("/authz")
	authz.Use(authFn)
	authz.Use(authzFn)
	authz.Use(csrfFn)

	role.Route(authz, controller.RoleI)
	permission.Route(authz, controller.PermissionI)
	scope.Route(authz, controller.ScopeI)
}
