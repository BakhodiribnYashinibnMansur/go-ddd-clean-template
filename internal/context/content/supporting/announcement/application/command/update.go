package command

import (
	"context"
	"time"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UpdateAnnouncementCommand represents a partial update to an existing announcement.
// Nil pointer fields are skipped (no change), while Publish triggers a one-way state transition
// from draft to published — already-published announcements ignore this flag.
type UpdateAnnouncementCommand struct {
	ID        announceentity.AnnouncementID
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
	repo     announcerepo.AnnouncementRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateAnnouncementHandler wires dependencies for announcement updates.
func NewUpdateAnnouncementHandler(
	repo announcerepo.AnnouncementRepository,
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
func (h *UpdateAnnouncementHandler) Handle(ctx context.Context, cmd UpdateAnnouncementCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateAnnouncementHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateAnnouncement", "announcement")()

	a, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	a.Update(cmd.Title, cmd.Content, cmd.Priority, cmd.StartDate, cmd.EndDate)

	if cmd.Publish && !a.Published() {
		a.Publish()
	}

	if err := h.repo.Update(ctx, a); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateAnnouncement", Entity: "announcement", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, a.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateAnnouncement", Entity: "announcement", Err: err}.KV()...)
	}

	return nil
}
