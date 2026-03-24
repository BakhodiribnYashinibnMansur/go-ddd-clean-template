package authz

import (
	"gct/internal/repo/persistent/postgres/authz/permission"
	"gct/internal/repo/persistent/postgres/authz/policy"
	"gct/internal/repo/persistent/postgres/authz/relation"
	"gct/internal/repo/persistent/postgres/authz/role"
	"gct/internal/repo/persistent/postgres/authz/scope"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"
)

type Authz struct {
	Role       role.RepoI
	Permission permission.RepoI
	Policy     policy.RepoI
	Relation   relation.RepoI
	Scope      scope.RepoI
}

func New(psql *postgres.Postgres, logger logger.Log) *Authz {
	return &Authz{
		Role:       role.New(psql, logger),
		Permission: permission.New(psql, logger),
		Policy:     policy.New(psql, logger),
		Relation:   relation.New(psql, logger),
		Scope:      scope.New(psql, logger),
	}
}
