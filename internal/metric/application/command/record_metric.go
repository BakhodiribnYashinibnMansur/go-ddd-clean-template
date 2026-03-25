package command

import (
	"context"

	"gct/internal/metric/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// RecordMetricCommand holds the input for recording a new function metric.
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
func (h *RecordMetricHandler) Handle(ctx context.Context, cmd RecordMetricCommand) error {
	fm := domain.NewFunctionMetric(cmd.Name, cmd.LatencyMs, cmd.IsPanic, cmd.PanicError)

	if err := h.repo.Save(ctx, fm); err != nil {
		h.logger.Errorf("failed to save function metric: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, fm.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
