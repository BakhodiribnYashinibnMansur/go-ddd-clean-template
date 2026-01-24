// Package authz provides the API layer for managing Authorization components:
// Roles, Permissions, Scopes, and Policies.
package authz

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/controller/restapi/v1/authz/scope"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

// Controller integrates sub-controllers for the entire Authorization subsystem.
// It manages the lifecycle and access endpoints for RBAC and ABAC entities.
type Controller struct {
	RoleI       role.ControllerI       // Controller for Role-based management.
	PermissionI permission.ControllerI // Controller for granular Permission management.
	ScopeI      scope.ControllerI      // Controller for Scope-based (resource/action) access.
}

// New instantiates a composite Authorization controller with all its specialized sub-handlers.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		RoleI:       role.New(u, cfg, l),
		PermissionI: permission.New(u, cfg, l),
		ScopeI:      scope.New(u, cfg, l),
	}
}
