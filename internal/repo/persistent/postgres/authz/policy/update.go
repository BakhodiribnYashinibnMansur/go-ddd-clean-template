package policy

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, p *domain.Policy) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	sql, args, err := r.builder.
		Update(tableName).
		Set("permission_id", p.PermissionID).
		Set("effect", p.Effect).
		Set("priority", p.Priority).
		Set("active", p.Active).
		Set("conditions", p.Conditions).
		Where(squirrel.Eq{"id": p.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build update query")
	}

	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "policy not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to commit transaction")
	}

	return nil
}
