package scope

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, scope *domain.Scope) error {
	sql, args, err := r.builder.
		Insert("scope").
		Columns("path", "method", "created_at").
		Values(scope.Path, scope.Method, time.Now()).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&scope.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "scope", nil)
	}

	return nil
}
