package command

import (
	"context"

	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/context/admin/sitesetting/domain"
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
// Domain events are published after a successful save; event bus failures are logged but do not roll back the write.
type CreateSiteSettingHandler struct {
	repo     domain.SiteSettingRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateSiteSettingHandler creates a new CreateSiteSettingHandler.
func NewCreateSiteSettingHandler(
	repo domain.SiteSettingRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateSiteSettingHandler {
	return &CreateSiteSettingHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle persists a new site setting and publishes resulting domain events.
// Returns repository errors (e.g., duplicate key, connection failure) directly to the caller.
func (h *CreateSiteSettingHandler) Handle(ctx context.Context, cmd CreateSiteSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateSiteSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateSiteSetting", "site_setting")()

	s := domain.NewSiteSetting(cmd.Key, cmd.Value, cmd.Type, cmd.Description)

	if err := h.repo.Save(ctx, s); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateSiteSetting", Entity: "site_setting", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, s.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateSiteSetting", Entity: "site_setting", Err: err}.KV()...)
	}

	return nil
}
