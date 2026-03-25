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

// UpdateAnnouncementCommand holds the input for updating an announcement.
type UpdateAnnouncementCommand struct {
	ID        uuid.UUID
	Title     *shared.Lang
	Content   *shared.Lang
	Priority  *int
	StartDate *time.Time
	EndDate   *time.Time
	Publish   bool
}

// UpdateAnnouncementHandler handles the UpdateAnnouncementCommand.
type UpdateAnnouncementHandler struct {
	repo     domain.AnnouncementRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateAnnouncementHandler creates a new UpdateAnnouncementHandler.
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

// Handle executes the UpdateAnnouncementCommand.
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
