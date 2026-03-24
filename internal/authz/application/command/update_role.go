package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateRoleCommand holds the input for updating an existing role.
type UpdateRoleCommand struct {
	ID          uuid.UUID
	Name        *string
	Description *string
}

// UpdateRoleHandler handles the UpdateRoleCommand.
type UpdateRoleHandler struct {
	repo     domain.RoleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateRoleHandler creates a new UpdateRoleHandler.
func NewUpdateRoleHandler(
	repo domain.RoleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateRoleHandler {
	return &UpdateRoleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateRoleCommand.
func (h *UpdateRoleHandler) Handle(ctx context.Context, cmd UpdateRoleCommand) error {
	role, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if cmd.Name != nil {
		role.Rename(*cmd.Name)
	}
	if cmd.Description != nil {
		role.SetDescription(cmd.Description)
	}

	if err := h.repo.Update(ctx, role); err != nil {
		h.logger.Errorf("failed to update role: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, role.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
