package command

import (
	"context"
	"time"

	"gct/internal/announcement/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
)

// CreateAnnouncementCommand represents an intent to create a new system announcement.
// Title and Content are multilingual; Priority controls display ordering (higher = more prominent).
// StartDate/EndDate define the visibility window — nil means unbounded on that side.
type CreateAnnouncementCommand struct {
	Title     shared.Lang
	Content   shared.Lang
	Priority  int
	StartDate *time.Time
	EndDate   *time.Time
}

// CreateAnnouncementHandler orchestrates announcement creation through the repository layer.
// It emits domain events after a successful save; event publish failures are logged but do not roll back the save.
type CreateAnnouncementHandler struct {
	repo     domain.AnnouncementRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateAnnouncementHandler wires dependencies for announcement creation.
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

// Handle persists a new announcement and publishes its domain events.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreateAnnouncementHandler) Handle(ctx context.Context, cmd CreateAnnouncementCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateAnnouncementHandler.Handle")
	defer func() { end(err) }()

	a := domain.NewAnnouncement(cmd.Title, cmd.Content, cmd.Priority, cmd.StartDate, cmd.EndDate)

	if err := h.repo.Save(ctx, a); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateAnnouncement", Entity: "announcement", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, a.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateAnnouncement", Entity: "announcement", Err: err}.KV()...)
	}

	return nil
}
