package history

import (
	"context"
	"fmt"

	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.EndpointHistoriesFilter) ([]*domain.EndpointHistory, int, error) {
	query := r.builder.Select(
		schema.EndpointHistoryID,
		schema.EndpointHistoryUserID,
		schema.EndpointHistorySessionID,
		schema.EndpointHistoryMethod,
		schema.EndpointHistoryPath,
		schema.EndpointHistoryStatusCode,
		schema.EndpointHistoryDurationMs,
		schema.EndpointHistoryPlatform,
		schema.EndpointHistoryIPAddress,
		schema.EndpointHistoryUserAgent,
		schema.EndpointHistoryPermission,
		schema.EndpointHistoryDecision,
		schema.EndpointHistoryRequestID,
		schema.EndpointHistoryRateLimited,
		schema.EndpointHistoryResponseSize,
		schema.EndpointHistoryErrorMessage,
		schema.EndpointHistoryCreatedAt,
	).From(tableName)

	if filter.UserID != nil {
		query = query.Where(squirrel.Eq{schema.EndpointHistoryUserID: filter.UserID})
	}
	if filter.Method != nil {
		query = query.Where(squirrel.Eq{schema.EndpointHistoryMethod: filter.Method})
	}
	if filter.Path != nil {
		query = query.Where(squirrel.Eq{schema.EndpointHistoryPath: filter.Path})
	}
	if filter.StatusCode != nil {
		query = query.Where(squirrel.Eq{schema.EndpointHistoryStatusCode: filter.StatusCode})
	}
	if filter.FromDate != nil {
		query = query.Where(squirrel.GtOrEq{schema.EndpointHistoryCreatedAt: filter.FromDate})
	}
	if filter.ToDate != nil {
		query = query.Where(squirrel.LtOrEq{schema.EndpointHistoryCreatedAt: filter.ToDate})
	}

	// Count total
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	if filter.UserID != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.EndpointHistoryUserID: filter.UserID})
	}
	if filter.Method != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.EndpointHistoryMethod: filter.Method})
	}
	if filter.Path != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.EndpointHistoryPath: filter.Path})
	}
	if filter.StatusCode != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.EndpointHistoryStatusCode: filter.StatusCode})
	}
	if filter.FromDate != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{schema.EndpointHistoryCreatedAt: filter.FromDate})
	}
	if filter.ToDate != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{schema.EndpointHistoryCreatedAt: filter.ToDate})
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

	// Apply pagination
	if filter.Pagination != nil {
		if filter.Pagination.Limit > 0 {
			query = query.Limit(uint64(filter.Pagination.Limit))
		}
		if filter.Pagination.Offset > 0 {
			query = query.Offset(uint64(filter.Pagination.Offset))
		}
	}

	// Order by created_at desc
	query = query.OrderBy(schema.EndpointHistoryCreatedAt + " DESC")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var histories []*domain.EndpointHistory
	for rows.Next() {
		var h domain.EndpointHistory
		err = rows.Scan(
			&h.ID,
			&h.UserID,
			&h.SessionID,
			&h.Method,
			&h.Path,
			&h.StatusCode,
			&h.DurationMs,
			&h.Platform,
			&h.IPAddress,
			&h.UserAgent,
			&h.Permission,
			&h.Decision,
			&h.RequestID,
			&h.RateLimited,
			&h.ResponseSize,
			&h.ErrorMessage,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %s", err))
		}
		histories = append(histories, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	return histories, count, nil
}
