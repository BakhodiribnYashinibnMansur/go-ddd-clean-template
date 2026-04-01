package command

import (
	"context"

	"gct/internal/dataexport/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpdateDataExportCommand represents a state transition for an in-progress data export.
// Status drives the export through its lifecycle: pending -> processing -> completed|failed.
// FileURL is required when completing; Error is required when failing. Nil fields are ignored.
type UpdateDataExportCommand struct {
	ID      uuid.UUID
	Status  *string
	FileURL *string
	Error   *string
}

// UpdateDataExportHandler drives data export state transitions and emits lifecycle events.
// The handler delegates state-machine logic to the domain aggregate (StartProcessing, Complete, Fail).
// Event publish failures are logged but do not roll back the status change.
type UpdateDataExportHandler struct {
	repo     domain.DataExportRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateDataExportHandler wires dependencies for data export status updates.
func NewUpdateDataExportHandler(
	repo domain.DataExportRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateDataExportHandler {
	return &UpdateDataExportHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the export by ID, applies the status transition, persists, and publishes lifecycle events.
// Returns a repository error if the export is not found or the update fails.
func (h *UpdateDataExportHandler) Handle(ctx context.Context, cmd UpdateDataExportCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateDataExportHandler.Handle")
	defer func() { end(err) }()

	de, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if cmd.Status != nil {
		switch *cmd.Status {
		case domain.ExportStatusProcessing:
			de.StartProcessing()
		case domain.ExportStatusCompleted:
			fileURL := ""
			if cmd.FileURL != nil {
				fileURL = *cmd.FileURL
			}
			de.Complete(fileURL)
		case domain.ExportStatusFailed:
			errMsg := ""
			if cmd.Error != nil {
				errMsg = *cmd.Error
			}
			de.Fail(errMsg)
		}
	}

	if err := h.repo.Update(ctx, de); err != nil {
		h.logger.Errorf("failed to update data export: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, de.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
