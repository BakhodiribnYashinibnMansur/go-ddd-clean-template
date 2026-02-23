package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, p *domain.Permission) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("parent_id", p.ParentID).
		Set("name", p.Name).
		Set("description", p.Description).
		Where(squirrel.Eq{"id": p.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build update query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "permission not found")
	}

	return nil
}
