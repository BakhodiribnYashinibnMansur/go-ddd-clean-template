package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreatePermissionCommand holds the input for creating a new permission.
type CreatePermissionCommand struct {
	Name        string
	ParentID    *uuid.UUID
	Description *string
}

// CreatePermissionHandler handles the CreatePermissionCommand.
type CreatePermissionHandler struct {
	repo   domain.PermissionRepository
	logger logger.Log
}

// NewCreatePermissionHandler creates a new CreatePermissionHandler.
func NewCreatePermissionHandler(
	repo domain.PermissionRepository,
	logger logger.Log,
) *CreatePermissionHandler {
	return &CreatePermissionHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the CreatePermissionCommand.
func (h *CreatePermissionHandler) Handle(ctx context.Context, cmd CreatePermissionCommand) error {
	perm := domain.NewPermission(cmd.Name, cmd.ParentID)
	if cmd.Description != nil {
		perm.SetDescription(cmd.Description)
	}

	if err := h.repo.Save(ctx, perm); err != nil {
		h.logger.Errorf("failed to save permission: %v", err)
		return err
	}

	return nil
}
