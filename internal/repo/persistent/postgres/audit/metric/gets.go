package metric

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.FunctionMetricsFilter) ([]*domain.FunctionMetric, int, error) {
	query := r.builder.Select(
		"id",
		"name",
		"latency_ms",
		"is_panic",
		"panic_error",
		"created_at",
	).From(tableName)

	if filter.Name != nil {
		query = query.Where(squirrel.Eq{"name": filter.Name})
	}
	if filter.IsPanic != nil {
		query = query.Where(squirrel.Eq{"is_panic": filter.IsPanic})
	}
	if filter.FromDate != nil {
		query = query.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		query = query.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	// Count
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	if filter.Name != nil {
		countQuery = countQuery.Where(squirrel.Eq{"name": filter.Name})
	}
	if filter.IsPanic != nil {
		countQuery = countQuery.Where(squirrel.Eq{"is_panic": filter.IsPanic})
	}
	if filter.FromDate != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build count query")
	}

	var count int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Pagination
	if filter.Pagination != nil {
		if filter.Pagination.Limit > 0 {
			query = query.Limit(uint64(filter.Pagination.Limit))
		}
		if filter.Pagination.Offset > 0 {
			query = query.Offset(uint64(filter.Pagination.Offset))
		}
	}

	query = query.OrderBy("created_at DESC")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var metrics []*domain.FunctionMetric
	for rows.Next() {
		var m domain.FunctionMetric
		err = rows.Scan(
			&m.ID,
			&m.Name,
			&m.LatencyMs,
			&m.IsPanic,
			&m.PanicError,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to scan row")
		}
		metrics = append(metrics, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	return metrics, count, nil
}
