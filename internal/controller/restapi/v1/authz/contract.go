package authz

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/controller/restapi/v1/authz/scope"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	RoleI       role.ControllerI
	PermissionI permission.ControllerI
	ScopeI      scope.ControllerI
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		RoleI:       role.New(u, cfg, l),
		PermissionI: permission.New(u, cfg, l),
		ScopeI:      scope.New(u, cfg, l),
	}
}
