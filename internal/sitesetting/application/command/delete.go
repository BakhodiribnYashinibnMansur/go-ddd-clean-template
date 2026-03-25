package command

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/sitesetting/domain"

	"github.com/google/uuid"
)

// DeleteSiteSettingCommand holds the input for deleting a site setting.
type DeleteSiteSettingCommand struct {
	ID uuid.UUID
}

// DeleteSiteSettingHandler handles the DeleteSiteSettingCommand.
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

// Handle executes the DeleteSiteSettingCommand.
func (h *DeleteSiteSettingHandler) Handle(ctx context.Context, cmd DeleteSiteSettingCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete site setting: %v", err)
		return err
	}
	return nil
}
