package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/ops/generic/metric/application/dto"
	metricrepo "gct/internal/context/ops/generic/metric/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListMetricsQuery holds the input for listing function metrics with filtering.
type ListMetricsQuery struct {
	Filter metricrepo.MetricFilter
}

// ListMetricsResult holds the output of the list metrics query.
type ListMetricsResult struct {
	Metrics []*dto.MetricView
	Total   int64
}

// ListMetricsHandler handles the ListMetricsQuery.
type ListMetricsHandler struct {
	readRepo metricrepo.MetricReadRepository
	logger   logger.Log
}

// NewListMetricsHandler creates a new ListMetricsHandler.
func NewListMetricsHandler(readRepo metricrepo.MetricReadRepository, l logger.Log) *ListMetricsHandler {
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

	result := make([]*dto.MetricView, len(views))
	for i, v := range views {
		result[i] = &dto.MetricView{
			ID:         uuid.UUID(v.ID),
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
