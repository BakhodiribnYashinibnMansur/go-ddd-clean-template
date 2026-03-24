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
	"gct/internal/shared/infrastructure/logger"
)

type UseCaseI interface {
	Access() access.UseCaseI
	Role() role.UseCaseI
	Permission() permission.UseCaseI
	Policy() policy.UseCaseI
	Relation() relation.UseCaseI
	Scope() scope.UseCaseI
}

type UseCase struct {
	access     access.UseCaseI
	role       role.UseCaseI
	permission permission.UseCaseI
	policy     policy.UseCaseI
	relation   relation.UseCaseI
	scope      scope.UseCaseI
}

func New(r *repo.Repo, logger logger.Log, cfg *config.Config) UseCaseI {
	return &UseCase{
		access:     access.New(r.Persistent, logger),
		role:       role.New(r.Persistent, logger),
		permission: permission.New(r.Persistent, logger),
		policy:     policy.New(r.Persistent, logger),
		relation:   relation.New(r.Persistent, logger),
		scope:      scope.New(r.Persistent, logger),
	}
}

func (uc *UseCase) Access() access.UseCaseI         { return uc.access }
func (uc *UseCase) Role() role.UseCaseI             { return uc.role }
func (uc *UseCase) Permission() permission.UseCaseI { return uc.permission }
func (uc *UseCase) Policy() policy.UseCaseI         { return uc.policy }
func (uc *UseCase) Relation() relation.UseCaseI     { return uc.relation }
func (uc *UseCase) Scope() scope.UseCaseI           { return uc.scope }
