package command

import (
	"context"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteSiteSettingCommand represents an intent to permanently remove a site setting.
// Once deleted, any feature relying on this setting will fall back to its default behavior.
type DeleteSiteSettingCommand struct {
	ID siteentity.SiteSettingID
}

// DeleteSiteSettingHandler performs hard-delete of site settings through the repository.
// No domain events are emitted — callers needing audit trails should handle that separately.
type DeleteSiteSettingHandler struct {
	repo   siterepo.SiteSettingRepository
	pool   shareddomain.Querier
	logger logger.Log
}

// NewDeleteSiteSettingHandler creates a new DeleteSiteSettingHandler.
func NewDeleteSiteSettingHandler(
	repo siterepo.SiteSettingRepository,
	pool shareddomain.Querier,
	logger logger.Log,
) *DeleteSiteSettingHandler {
	return &DeleteSiteSettingHandler{
		repo:   repo,
		pool:   pool,
		logger: logger,
	}
}

// Handle deletes the site setting identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteSiteSettingHandler) Handle(ctx context.Context, cmd DeleteSiteSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteSiteSettingHandler.Handle")
	defer func() { end(err) }()

	if err := h.repo.Delete(ctx, h.pool, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteSiteSetting", Entity: "site_setting", EntityID: cmd.ID.String(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}
	return nil
}
