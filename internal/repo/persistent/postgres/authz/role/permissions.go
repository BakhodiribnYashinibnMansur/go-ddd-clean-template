package role

import (
	"context"
	"time"

	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	sql, args, err := r.builder.
		Insert("role_permission").
		Columns("role_id", "permission_id", "created_at").
		Values(roleID, permID, time.Now()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, "role_permission", nil)
	}

	return nil
}

func (r *Repo) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete("role_permission").
		Where(squirrel.Eq{"role_id": roleID}).
		Where(squirrel.Eq{"permission_id": permID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, "role_permission", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "role permission not found")
	}

	return nil
}
