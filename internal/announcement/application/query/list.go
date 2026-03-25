package query

import (
	"context"

	appdto "gct/internal/announcement/application"
	"gct/internal/announcement/domain"
	shared "gct/internal/shared/domain"
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
}

// NewListAnnouncementsHandler creates a new ListAnnouncementsHandler.
func NewListAnnouncementsHandler(readRepo domain.AnnouncementReadRepository) *ListAnnouncementsHandler {
	return &ListAnnouncementsHandler{readRepo: readRepo}
}

// Handle executes the ListAnnouncementsQuery and returns a list of AnnouncementView with total count.
func (h *ListAnnouncementsHandler) Handle(ctx context.Context, q ListAnnouncementsQuery) (*ListAnnouncementsResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.AnnouncementView, len(views))
	for i, v := range views {
		result[i] = toAppView(v)
	}

	return &ListAnnouncementsResult{
		Announcements: result,
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
