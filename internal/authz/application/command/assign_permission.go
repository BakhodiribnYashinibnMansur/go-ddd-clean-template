package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// AssignPermissionCommand holds the input for assigning a permission to a role.
type AssignPermissionCommand struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}

// AssignPermissionHandler handles the AssignPermissionCommand.
type AssignPermissionHandler struct {
	rolePermRepo domain.RolePermissionRepository
	eventBus     application.EventBus
	logger       logger.Log
}

// NewAssignPermissionHandler creates a new AssignPermissionHandler.
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

// Handle executes the AssignPermissionCommand.
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
