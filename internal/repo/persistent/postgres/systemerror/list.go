package systemerror

import (
	"context"


	"github.com/Masterminds/squirrel"
)

// List retrieves a list of system errors with optional filters
func (r *Repo) List(ctx context.Context, filter ListFilter) ([]*SystemError, error) {
	builder := r.db.Builder.
		Select(
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
		).
		From("system_errors")

	if filter.Code != nil {
		builder = builder.Where(squirrel.Eq{"code": *filter.Code})
	}

	if filter.Severity != nil {
		builder = builder.Where(squirrel.Eq{"severity": *filter.Severity})
	}

	if filter.IsResolved != nil {
		builder = builder.Where(squirrel.Eq{"is_resolved": *filter.IsResolved})
	}

	builder = builder.OrderBy("created_at" + " DESC")

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
