package command

import (
	"context"

	"gct/internal/context/ops/metric/domain"
	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
)

// RecordMetricCommand captures a single function execution observation for performance monitoring.
// IsPanic flags whether the function recovered from a panic; PanicError holds the recovered message if so.
// This is typically dispatched by middleware or instrumentation wrappers, not by business logic directly.
type RecordMetricCommand struct {
	Name       string
	LatencyMs  float64
	IsPanic    bool
	PanicError *string
}

// RecordMetricHandler handles the RecordMetricCommand.
type RecordMetricHandler struct {
	repo     domain.MetricRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewRecordMetricHandler creates a new RecordMetricHandler.
func NewRecordMetricHandler(
	repo domain.MetricRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *RecordMetricHandler {
	return &RecordMetricHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the RecordMetricCommand.
func (h *RecordMetricHandler) Handle(ctx context.Context, cmd RecordMetricCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "RecordMetricHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "RecordMetric", "metric")()

	fm := domain.NewFunctionMetric(cmd.Name, cmd.LatencyMs, cmd.IsPanic, cmd.PanicError)

	if err := h.repo.Save(ctx, fm); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "RecordMetric", Entity: "metric", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, fm.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "RecordMetric", Entity: "metric", Err: err}.KV()...)
	}

	return nil
}
