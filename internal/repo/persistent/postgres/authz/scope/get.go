package scope

import (
	"context"

	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.ScopeFilter) (*domain.Scope, error) {
	query := r.builder.Select(schema.ScopePath, schema.ScopeMethod, schema.ScopeCreatedAt).From(tableName)

	if filter.Path != nil {
		query = query.Where(squirrel.Eq{schema.ScopePath: *filter.Path})
	}
	if filter.Method != nil {
		query = query.Where(squirrel.Eq{schema.ScopeMethod: *filter.Method})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	var scope domain.Scope
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&scope.Path, &scope.Method, &scope.CreatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &scope, nil
}
