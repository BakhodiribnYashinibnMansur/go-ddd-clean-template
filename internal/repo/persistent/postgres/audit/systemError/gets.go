package systemError

import (
	"context"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	query := r.builder.Select(
		"id",
		"code",
		"message",
		"stack_trace",
		"metadata",
		"severity",
		"service_name",
		"request_id",
		"user_id",
		"ip_address",
		"path",
		"method",
		"is_resolved",
		"resolved_at",
		"resolved_by",
		"created_at",
	).From(tableName)

	if filter.Code != nil {
		query = query.Where(squirrel.Eq{"code": filter.Code})
	}
	if filter.Severity != nil {
		query = query.Where(squirrel.Eq{"severity": filter.Severity})
	}
	if filter.IsResolved != nil {
		query = query.Where(squirrel.Eq{"is_resolved": filter.IsResolved})
	}
	if filter.RequestID != nil {
		query = query.Where(squirrel.Eq{"request_id": filter.RequestID})
	}
	if filter.UserID != nil {
		query = query.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	if filter.FromDate != nil {
		query = query.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		query = query.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	// Count
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	if filter.Code != nil {
		countQuery = countQuery.Where(squirrel.Eq{"code": filter.Code})
	}
	if filter.Severity != nil {
		countQuery = countQuery.Where(squirrel.Eq{"severity": filter.Severity})
	}
	if filter.IsResolved != nil {
		countQuery = countQuery.Where(squirrel.Eq{"is_resolved": filter.IsResolved})
	}
	if filter.RequestID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"request_id": filter.RequestID})
	}
	if filter.UserID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	if filter.FromDate != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build count query")
	}

	var count int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, tableName, nil)
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

	// Order by
	query = query.OrderBy("created_at DESC")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, tableName, nil)
	}
	defer rows.Close()

	var errors []*domain.SystemError
	for rows.Next() {
		var e domain.SystemError
		err = rows.Scan(
			&e.ID,
			&e.Code,
			&e.Message,
			&e.StackTrace,
			&e.Metadata,
			&e.Severity,
			&e.ServiceName,
			&e.RequestID,
			&e.UserID,
			&e.IPAddress,
			&e.Path,
			&e.Method,
			&e.IsResolved,
			&e.ResolvedAt,
			&e.ResolvedBy,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %s", err))
		}
		errors = append(errors, &e)
	}

	return errors, count, nil
}
