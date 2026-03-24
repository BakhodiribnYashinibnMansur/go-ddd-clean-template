package emailtemplate

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Delete(ctx context.Context, id string) error {
	sql, args, _ := r.builder.Delete(table).Where(squirrel.Eq{"id": id}).ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	return apperrors.HandlePgError(err, table, nil)
}
