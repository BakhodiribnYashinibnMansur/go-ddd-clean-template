package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteRoleCommand holds the input for deleting a role.
type DeleteRoleCommand struct {
	ID uuid.UUID
}

// DeleteRoleHandler handles the DeleteRoleCommand.
type DeleteRoleHandler struct {
	repo     domain.RoleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteRoleHandler creates a new DeleteRoleHandler.
func NewDeleteRoleHandler(
	repo domain.RoleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteRoleHandler {
	return &DeleteRoleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the DeleteRoleCommand.
func (h *DeleteRoleHandler) Handle(ctx context.Context, cmd DeleteRoleCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete role: %v", err)
		return err
	}

	event := domain.NewRoleDeleted(cmd.ID)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
