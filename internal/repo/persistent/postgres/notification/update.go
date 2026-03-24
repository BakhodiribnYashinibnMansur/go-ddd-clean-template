package notification

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, n *domain.Notification) error {
	sql, args, err := r.builder.
		Update(table).
		Set("title", n.Title).
		Set("body", n.Body).
		Set("type", n.Type).
		Set("target_type", n.TargetType).
		Set("is_active", n.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": n.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&n.UpdatedAt)
}
