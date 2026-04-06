package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/supporting/sitesetting/application/dto"
	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetSiteSettingQuery holds the input for getting a single site setting.
type GetSiteSettingQuery struct {
	ID siteentity.SiteSettingID
}

// GetSiteSettingHandler handles the GetSiteSettingQuery.
type GetSiteSettingHandler struct {
	readRepo siterepo.SiteSettingReadRepository
	logger   logger.Log
}

// NewGetSiteSettingHandler creates a new GetSiteSettingHandler.
func NewGetSiteSettingHandler(readRepo siterepo.SiteSettingReadRepository, l logger.Log) *GetSiteSettingHandler {
	return &GetSiteSettingHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetSiteSettingQuery and returns a SiteSettingView.
func (h *GetSiteSettingHandler) Handle(ctx context.Context, q GetSiteSettingQuery) (result *dto.SiteSettingView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSiteSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetSiteSetting", "site_setting")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetSiteSetting", Entity: "site_setting", EntityID: q.ID.String(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.SiteSettingView{
		ID:          uuid.UUID(v.ID),
		Key:         v.Key,
		Value:       v.Value,
		Type:        v.Type,
		Description: v.Description,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}, nil
}
