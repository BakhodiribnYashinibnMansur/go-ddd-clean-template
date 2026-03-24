package announcement

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	sql, args, err := r.builder.
		Select("id", "title", "content", "type", "is_active", "starts_at", "ends_at", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var a domain.Announcement
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&a.ID, &a.Title, &a.Content, &a.Type, &a.IsActive,
		&a.StartsAt, &a.EndsAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &a, nil
}
