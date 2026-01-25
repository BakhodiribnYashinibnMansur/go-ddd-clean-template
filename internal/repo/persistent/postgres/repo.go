package postgres

import (
	"context"

	"gct/internal/repo/persistent/postgres/audit"
	"gct/internal/repo/persistent/postgres/authz"
	errorcode "gct/internal/repo/persistent/postgres/errorcode"
	sitesetting "gct/internal/repo/persistent/postgres/sitesetting"
	systemerror "gct/internal/repo/persistent/postgres/systemerror"
	"gct/internal/repo/persistent/postgres/user"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
)

type Repo struct {
	User        *user.User
	Authz       *authz.Authz
	Audit       *audit.Audit
	SiteSetting *sitesetting.Repo
	SystemError *systemerror.Repo
	ErrorCode   *errorcode.Repo
	DB          *postgres.Postgres
}

func New(pg *postgres.Postgres, logger logger.Log) (*Repo, error) {
	pg.Builder.PlaceholderFormat(squirrel.Dollar)
	return &Repo{
		User:        user.New(pg, logger),
		Authz:       authz.New(pg, logger),
		Audit:       audit.New(pg, logger),
		SiteSetting: sitesetting.New(pg.Pool),
		SystemError: systemerror.New(pg, logger),
		ErrorCode:   errorcode.New(pg, logger),
		DB:          pg,
	}, nil
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.DB.Pool.Ping(ctx)
}
