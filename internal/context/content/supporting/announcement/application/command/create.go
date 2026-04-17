package command

import (
	"context"
	"fmt"
	"time"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// CreateAnnouncementCommand represents an intent to create a new system announcement.
// Title and Content are multilingual; Priority controls display ordering (higher = more prominent).
// StartDate/EndDate define the visibility window — nil means unbounded on that side.
type CreateAnnouncementCommand struct {
	Title     shareddomain.Lang
	Content   shareddomain.Lang
	Priority  int
	StartDate *time.Time
	EndDate   *time.Time
}

// CreateAnnouncementHandler orchestrates announcement creation through the repository layer.
// It emits domain events after a successful save; event publish failures are logged but do not roll back the save.
type CreateAnnouncementHandler struct {
	repo      announcerepo.AnnouncementRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateAnnouncementHandler wires dependencies for announcement creation.
func NewCreateAnnouncementHandler(
	repo announcerepo.AnnouncementRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateAnnouncementHandler {
	return &CreateAnnouncementHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists a new announcement and publishes its domain events.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreateAnnouncementHandler) Handle(ctx context.Context, cmd CreateAnnouncementCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateAnnouncementHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateAnnouncement", "announcement")()

	a, err := announceentity.NewAnnouncement(cmd.Title, cmd.Content, cmd.Priority, cmd.StartDate, cmd.EndDate)
	if err != nil {
		return fmt.Errorf("create_announcement: %w", err)
	}

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, a); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateAnnouncement", Entity: "announcement", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, a.Events)
}
