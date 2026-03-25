package command

import (
	"context"

	"gct/internal/dataexport/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteDataExportCommand holds the input for deleting a data export.
type DeleteDataExportCommand struct {
	ID uuid.UUID
}

// DeleteDataExportHandler handles the DeleteDataExportCommand.
type DeleteDataExportHandler struct {
	repo   domain.DataExportRepository
	logger logger.Log
}

// NewDeleteDataExportHandler creates a new DeleteDataExportHandler.
func NewDeleteDataExportHandler(
	repo domain.DataExportRepository,
	logger logger.Log,
) *DeleteDataExportHandler {
	return &DeleteDataExportHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteDataExportCommand.
func (h *DeleteDataExportHandler) Handle(ctx context.Context, cmd DeleteDataExportCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete data export: %v", err)
		return err
	}
	return nil
}
