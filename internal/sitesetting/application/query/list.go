package query

import (
	"context"

	"gct/internal/shared/infrastructure/pgxutil"
	appdto "gct/internal/sitesetting/application"
	"gct/internal/sitesetting/domain"
)

// ListSiteSettingsQuery holds the input for listing site settings.
type ListSiteSettingsQuery struct {
	Filter domain.SiteSettingFilter
}

// ListSiteSettingsResult holds the output of the list site settings query.
type ListSiteSettingsResult struct {
	Settings []*appdto.SiteSettingView
	Total    int64
}

// ListSiteSettingsHandler handles the ListSiteSettingsQuery.
type ListSiteSettingsHandler struct {
	readRepo domain.SiteSettingReadRepository
}

// NewListSiteSettingsHandler creates a new ListSiteSettingsHandler.
func NewListSiteSettingsHandler(readRepo domain.SiteSettingReadRepository) *ListSiteSettingsHandler {
	return &ListSiteSettingsHandler{readRepo: readRepo}
}

// Handle executes the ListSiteSettingsQuery and returns a list of SiteSettingView with total count.
func (h *ListSiteSettingsHandler) Handle(ctx context.Context, q ListSiteSettingsQuery) (result *ListSiteSettingsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListSiteSettingsHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	items := make([]*appdto.SiteSettingView, len(views))
	for i, v := range views {
		items[i] = &appdto.SiteSettingView{
			ID:          v.ID,
			Key:         v.Key,
			Value:       v.Value,
			Type:        v.Type,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
	}

	return &ListSiteSettingsResult{
		Settings: items,
		Total:    total,
	}, nil
}
