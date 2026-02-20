package systemerror

import (
	"context"


	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// GetByID retrieves a system error by its ID
func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*SystemError, error) {
	query, args, err := r.db.Builder.
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
		From("system_errors").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		r.logger.Error("failed to build get query", "error", err)
		return nil, err
	}

	var se SystemError
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(
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
		r.logger.Error("failed to get system error by ID", "error", err, "id", id)
		return nil, err
	}

	return &se, nil
}
