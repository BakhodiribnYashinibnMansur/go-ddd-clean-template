package query

import (
	"context"

	appdto "gct/internal/metric/application"
	"gct/internal/metric/domain"
	"gct/internal/shared/infrastructure/pgxutil"
)

// ListMetricsQuery holds the input for listing function metrics with filtering.
type ListMetricsQuery struct {
	Filter domain.MetricFilter
}

// ListMetricsResult holds the output of the list metrics query.
type ListMetricsResult struct {
	Metrics []*appdto.MetricView
	Total   int64
}

// ListMetricsHandler handles the ListMetricsQuery.
type ListMetricsHandler struct {
	readRepo domain.MetricReadRepository
}

// NewListMetricsHandler creates a new ListMetricsHandler.
func NewListMetricsHandler(readRepo domain.MetricReadRepository) *ListMetricsHandler {
	return &ListMetricsHandler{readRepo: readRepo}
}

// Handle executes the ListMetricsQuery and returns a list of MetricView with total count.
func (h *ListMetricsHandler) Handle(ctx context.Context, q ListMetricsQuery) (_ *ListMetricsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListMetricsHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.MetricView, len(views))
	for i, v := range views {
		result[i] = &appdto.MetricView{
			ID:         v.ID,
			Name:       v.Name,
			LatencyMs:  v.LatencyMs,
			IsPanic:    v.IsPanic,
			PanicError: v.PanicError,
			CreatedAt:  v.CreatedAt,
		}
	}

	return &ListMetricsResult{
		Metrics: result,
		Total:   total,
	}, nil
}
