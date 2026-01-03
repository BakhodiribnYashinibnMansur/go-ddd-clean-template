package authz

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/authz/access"
	"gct/internal/usecase/authz/permission"
	"gct/internal/usecase/authz/policy"
	"gct/internal/usecase/authz/relation"
	"gct/internal/usecase/authz/role"
	"gct/internal/usecase/authz/scope"
	"gct/pkg/logger"
)

type UseCase struct {
	Access     access.UseCaseI
	Role       role.UseCaseI
	Permission permission.UseCaseI
	Policy     policy.UseCaseI
	Relation   relation.UseCaseI
	Scope      scope.UseCaseI
}

func New(r *repo.Repo, logger logger.Log, cfg *config.Config) *UseCase {
	return &UseCase{
		Access:     access.New(r.Persistent, logger),
		Role:       role.New(r.Persistent, logger),
		Permission: permission.New(r.Persistent, logger),
		Policy:     policy.New(r.Persistent, logger),
		Relation:   relation.New(r.Persistent, logger),
		Scope:      scope.New(r.Persistent, logger),
	}
}
