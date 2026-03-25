package command

import (
	"context"
	"time"

	"gct/internal/announcement/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/logger"
)

// CreateAnnouncementCommand holds the input for creating a new announcement.
type CreateAnnouncementCommand struct {
	Title     shared.Lang
	Content   shared.Lang
	Priority  int
	StartDate *time.Time
	EndDate   *time.Time
}

// CreateAnnouncementHandler handles the CreateAnnouncementCommand.
type CreateAnnouncementHandler struct {
	repo     domain.AnnouncementRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateAnnouncementHandler creates a new CreateAnnouncementHandler.
func NewCreateAnnouncementHandler(
	repo domain.AnnouncementRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateAnnouncementHandler {
	return &CreateAnnouncementHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateAnnouncementCommand.
func (h *CreateAnnouncementHandler) Handle(ctx context.Context, cmd CreateAnnouncementCommand) error {
	a := domain.NewAnnouncement(cmd.Title, cmd.Content, cmd.Priority, cmd.StartDate, cmd.EndDate)

	if err := h.repo.Save(ctx, a); err != nil {
		h.logger.Errorf("failed to save announcement: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, a.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
