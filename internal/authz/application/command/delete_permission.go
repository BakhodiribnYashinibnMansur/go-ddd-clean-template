package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeletePermissionCommand holds the input for deleting a permission.
type DeletePermissionCommand struct {
	ID uuid.UUID
}

// DeletePermissionHandler handles the DeletePermissionCommand.
type DeletePermissionHandler struct {
	repo   domain.PermissionRepository
	logger logger.Log
}

// NewDeletePermissionHandler creates a new DeletePermissionHandler.
func NewDeletePermissionHandler(
	repo domain.PermissionRepository,
	logger logger.Log,
) *DeletePermissionHandler {
	return &DeletePermissionHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeletePermissionCommand.
func (h *DeletePermissionHandler) Handle(ctx context.Context, cmd DeletePermissionCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete permission: %v", err)
		return err
	}

	return nil
}
