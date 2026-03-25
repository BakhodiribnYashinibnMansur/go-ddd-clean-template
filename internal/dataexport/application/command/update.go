package command

import (
	"context"

	"gct/internal/dataexport/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateDataExportCommand holds the input for updating a data export.
type UpdateDataExportCommand struct {
	ID      uuid.UUID
	Status  *string
	FileURL *string
	Error   *string
}

// UpdateDataExportHandler handles the UpdateDataExportCommand.
type UpdateDataExportHandler struct {
	repo     domain.DataExportRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateDataExportHandler creates a new UpdateDataExportHandler.
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

// Handle executes the UpdateDataExportCommand.
func (h *UpdateDataExportHandler) Handle(ctx context.Context, cmd UpdateDataExportCommand) error {
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
