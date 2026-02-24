package ratelimit

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error) {
	sql, args, err := r.builder.
		Select("id", "name", "path_pattern", "method", "limit_count", "window_seconds", "is_active", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var rl domain.RateLimit
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&rl.ID, &rl.Name, &rl.PathPattern, &rl.Method, &rl.LimitCount, &rl.WindowSeconds,
		&rl.IsActive, &rl.CreatedAt, &rl.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &rl, nil
}
