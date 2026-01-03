package permission

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	apperrors "gct/pkg/errors"
)

func (r *Repo) AddScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	sql, args, err := r.builder.
		Insert("permission_scope").
		Columns("permission_id", "path", "method", "created_at").
		Values(permID, path, method, time.Now()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "permission_scope", nil)
	}

	return nil
}

func (r *Repo) RemoveScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	sql, args, err := r.builder.
		Delete("permission_scope").
		Where(squirrel.Eq{"permission_id": permID}).
		Where(squirrel.Eq{"path": path}).
		Where(squirrel.Eq{"method": method}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "permission_scope", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound, "permission scope not found")
	}

	return nil
}
