package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/sitesetting/domain"
)

// CreateSiteSettingCommand holds the input for creating a new site setting.
type CreateSiteSettingCommand struct {
	Key         string
	Value       string
	Type        string
	Description string
}

// CreateSiteSettingHandler handles the CreateSiteSettingCommand.
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

// Handle executes the CreateSiteSettingCommand.
func (h *CreateSiteSettingHandler) Handle(ctx context.Context, cmd CreateSiteSettingCommand) error {
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
