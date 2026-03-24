package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// AssignScopeCommand holds the input for assigning a scope to a permission.
type AssignScopeCommand struct {
	PermissionID uuid.UUID
	Path         string
	Method       string
}

// AssignScopeHandler handles the AssignScopeCommand.
type AssignScopeHandler struct {
	permScopeRepo domain.PermissionScopeRepository
	logger        logger.Log
}

// NewAssignScopeHandler creates a new AssignScopeHandler.
func NewAssignScopeHandler(
	permScopeRepo domain.PermissionScopeRepository,
	logger logger.Log,
) *AssignScopeHandler {
	return &AssignScopeHandler{
		permScopeRepo: permScopeRepo,
		logger:        logger,
	}
}

// Handle executes the AssignScopeCommand.
func (h *AssignScopeHandler) Handle(ctx context.Context, cmd AssignScopeCommand) error {
	if err := h.permScopeRepo.Assign(ctx, cmd.PermissionID, cmd.Path, cmd.Method); err != nil {
		h.logger.Errorf("failed to assign scope to permission: %v", err)
		return err
	}

	return nil
}
