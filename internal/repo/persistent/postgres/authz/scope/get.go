package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.ScopeFilter) (*domain.Scope, error) {
	query := r.builder.Select("path", "method", "created_at").From("scope")

	if filter.Path != nil {
		query = query.Where(squirrel.Eq{"path": *filter.Path})
	}
	if filter.Method != nil {
		query = query.Where(squirrel.Eq{"method": *filter.Method})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	var scope domain.Scope
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&scope.Path, &scope.Method, &scope.CreatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "scope", nil)
	}

	return &scope, nil
}
