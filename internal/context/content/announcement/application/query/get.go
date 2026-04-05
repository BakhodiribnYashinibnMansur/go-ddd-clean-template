package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/content/announcement/application"
	"gct/internal/context/content/announcement/domain"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetAnnouncementQuery holds the input for getting a single announcement.
type GetAnnouncementQuery struct {
	ID uuid.UUID
}

// GetAnnouncementHandler handles the GetAnnouncementQuery.
type GetAnnouncementHandler struct {
	readRepo domain.AnnouncementReadRepository
	logger   logger.Log
}

// NewGetAnnouncementHandler creates a new GetAnnouncementHandler.
func NewGetAnnouncementHandler(readRepo domain.AnnouncementReadRepository, l logger.Log) *GetAnnouncementHandler {
	return &GetAnnouncementHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetAnnouncementQuery and returns an AnnouncementView.
func (h *GetAnnouncementHandler) Handle(ctx context.Context, q GetAnnouncementQuery) (result *appdto.AnnouncementView, err error) {
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
