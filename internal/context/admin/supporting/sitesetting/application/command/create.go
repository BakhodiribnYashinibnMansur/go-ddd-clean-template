package command

import (
	"context"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// CreateSiteSettingCommand represents an intent to register a new site-wide configuration entry.
// Key must be unique across all settings; the repository will reject duplicates.
type CreateSiteSettingCommand struct {
	Key         string
	Value       string
	Type        string
	Description string
}

// CreateSiteSettingHandler orchestrates site setting creation through the repository layer.
type CreateSiteSettingHandler struct {
	repo      siterepo.SiteSettingRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateSiteSettingHandler creates a new CreateSiteSettingHandler.
func NewCreateSiteSettingHandler(
	repo siterepo.SiteSettingRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateSiteSettingHandler {
	return &CreateSiteSettingHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists a new site setting and publishes resulting domain events.
// Returns repository errors (e.g., duplicate key, connection failure) directly to the caller.
func (h *CreateSiteSettingHandler) Handle(ctx context.Context, cmd CreateSiteSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateSiteSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateSiteSetting", "site_setting")()

	s := siteentity.NewSiteSetting(cmd.Key, cmd.Value, cmd.Type, cmd.Description)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, s); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateSiteSetting", Entity: "site_setting", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, s.Events)
}
