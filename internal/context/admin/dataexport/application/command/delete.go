package command

import (
	"context"

	"gct/internal/context/admin/dataexport/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteDataExportCommand represents an intent to permanently remove a data export record.
// This deletes the metadata only — callers are responsible for cleaning up the exported file from storage.
type DeleteDataExportCommand struct {
	ID domain.DataExportID
}

// DeleteDataExportHandler performs hard deletion of a data export record via the repository.
// No domain events are emitted — callers needing file cleanup should handle that at a higher layer.
type DeleteDataExportHandler struct {
	repo   domain.DataExportRepository
	logger logger.Log
}

// NewDeleteDataExportHandler wires dependencies for data export deletion.
func NewDeleteDataExportHandler(
	repo domain.DataExportRepository,
	logger logger.Log,
) *DeleteDataExportHandler {
	return &DeleteDataExportHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the data export record identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeleteDataExportHandler) Handle(ctx context.Context, cmd DeleteDataExportCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteDataExportHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteDataExport", "data_export")()

	if err := h.repo.Delete(ctx, cmd.ID.UUID()); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "DeleteDataExport", Entity: "data_export", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}
	return nil
}
