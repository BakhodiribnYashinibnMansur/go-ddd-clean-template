package command

import (
	"context"

	"gct/internal/dataexport/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreateDataExportCommand holds the input for creating a new data export.
type CreateDataExportCommand struct {
	UserID   uuid.UUID
	DataType string
	Format   string
}

// CreateDataExportHandler handles the CreateDataExportCommand.
type CreateDataExportHandler struct {
	repo     domain.DataExportRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateDataExportHandler creates a new CreateDataExportHandler.
func NewCreateDataExportHandler(
	repo domain.DataExportRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateDataExportHandler {
	return &CreateDataExportHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateDataExportCommand.
func (h *CreateDataExportHandler) Handle(ctx context.Context, cmd CreateDataExportCommand) error {
	de := domain.NewDataExport(cmd.UserID, cmd.DataType, cmd.Format)

	if err := h.repo.Save(ctx, de); err != nil {
		h.logger.Errorf("failed to save data export: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, de.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
