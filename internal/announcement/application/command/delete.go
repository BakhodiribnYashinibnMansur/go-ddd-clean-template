package command

import (
	"context"

	"gct/internal/announcement/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteAnnouncementCommand holds the input for deleting an announcement.
type DeleteAnnouncementCommand struct {
	ID uuid.UUID
}

// DeleteAnnouncementHandler handles the DeleteAnnouncementCommand.
type DeleteAnnouncementHandler struct {
	repo   domain.AnnouncementRepository
	logger logger.Log
}

// NewDeleteAnnouncementHandler creates a new DeleteAnnouncementHandler.
func NewDeleteAnnouncementHandler(
	repo domain.AnnouncementRepository,
	logger logger.Log,
) *DeleteAnnouncementHandler {
	return &DeleteAnnouncementHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteAnnouncementCommand.
func (h *DeleteAnnouncementHandler) Handle(ctx context.Context, cmd DeleteAnnouncementCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete announcement: %v", err)
		return err
	}
	return nil
}
