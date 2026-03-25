package command

import (
	"context"

	"gct/internal/integration/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateCommand holds the input for updating an integration.
type UpdateCommand struct {
	ID         uuid.UUID
	Name       *string
	Type       *string
	APIKey     *string
	WebhookURL *string
	Enabled    *bool
	Config     *map[string]any
}

// UpdateHandler handles the UpdateCommand.
type UpdateHandler struct {
	repo     domain.IntegrationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler creates a new UpdateHandler.
func NewUpdateHandler(
	repo domain.IntegrationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateHandler {
	return &UpdateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateCommand.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) error {
	i, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	i.UpdateDetails(cmd.Name, cmd.Type, cmd.APIKey, cmd.WebhookURL, cmd.Enabled, cmd.Config)

	if err := h.repo.Update(ctx, i); err != nil {
		h.logger.Errorf("failed to update integration: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, i.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
