package systemerror

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	query := r.builder.Select(
		schema.SystemErrorID,
		schema.SystemErrorCode,
		schema.SystemErrorMessage,
		schema.SystemErrorStackTrace,
		schema.SystemErrorMetadata,
		schema.SystemErrorSeverity,
		schema.SystemErrorServiceName,
		schema.SystemErrorRequestID,
		schema.SystemErrorUserID,
		schema.SystemErrorIPAddress,
		schema.SystemErrorPath,
		schema.SystemErrorMethod,
		schema.SystemErrorIsResolved,
		schema.SystemErrorResolvedAt,
		schema.SystemErrorResolvedBy,
		schema.SystemErrorCreatedAt,
	).From(tableName)

	if filter.Code != nil {
		query = query.Where(squirrel.Eq{schema.SystemErrorCode: filter.Code})
	}
	if filter.Severity != nil {
		query = query.Where(squirrel.Eq{schema.SystemErrorSeverity: filter.Severity})
	}
	if filter.IsResolved != nil {
		query = query.Where(squirrel.Eq{schema.SystemErrorIsResolved: filter.IsResolved})
	}
	if filter.RequestID != nil {
		query = query.Where(squirrel.Eq{schema.SystemErrorRequestID: filter.RequestID})
	}
	if filter.UserID != nil {
		query = query.Where(squirrel.Eq{schema.SystemErrorUserID: filter.UserID})
	}
	if filter.FromDate != nil {
		query = query.Where(squirrel.GtOrEq{schema.SystemErrorCreatedAt: filter.FromDate})
	}
	if filter.ToDate != nil {
		query = query.Where(squirrel.LtOrEq{schema.SystemErrorCreatedAt: filter.ToDate})
	}

	// Count
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	if filter.Code != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SystemErrorCode: filter.Code})
	}
	if filter.Severity != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SystemErrorSeverity: filter.Severity})
	}
	if filter.IsResolved != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SystemErrorIsResolved: filter.IsResolved})
	}
	if filter.RequestID != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SystemErrorRequestID: filter.RequestID})
	}
	if filter.UserID != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SystemErrorUserID: filter.UserID})
	}
	if filter.FromDate != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{schema.SystemErrorCreatedAt: filter.FromDate})
	}
	if filter.ToDate != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{schema.SystemErrorCreatedAt: filter.ToDate})
	}

	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
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

	// Order by
	query = query.OrderBy(schema.SystemErrorCreatedAt + " DESC")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
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
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToScanRow)
		}
		errors = append(errors, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	return errors, count, nil
}
