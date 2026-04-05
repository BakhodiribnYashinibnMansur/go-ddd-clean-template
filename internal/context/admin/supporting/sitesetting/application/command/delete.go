package command

import (
	"context"

	"gct/internal/context/admin/supporting/sitesetting/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteSiteSettingCommand represents an intent to permanently remove a site setting.
// Once deleted, any feature relying on this setting will fall back to its default behavior.
type DeleteSiteSettingCommand struct {
	ID domain.SiteSettingID
}

// DeleteSiteSettingHandler performs hard-delete of site settings through the repository.
// No domain events are emitted — callers needing audit trails should handle that separately.
type DeleteSiteSettingHandler struct {
	repo   domain.SiteSettingRepository
	logger logger.Log
}

// NewDeleteSiteSettingHandler creates a new DeleteSiteSettingHandler.
func NewDeleteSiteSettingHandler(
	repo domain.SiteSettingRepository,
	logger logger.Log,
) *DeleteSiteSettingHandler {
	return &DeleteSiteSettingHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the site setting identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteSiteSettingHandler) Handle(ctx context.Context, cmd DeleteSiteSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteSiteSettingHandler.Handle")
	defer func() { end(err) }()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteSiteSetting", Entity: "site_setting", EntityID: cmd.ID.String(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}
	return nil
}
