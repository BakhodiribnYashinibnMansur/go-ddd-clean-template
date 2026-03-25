package command

import (
	"context"

	"gct/internal/integration/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateCommand holds the input for creating a new integration.
type CreateCommand struct {
	Name       string
	Type       string
	APIKey     string
	WebhookURL string
	Enabled    bool
	Config     map[string]any
}

// CreateHandler handles the CreateCommand.
type CreateHandler struct {
	repo     domain.IntegrationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler creates a new CreateHandler.
func NewCreateHandler(
	repo domain.IntegrationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateCommand.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) error {
	i := domain.NewIntegration(cmd.Name, cmd.Type, cmd.APIKey, cmd.WebhookURL, cmd.Enabled, cmd.Config)

	if err := h.repo.Save(ctx, i); err != nil {
		h.logger.Errorf("failed to save integration: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, i.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
