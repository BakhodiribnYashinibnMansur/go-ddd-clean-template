package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/content/supporting/announcement/application/dto"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListAnnouncementsQuery holds the input for listing announcements.
type ListAnnouncementsQuery struct {
	Filter announcerepo.AnnouncementFilter
}

// ListAnnouncementsResult holds the output of the list announcements query.
type ListAnnouncementsResult struct {
	Announcements []*dto.AnnouncementView
	Total         int64
}

// ListAnnouncementsHandler handles the ListAnnouncementsQuery.
type ListAnnouncementsHandler struct {
	readRepo announcerepo.AnnouncementReadRepository
	logger   logger.Log
}

// NewListAnnouncementsHandler creates a new ListAnnouncementsHandler.
func NewListAnnouncementsHandler(readRepo announcerepo.AnnouncementReadRepository, l logger.Log) *ListAnnouncementsHandler {
	return &ListAnnouncementsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListAnnouncementsQuery and returns a list of AnnouncementView with total count.
func (h *ListAnnouncementsHandler) Handle(ctx context.Context, q ListAnnouncementsQuery) (result *ListAnnouncementsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListAnnouncementsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListAnnouncements", "announcement")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListAnnouncements", Entity: "announcement", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*dto.AnnouncementView, len(views))
	for i, v := range views {
		items[i] = toAppView(v)
	}

	return &ListAnnouncementsResult{
		Announcements: items,
		Total:         total,
	}, nil
}

func toAppView(v *announcerepo.AnnouncementView) *dto.AnnouncementView {
	return &dto.AnnouncementView{
		ID:          uuid.UUID(v.ID),
		Title:       shared.Lang{Uz: v.TitleUz, Ru: v.TitleRu, En: v.TitleEn},
		Content:     shared.Lang{Uz: v.ContentUz, Ru: v.ContentRu, En: v.ContentEn},
		Published:   v.Published,
		PublishedAt: v.PublishedAt,
		Priority:    v.Priority,
		StartDate:   v.StartDate,
		EndDate:     v.EndDate,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
