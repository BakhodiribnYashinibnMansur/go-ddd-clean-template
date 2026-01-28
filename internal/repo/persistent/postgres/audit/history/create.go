package history

import (
	"context"

	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, h *domain.EndpointHistory) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
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
