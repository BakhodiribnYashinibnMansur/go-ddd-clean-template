package ratelimit

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, rl *domain.RateLimit) error {
	sql, args, err := r.builder.
		Update(table).
		Set("name", rl.Name).
		Set("path_pattern", rl.PathPattern).
		Set("method", rl.Method).
		Set("limit_count", rl.LimitCount).
		Set("window_seconds", rl.WindowSeconds).
		Set("is_active", rl.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": rl.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&rl.UpdatedAt)
}
