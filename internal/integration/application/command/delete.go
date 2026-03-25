package command

import (
	"context"

	"gct/internal/integration/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteCommand holds the input for deleting an integration.
type DeleteCommand struct {
	ID uuid.UUID
}

// DeleteHandler handles the DeleteCommand.
type DeleteHandler struct {
	repo     domain.IntegrationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler creates a new DeleteHandler.
func NewDeleteHandler(
	repo domain.IntegrationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteHandler {
	return &DeleteHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the DeleteCommand.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete integration: %v", err)
		return err
	}

	return nil
}
