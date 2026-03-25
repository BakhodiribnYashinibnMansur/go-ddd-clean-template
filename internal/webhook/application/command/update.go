package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/webhook/domain"

	"github.com/google/uuid"
)

// UpdateCommand holds the input for updating a webhook.
type UpdateCommand struct {
	ID      uuid.UUID
	Name    *string
	URL     *string
	Secret  *string
	Events  []string
	Enabled *bool
}

// UpdateHandler handles the UpdateCommand.
type UpdateHandler struct {
	repo     domain.WebhookRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler creates a new UpdateHandler.
func NewUpdateHandler(
	repo domain.WebhookRepository,
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
	w, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	w.UpdateDetails(cmd.Name, cmd.URL, cmd.Secret, cmd.Events, cmd.Enabled)

	if err := h.repo.Update(ctx, w); err != nil {
		h.logger.Errorf("failed to update webhook: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, w.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
