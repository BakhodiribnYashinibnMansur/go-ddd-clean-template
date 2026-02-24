package ratelimit

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, rl *domain.RateLimit) error {
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "name", "path_pattern", "method", "limit_count", "window_seconds", "is_active").
		Values(rl.ID, rl.Name, rl.PathPattern, rl.Method, rl.LimitCount, rl.WindowSeconds, rl.IsActive).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&rl.CreatedAt, &rl.UpdatedAt)
}
