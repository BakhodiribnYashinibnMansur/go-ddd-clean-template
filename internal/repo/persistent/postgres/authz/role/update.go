package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, role *domain.Role) error {
	sql, args, err := r.builder.
		Update("role").
		Set("name", role.Name).
		Where(squirrel.Eq{"id": role.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build update query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "role", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound, "role not found")
	}

	return nil
}
