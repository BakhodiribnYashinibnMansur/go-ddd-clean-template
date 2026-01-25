package systemerror

import (
	"context"

	"gct/internal/repo/schema"

	"github.com/Masterminds/squirrel"
)

// List retrieves a list of system errors with optional filters
func (r *Repo) List(ctx context.Context, filter ListFilter) ([]*SystemError, error) {
	builder := r.db.Builder.
		Select(
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
		).
		From(schema.TableSystemError)

	if filter.Code != nil {
		builder = builder.Where(squirrel.Eq{schema.SystemErrorCode: *filter.Code})
	}

	if filter.Severity != nil {
		builder = builder.Where(squirrel.Eq{schema.SystemErrorSeverity: *filter.Severity})
	}

	if filter.IsResolved != nil {
		builder = builder.Where(squirrel.Eq{schema.SystemErrorIsResolved: *filter.IsResolved})
	}

	builder = builder.OrderBy(schema.SystemErrorCreatedAt + " DESC")

	if filter.Pagination != nil {
		if filter.Pagination.Limit > 0 {
			builder = builder.Limit(uint64(filter.Pagination.Limit))
		}
		if filter.Pagination.Offset > 0 {
			builder = builder.Offset(uint64(filter.Pagination.Offset))
		}
	}

	query, args, err := builder.ToSql()
	if err != nil {
		r.logger.Error("failed to build list query", "error", err)
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list system errors", "error", err)
		return nil, err
	}
	defer rows.Close()

	var errors []*SystemError
	for rows.Next() {
		var se SystemError
		err := rows.Scan(
			&se.ID,
			&se.Code,
			&se.Message,
			&se.StackTrace,
			&se.Metadata,
			&se.Severity,
			&se.ServiceName,
			&se.RequestID,
			&se.UserID,
			&se.IPAddress,
			&se.Path,
			&se.Method,
			&se.IsResolved,
			&se.ResolvedAt,
			&se.ResolvedBy,
			&se.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan system error", "error", err)
			return nil, err
		}
		errors = append(errors, &se)
	}

	return errors, nil
}
