package notification

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Create(ctx context.Context, n *domain.Notification) error {
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "title", "body", "type", "target_type", "is_active").
		Values(n.ID, n.Title, n.Body, n.Type, n.TargetType, n.IsActive).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&n.CreatedAt, &n.UpdatedAt)
}
