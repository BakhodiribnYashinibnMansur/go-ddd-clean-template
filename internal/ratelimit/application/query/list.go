package query

import (
	"context"

	appdto "gct/internal/ratelimit/application"
	"gct/internal/ratelimit/domain"
	"gct/internal/shared/infrastructure/pgxutil"
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
}

// NewListRateLimitsHandler creates a new ListRateLimitsHandler.
func NewListRateLimitsHandler(readRepo domain.RateLimitReadRepository) *ListRateLimitsHandler {
	return &ListRateLimitsHandler{readRepo: readRepo}
}

// Handle executes the ListRateLimitsQuery and returns a list of RateLimitView with total count.
func (h *ListRateLimitsHandler) Handle(ctx context.Context, q ListRateLimitsQuery) (result *ListRateLimitsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListRateLimitsHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
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
