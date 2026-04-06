package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/content/supporting/announcement/application/dto"
	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetAnnouncementQuery holds the input for getting a single announcement.
type GetAnnouncementQuery struct {
	ID announceentity.AnnouncementID
}

// GetAnnouncementHandler handles the GetAnnouncementQuery.
type GetAnnouncementHandler struct {
	readRepo announcerepo.AnnouncementReadRepository
	logger   logger.Log
}

// NewGetAnnouncementHandler creates a new GetAnnouncementHandler.
func NewGetAnnouncementHandler(readRepo announcerepo.AnnouncementReadRepository, l logger.Log) *GetAnnouncementHandler {
	return &GetAnnouncementHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetAnnouncementQuery and returns an AnnouncementView.
func (h *GetAnnouncementHandler) Handle(ctx context.Context, q GetAnnouncementQuery) (result *dto.AnnouncementView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetAnnouncementHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetAnnouncement", "announcement")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetAnnouncement", Entity: "announcement", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return toAppView(v), nil
}
