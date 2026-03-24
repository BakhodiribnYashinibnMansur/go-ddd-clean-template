package announcement

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, a *domain.Announcement) error {
	sql, args, err := r.builder.
		Update(table).
		Set("title", a.Title).
		Set("content", a.Content).
		Set("type", a.Type).
		Set("is_active", a.IsActive).
		Set("starts_at", a.StartsAt).
		Set("ends_at", a.EndsAt).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": a.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&a.UpdatedAt)
}
