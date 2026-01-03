package permission

import (
	"context"

	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete("permission").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "permission", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound, "permission not found")
	}

	return nil
}
