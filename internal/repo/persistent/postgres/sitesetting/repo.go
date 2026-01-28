package sitesetting

import (
	"gct/consts"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableSiteSetting

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
