package command

import (
	"context"

	"gct/internal/announcement/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteAnnouncementCommand represents an intent to permanently remove an announcement.
// Once deleted, the announcement is no longer visible to any audience.
type DeleteAnnouncementCommand struct {
	ID uuid.UUID
}

// DeleteAnnouncementHandler performs hard deletion of announcements via the repository.
// No domain events are emitted — callers needing audit trails should record them separately.
type DeleteAnnouncementHandler struct {
	repo   domain.AnnouncementRepository
	logger logger.Log
}

// NewDeleteAnnouncementHandler wires dependencies for announcement deletion.
func NewDeleteAnnouncementHandler(
	repo domain.AnnouncementRepository,
	logger logger.Log,
) *DeleteAnnouncementHandler {
	return &DeleteAnnouncementHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the announcement identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteAnnouncementHandler) Handle(ctx context.Context, cmd DeleteAnnouncementCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete announcement: %v", err)
		return err
	}
	return nil
}
