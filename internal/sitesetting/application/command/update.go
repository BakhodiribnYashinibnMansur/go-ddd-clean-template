package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/sitesetting/domain"

	"github.com/google/uuid"
)

// UpdateSiteSettingCommand holds the input for updating a site setting.
type UpdateSiteSettingCommand struct {
	ID          uuid.UUID
	Key         *string
	Value       *string
	Type        *string
	Description *string
}

// UpdateSiteSettingHandler handles the UpdateSiteSettingCommand.
type UpdateSiteSettingHandler struct {
	repo     domain.SiteSettingRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateSiteSettingHandler creates a new UpdateSiteSettingHandler.
func NewUpdateSiteSettingHandler(
	repo domain.SiteSettingRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateSiteSettingHandler {
	return &UpdateSiteSettingHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateSiteSettingCommand.
func (h *UpdateSiteSettingHandler) Handle(ctx context.Context, cmd UpdateSiteSettingCommand) error {
	s, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	s.Update(cmd.Key, cmd.Value, cmd.Type, cmd.Description)

	if err := h.repo.Update(ctx, s); err != nil {
		h.logger.Errorf("failed to update site setting: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, s.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
