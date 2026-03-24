package announcement

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Create(ctx context.Context, a *domain.Announcement) error {
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "title", "content", "type", "is_active", "starts_at", "ends_at").
		Values(a.ID, a.Title, a.Content, a.Type, a.IsActive, a.StartsAt, a.EndsAt).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&a.CreatedAt, &a.UpdatedAt)
}
