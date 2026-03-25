package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/webhook/domain"
)

// CreateCommand holds the input for creating a new webhook.
type CreateCommand struct {
	Name    string
	URL     string
	Secret  string
	Events  []string
	Enabled bool
}

// CreateHandler handles the CreateCommand.
type CreateHandler struct {
	repo     domain.WebhookRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler creates a new CreateHandler.
func NewCreateHandler(
	repo domain.WebhookRepository,
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
	w := domain.NewWebhook(cmd.Name, cmd.URL, cmd.Secret, cmd.Events, cmd.Enabled)

	if err := h.repo.Save(ctx, w); err != nil {
		h.logger.Errorf("failed to save webhook: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, w.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
