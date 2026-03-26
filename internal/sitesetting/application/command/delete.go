package command

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/sitesetting/domain"

	"github.com/google/uuid"
)

// DeleteSiteSettingCommand represents an intent to permanently remove a site setting.
// Once deleted, any feature relying on this setting will fall back to its default behavior.
type DeleteSiteSettingCommand struct {
	ID uuid.UUID
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
func (h *DeleteSiteSettingHandler) Handle(ctx context.Context, cmd DeleteSiteSettingCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete site setting: %v", err)
		return err
	}
	return nil
}
