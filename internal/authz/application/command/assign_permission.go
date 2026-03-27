package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// AssignPermissionCommand represents an intent to grant a permission to a role in the RBAC hierarchy.
// This creates a role-permission binding; all users holding this role gain the permission immediately.
type AssignPermissionCommand struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}

// AssignPermissionHandler binds a permission to a role and emits a PermissionGranted event.
// The event enables downstream consumers (e.g., cache invalidation, real-time authorization updates) to react.
// Event publish failures are logged but do not roll back the assignment.
type AssignPermissionHandler struct {
	rolePermRepo domain.RolePermissionRepository
	eventBus     application.EventBus
	logger       logger.Log
}

// NewAssignPermissionHandler wires dependencies for permission assignment.
func NewAssignPermissionHandler(
	rolePermRepo domain.RolePermissionRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *AssignPermissionHandler {
	return &AssignPermissionHandler{
		rolePermRepo: rolePermRepo,
		eventBus:     eventBus,
		logger:       logger,
	}
}

// Handle creates the role-permission binding and publishes a PermissionGranted domain event.
// Returns nil on success; propagates repository errors (e.g., duplicate assignment, invalid FK) to the caller.
func (h *AssignPermissionHandler) Handle(ctx context.Context, cmd AssignPermissionCommand) error {
	if err := h.rolePermRepo.Assign(ctx, cmd.RoleID, cmd.PermissionID); err != nil {
		h.logger.Errorf("failed to assign permission to role: %v", err)
		return err
	}

	event := domain.NewPermissionGranted(cmd.RoleID, cmd.PermissionID)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
