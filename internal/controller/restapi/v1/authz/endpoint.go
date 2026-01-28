package authz

import (
	"gct/internal/controller/restapi/v1/authz/auth"
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/controller/restapi/v1/authz/policy"
	"gct/internal/controller/restapi/v1/authz/relation"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/controller/restapi/v1/authz/scope"

	"github.com/gin-gonic/gin"
)

// AuthzRoute defines a protected route group for Authorization management.
// It applies Authentication, Authorization (RBAC/ABAC), and CSRF protection middlewares
// to all routes within the Role, Permission, and Scope domains.
func AuthzRoute(api *gin.RouterGroup, controller *Controller, authFn, refreshFn, authzFn, csrfFn gin.HandlerFunc) {
	// Register authentication-related routes under /auth.
	auth.Route(api, controller.AuthI, authFn, refreshFn, csrfFn)

	authz := api.Group("/authz")

	// Apply comprehensive security middleware stack.
	authz.Use(authFn)
	authz.Use(authzFn)
	authz.Use(csrfFn)

	// Delegate route registration to sub-packages.
	role.Route(authz, controller.RoleI)
	permission.Route(authz, controller.PermissionI)
	scope.Route(authz, controller.ScopeI)
	policy.Route(authz, controller.PolicyI)
	relation.Route(authz, controller.RelationI)
}
