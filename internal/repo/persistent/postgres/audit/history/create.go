package history

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, h *domain.EndpointHistory) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			"user_id",
			"session_id",
			"method",
			"path",
			"status_code",
			"duration_ms",
			"platform",
			"ip_address",
			"user_agent",
			"permission",
			"decision",
			"request_id",
			"rate_limited",
			"response_size",
			"error_message",
			"created_at",
		).
		Values(
			h.UserID,
			h.SessionID,
			h.Method,
			h.Path,
			h.StatusCode,
			h.DurationMs,
			h.Platform,
			h.IPAddress,
			h.UserAgent,
			h.Permission,
			h.Decision,
			h.RequestID,
			h.RateLimited,
			h.ResponseSize,
			h.ErrorMessage,
			h.CreatedAt,
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
