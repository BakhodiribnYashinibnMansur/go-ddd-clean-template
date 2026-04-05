package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/content/supporting/announcement/application"
	"gct/internal/context/content/supporting/announcement/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListAnnouncementsQuery holds the input for listing announcements.
type ListAnnouncementsQuery struct {
	Filter domain.AnnouncementFilter
}

// ListAnnouncementsResult holds the output of the list announcements query.
type ListAnnouncementsResult struct {
	Announcements []*appdto.AnnouncementView
	Total         int64
}

// ListAnnouncementsHandler handles the ListAnnouncementsQuery.
type ListAnnouncementsHandler struct {
	readRepo domain.AnnouncementReadRepository
	logger   logger.Log
}

// NewListAnnouncementsHandler creates a new ListAnnouncementsHandler.
func NewListAnnouncementsHandler(readRepo domain.AnnouncementReadRepository, l logger.Log) *ListAnnouncementsHandler {
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

	items := make([]*appdto.AnnouncementView, len(views))
	for i, v := range views {
		items[i] = toAppView(v)
	}

	return &ListAnnouncementsResult{
		Announcements: items,
		Total:         total,
	}, nil
}

func toAppView(v *domain.AnnouncementView) *appdto.AnnouncementView {
	return &appdto.AnnouncementView{
		ID:          v.ID,
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
