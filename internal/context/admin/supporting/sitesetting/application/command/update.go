package command

import (
	"context"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// UpdateSiteSettingCommand represents a partial update to an existing site setting.
// Pointer fields use nil-means-unchanged semantics, so callers only populate the fields they want to modify.
type UpdateSiteSettingCommand struct {
	ID          siteentity.SiteSettingID
	Key         *string
	Value       *string
	Type        *string
	Description *string
}

// UpdateSiteSettingHandler applies partial updates to site settings via a load-modify-save cycle.
type UpdateSiteSettingHandler struct {
	repo      siterepo.SiteSettingRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewUpdateSiteSettingHandler creates a new UpdateSiteSettingHandler.
func NewUpdateSiteSettingHandler(
	repo siterepo.SiteSettingRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *UpdateSiteSettingHandler {
	return &UpdateSiteSettingHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle loads the setting by ID, applies the partial update, and persists the result.
// Returns not-found or repository errors to the caller; authorization is the caller's responsibility.
func (h *UpdateSiteSettingHandler) Handle(ctx context.Context, cmd UpdateSiteSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateSiteSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateSiteSetting", "site_setting")()

	s, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	s.Update(cmd.Key, cmd.Value, cmd.Type, cmd.Description)

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, s); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateSiteSetting", Entity: "site_setting", EntityID: cmd.ID.String(), Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, s.Events)
}
