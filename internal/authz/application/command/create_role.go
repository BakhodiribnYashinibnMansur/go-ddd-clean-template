package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateRoleCommand holds the input for creating a new role.
type CreateRoleCommand struct {
	Name        string
	Description *string
}

// CreateRoleHandler handles the CreateRoleCommand.
type CreateRoleHandler struct {
	repo     domain.RoleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateRoleHandler creates a new CreateRoleHandler.
func NewCreateRoleHandler(
	repo domain.RoleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateRoleHandler {
	return &CreateRoleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateRoleCommand.
func (h *CreateRoleHandler) Handle(ctx context.Context, cmd CreateRoleCommand) error {
	role := domain.NewRole(cmd.Name)
	if cmd.Description != nil {
		role.SetDescription(cmd.Description)
	}

	if err := h.repo.Save(ctx, role); err != nil {
		h.logger.Errorf("failed to save role: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, role.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
