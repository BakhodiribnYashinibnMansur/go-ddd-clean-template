package systemError

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, e *domain.SystemError) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
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
			"created_at",
		).
		Values(
			e.Code,
			e.Message,
			e.StackTrace,
			e.Metadata,
			e.Severity,
			e.ServiceName,
			e.RequestID,
			e.UserID,
			e.IPAddress,
			e.Path,
			e.Method,
			e.CreatedAt,
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
