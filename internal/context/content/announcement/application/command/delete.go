package command

import (
	"context"

	"gct/internal/context/content/announcement/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteAnnouncementCommand represents an intent to permanently remove an announcement.
// Once deleted, the announcement is no longer visible to any audience.
type DeleteAnnouncementCommand struct {
	ID domain.AnnouncementID
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
func (h *DeleteAnnouncementHandler) Handle(ctx context.Context, cmd DeleteAnnouncementCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteAnnouncementHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteAnnouncement", "announcement")()

	if err := h.repo.Delete(ctx, cmd.ID.UUID()); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "DeleteAnnouncement", Entity: "announcement", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}
	return nil
}
