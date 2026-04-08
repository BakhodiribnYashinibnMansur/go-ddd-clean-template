package query

import (
	"context"

	"gct/internal/context/ops/supporting/activitylog/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListActivityLogsQuery holds the input for listing activity logs with filtering.
type ListActivityLogsQuery struct {
	Filter domain.ActivityLogFilter
}

// ListActivityLogsResult holds the output of the list activity logs query.
type ListActivityLogsResult struct {
	Entries []*domain.ActivityLogView
	Total   int64
}

// ListActivityLogsHandler handles the ListActivityLogsQuery.
type ListActivityLogsHandler struct {
	readRepo domain.ActivityLogReadRepository
	logger   logger.Log
}

// NewListActivityLogsHandler creates a new ListActivityLogsHandler.
func NewListActivityLogsHandler(readRepo domain.ActivityLogReadRepository, l logger.Log) *ListActivityLogsHandler {
	return &ListActivityLogsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the query and returns a list of activity log views with total count.
func (h *ListActivityLogsHandler) Handle(ctx context.Context, q ListActivityLogsQuery) (_ *ListActivityLogsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListActivityLogsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListActivityLogs", "activity_log")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListActivityLogs", Entity: "activity_log", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &ListActivityLogsResult{
		Entries: views,
		Total:   total,
	}, nil
}
