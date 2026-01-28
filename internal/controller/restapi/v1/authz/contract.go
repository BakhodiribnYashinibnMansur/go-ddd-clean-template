// Package authz provides the API layer for managing Authorization components:
// Roles, Permissions, Scopes, and Policies.
package authz

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/authz/auth"
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/controller/restapi/v1/authz/policy"
	"gct/internal/controller/restapi/v1/authz/relation"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/controller/restapi/v1/authz/scope"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

// Controller integrates sub-controllers for the entire Authorization subsystem.
// It manages the lifecycle and access endpoints for RBAC and ABAC entities.
type Controller struct {
	AuthI       auth.ControllerI
	RoleI       role.ControllerI       // Controller for Role-based management.
	PermissionI permission.ControllerI // Controller for granular Permission management.
	ScopeI      scope.ControllerI      // Controller for Scope-based (resource/action) access.
	PolicyI     policy.ControllerI     // Controller for Policy-based management.
	RelationI   relation.ControllerI   // Controller for Relation management.
}

// New instantiates a composite Authorization controller with all its specialized sub-handlers.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		AuthI:       auth.New(u, cfg, l),
		RoleI:       role.New(u, cfg, l),
		PermissionI: permission.New(u, cfg, l),
		ScopeI:      scope.New(u, cfg, l),
		PolicyI:     policy.New(u, cfg, l),
		RelationI:   relation.New(u, cfg, l),
	}
}
