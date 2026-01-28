package scope

import (
	"context"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.ScopesFilter) ([]*domain.Scope, int, error) {
	query := r.builder.Select("path", "method", "created_at").From(tableName)
	countQuery := r.builder.Select("COUNT(*)").From(tableName)

	if filter.Path != nil {
		query = query.Where(squirrel.Eq{"path": *filter.Path})
		countQuery = countQuery.Where(squirrel.Eq{"path": *filter.Path})
	}
	if filter.Method != nil {
		query = query.Where(squirrel.Eq{"method": *filter.Method})
		countQuery = countQuery.Where(squirrel.Eq{"method": *filter.Method})
	}

	if filter.Pagination != nil {
		query = query.Limit(uint64(filter.Pagination.Limit)).Offset(uint64(filter.Pagination.Offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var scopes []*domain.Scope
	for rows.Next() {
		var s domain.Scope
		if err := rows.Scan(&s.Path, &s.Method, &s.CreatedAt); err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %v", err))
		}
		scopes = append(scopes, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	var count int
	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build count query")
	}
	err = r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	return scopes, count, nil
}
