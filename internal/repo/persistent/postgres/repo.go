package postgres

import (
	"gct/internal/repo/persistent/postgres/audit"
	"gct/internal/repo/persistent/postgres/authz"
	"gct/internal/repo/persistent/postgres/user"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"
	"github.com/Masterminds/squirrel"
)

type Repo struct {
	User  *user.User
	Authz *authz.Authz
	Audit *audit.Audit
}

func New(pg *postgres.Postgres, logger logger.Log) (*Repo, error) {
	pg.Builder.PlaceholderFormat(squirrel.Dollar)
	return &Repo{
		User:  user.New(pg, logger),
		Authz: authz.New(pg, logger),
		Audit: audit.New(pg, logger),
	}, nil
}
