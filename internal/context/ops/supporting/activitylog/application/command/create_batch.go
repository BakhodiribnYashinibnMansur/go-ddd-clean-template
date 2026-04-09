package command

import (
	"context"

	"gct/internal/context/ops/supporting/activitylog/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// CreateActivityLogBatchCommand carries a batch of activity log entries to persist.
type CreateActivityLogBatchCommand struct {
	Entries []*domain.ActivityLogEntry
}

// CreateActivityLogBatchHandler persists a batch of activity log entries.
type CreateActivityLogBatchHandler struct {
	repo   domain.ActivityLogWriteRepository
	logger logger.Log
}

// NewCreateActivityLogBatchHandler wires dependencies.
func NewCreateActivityLogBatchHandler(
	repo domain.ActivityLogWriteRepository,
	l logger.Log,
) *CreateActivityLogBatchHandler {
	return &CreateActivityLogBatchHandler{repo: repo, logger: l}
}

// Handle persists the batch of activity log entries.
func (h *CreateActivityLogBatchHandler) Handle(ctx context.Context, cmd CreateActivityLogBatchCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateActivityLogBatchHandler.Handle")
	defer func() { end(err) }()

	if len(cmd.Entries) == 0 {
		return nil
	}

	if err := h.repo.SaveBatch(ctx, cmd.Entries); err != nil {
		h.logger.Errorc(ctx, "failed to save activity log batch",
			logger.F{Op: "CreateActivityLogBatch", Entity: "activity_log", Err: err}.KV()...)
		return err
	}

	return nil
}
