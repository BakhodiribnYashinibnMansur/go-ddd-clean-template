package notification

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	sql, args, err := r.builder.
		Select("id", "title", "body", "type", "target_type", "is_active", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var n domain.Notification
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&n.ID, &n.Title, &n.Body, &n.Type, &n.TargetType, &n.IsActive, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &n, nil
}
