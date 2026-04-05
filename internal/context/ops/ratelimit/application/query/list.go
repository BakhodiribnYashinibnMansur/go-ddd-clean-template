package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/ops/ratelimit/application"
	"gct/internal/context/ops/ratelimit/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListRateLimitsQuery holds the input for listing rate limits.
type ListRateLimitsQuery struct {
	Filter domain.RateLimitFilter
}

// ListRateLimitsResult holds the output of the list rate limits query.
type ListRateLimitsResult struct {
	RateLimits []*appdto.RateLimitView
	Total      int64
}

// ListRateLimitsHandler handles the ListRateLimitsQuery.
type ListRateLimitsHandler struct {
	readRepo domain.RateLimitReadRepository
	logger   logger.Log
}

// NewListRateLimitsHandler creates a new ListRateLimitsHandler.
func NewListRateLimitsHandler(readRepo domain.RateLimitReadRepository, l logger.Log) *ListRateLimitsHandler {
	return &ListRateLimitsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListRateLimitsQuery and returns a list of RateLimitView with total count.
func (h *ListRateLimitsHandler) Handle(ctx context.Context, q ListRateLimitsQuery) (result *ListRateLimitsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListRateLimitsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListRateLimits", "rate_limit")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListRateLimits", Entity: "rate_limit", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*appdto.RateLimitView, len(views))
	for i, v := range views {
		items[i] = &appdto.RateLimitView{
			ID:                v.ID,
			Name:              v.Name,
			Rule:              v.Rule,
			RequestsPerWindow: v.RequestsPerWindow,
			WindowDuration:    v.WindowDuration,
			Enabled:           v.Enabled,
			CreatedAt:         v.CreatedAt,
			UpdatedAt:         v.UpdatedAt,
		}
	}

	return &ListRateLimitsResult{
		RateLimits: items,
		Total:      total,
	}, nil
}
