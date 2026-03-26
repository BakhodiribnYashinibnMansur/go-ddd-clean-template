package command

import (
	"context"
	"time"

	"gct/internal/announcement/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateAnnouncementCommand represents a partial update to an existing announcement.
// Nil pointer fields are skipped (no change), while Publish triggers a one-way state transition
// from draft to published — already-published announcements ignore this flag.
type UpdateAnnouncementCommand struct {
	ID        uuid.UUID
	Title     *shared.Lang
	Content   *shared.Lang
	Priority  *int
	StartDate *time.Time
	EndDate   *time.Time
	Publish   bool
}

// UpdateAnnouncementHandler applies partial updates and optional publication to an announcement.
// Event publish failures are logged but do not cause the handler to return an error.
type UpdateAnnouncementHandler struct {
	repo     domain.AnnouncementRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateAnnouncementHandler wires dependencies for announcement updates.
func NewUpdateAnnouncementHandler(
	repo domain.AnnouncementRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateAnnouncementHandler {
	return &UpdateAnnouncementHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the announcement by ID, applies field-level changes, optionally publishes it, then persists.
// Returns a repository error if the announcement is not found or the update fails.
func (h *UpdateAnnouncementHandler) Handle(ctx context.Context, cmd UpdateAnnouncementCommand) error {
	a, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	a.Update(cmd.Title, cmd.Content, cmd.Priority, cmd.StartDate, cmd.EndDate)

	if cmd.Publish && !a.Published() {
		a.Publish()
	}

	if err := h.repo.Update(ctx, a); err != nil {
		h.logger.Errorf("failed to update announcement: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, a.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
