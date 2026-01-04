package scope

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Delete(ctx context.Context, path, method string) error {
	sql, args, err := r.builder.
		Delete("scope").
		Where(squirrel.Eq{"path": path}).
		Where(squirrel.Eq{"method": method}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "scope", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound, "scope not found")
	}

	return nil
}
