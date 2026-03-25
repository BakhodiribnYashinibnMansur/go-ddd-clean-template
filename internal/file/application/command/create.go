package command

import (
	"context"

	"gct/internal/file/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreateFileCommand holds the input for creating a new file record.
type CreateFileCommand struct {
	Name         string
	OriginalName string
	MimeType     string
	Size         int64
	Path         string
	URL          string
	UploadedBy   *uuid.UUID
}

// CreateFileHandler handles the CreateFileCommand.
type CreateFileHandler struct {
	repo     domain.FileRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateFileHandler creates a new CreateFileHandler.
func NewCreateFileHandler(
	repo domain.FileRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateFileHandler {
	return &CreateFileHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateFileCommand.
func (h *CreateFileHandler) Handle(ctx context.Context, cmd CreateFileCommand) error {
	f := domain.NewFile(cmd.Name, cmd.OriginalName, cmd.MimeType, cmd.Size, cmd.Path, cmd.URL, cmd.UploadedBy)

	if err := h.repo.Save(ctx, f); err != nil {
		h.logger.Errorf("failed to save file: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, f.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
