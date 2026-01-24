package scope

import (
	"context"

	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Delete(ctx context.Context, path, method string) error {
	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"path": path}).
		Where(squirrel.Eq{"method": method}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "scope not found")
	}

	return nil
}
