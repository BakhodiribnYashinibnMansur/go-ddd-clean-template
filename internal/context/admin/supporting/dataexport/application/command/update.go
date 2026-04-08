package command

import (
	"context"

	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	exportrepo "gct/internal/context/admin/supporting/dataexport/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// UpdateDataExportCommand represents a state transition for an in-progress data export.
// Status drives the export through its lifecycle: pending -> processing -> completed|failed.
// FileURL is required when completing; Error is required when failing. Nil fields are ignored.
type UpdateDataExportCommand struct {
	ID      exportentity.DataExportID
	Status  *string
	FileURL *string
	Error   *string
}

// UpdateDataExportHandler drives data export state transitions and emits lifecycle events.
// The handler delegates state-machine logic to the domain aggregate (StartProcessing, Complete, Fail).
type UpdateDataExportHandler struct {
	repo      exportrepo.DataExportRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewUpdateDataExportHandler wires dependencies for data export status updates.
func NewUpdateDataExportHandler(
	repo exportrepo.DataExportRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *UpdateDataExportHandler {
	return &UpdateDataExportHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle fetches the export by ID, applies the status transition, persists, and publishes lifecycle events.
// Returns a repository error if the export is not found or the update fails.
func (h *UpdateDataExportHandler) Handle(ctx context.Context, cmd UpdateDataExportCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateDataExportHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateDataExport", "data_export")()

	de, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	if cmd.Status != nil {
		switch *cmd.Status {
		case exportentity.ExportStatusProcessing:
			de.StartProcessing()
		case exportentity.ExportStatusCompleted:
			fileURL := ""
			if cmd.FileURL != nil {
				fileURL = *cmd.FileURL
			}
			de.Complete(fileURL)
		case exportentity.ExportStatusFailed:
			errMsg := ""
			if cmd.Error != nil {
				errMsg = *cmd.Error
			}
			de.Fail(errMsg)
		}
	}

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, de); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateDataExport", Entity: "data_export", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, de.Events)
}
