package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/ops/metric/application"
	"gct/internal/context/ops/metric/domain"
	"gct/internal/platform/infrastructure/pgxutil"
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
	logger   logger.Log
}

// NewListMetricsHandler creates a new ListMetricsHandler.
func NewListMetricsHandler(readRepo domain.MetricReadRepository, l logger.Log) *ListMetricsHandler {
	return &ListMetricsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListMetricsQuery and returns a list of MetricView with total count.
func (h *ListMetricsHandler) Handle(ctx context.Context, q ListMetricsQuery) (_ *ListMetricsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListMetricsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListMetrics", "metric")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListMetrics", Entity: "metric", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
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
