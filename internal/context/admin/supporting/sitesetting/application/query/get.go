package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/admin/supporting/sitesetting/application"
	"gct/internal/context/admin/supporting/sitesetting/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetSiteSettingQuery holds the input for getting a single site setting.
type GetSiteSettingQuery struct {
	ID domain.SiteSettingID
}

// GetSiteSettingHandler handles the GetSiteSettingQuery.
type GetSiteSettingHandler struct {
	readRepo domain.SiteSettingReadRepository
	logger   logger.Log
}

// NewGetSiteSettingHandler creates a new GetSiteSettingHandler.
func NewGetSiteSettingHandler(readRepo domain.SiteSettingReadRepository, l logger.Log) *GetSiteSettingHandler {
	return &GetSiteSettingHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetSiteSettingQuery and returns a SiteSettingView.
func (h *GetSiteSettingHandler) Handle(ctx context.Context, q GetSiteSettingQuery) (result *appdto.SiteSettingView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSiteSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetSiteSetting", "site_setting")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetSiteSetting", Entity: "site_setting", EntityID: q.ID.String(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.SiteSettingView{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Type:        v.Type,
		Description: v.Description,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}, nil
}
