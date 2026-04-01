package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/sitesetting/domain"
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

	s := domain.NewSiteSetting(cmd.Key, cmd.Value, cmd.Type, cmd.Description)

	if err := h.repo.Save(ctx, s); err != nil {
		h.logger.Errorf("failed to save site setting: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, s.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
