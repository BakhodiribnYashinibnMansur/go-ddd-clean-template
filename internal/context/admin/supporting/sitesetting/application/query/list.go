package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/supporting/sitesetting/application/dto"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListSiteSettingsQuery holds the input for listing site settings.
type ListSiteSettingsQuery struct {
	Filter siterepo.SiteSettingFilter
}

// ListSiteSettingsResult holds the output of the list site settings query.
type ListSiteSettingsResult struct {
	Settings []*dto.SiteSettingView
	Total    int64
}

// ListSiteSettingsHandler handles the ListSiteSettingsQuery.
type ListSiteSettingsHandler struct {
	readRepo siterepo.SiteSettingReadRepository
	logger   logger.Log
}

// NewListSiteSettingsHandler creates a new ListSiteSettingsHandler.
func NewListSiteSettingsHandler(readRepo siterepo.SiteSettingReadRepository, l logger.Log) *ListSiteSettingsHandler {
	return &ListSiteSettingsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListSiteSettingsQuery and returns a list of SiteSettingView with total count.
func (h *ListSiteSettingsHandler) Handle(ctx context.Context, q ListSiteSettingsQuery) (result *ListSiteSettingsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListSiteSettingsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListSiteSettings", "site_setting")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListSiteSettings", Entity: "site_setting", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*dto.SiteSettingView, len(views))
	for i, v := range views {
		items[i] = &dto.SiteSettingView{
			ID:          uuid.UUID(v.ID),
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
