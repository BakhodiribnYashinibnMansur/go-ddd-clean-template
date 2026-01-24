package sitesetting

import (
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = "site_settings"

type Repo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
