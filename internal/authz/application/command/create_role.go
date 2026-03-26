package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateRoleCommand represents an intent to create a new authorization role.
// Roles are the top-level grouping in the RBAC hierarchy; permissions are assigned to roles separately.
type CreateRoleCommand struct {
	Name        string
	Description *string
}

// CreateRoleHandler orchestrates role creation and emits domain events for downstream authorization cache invalidation.
// Event publish failures are logged but do not roll back the persisted role.
type CreateRoleHandler struct {
	repo     domain.RoleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateRoleHandler wires dependencies for role creation.
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

// Handle creates a new role, optionally sets its description, persists it, and publishes domain events.
// Returns nil on success; propagates repository errors (e.g., duplicate name) to the caller.
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
