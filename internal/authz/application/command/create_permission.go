package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreatePermissionCommand represents an intent to register a new permission in the authorization system.
// ParentID enables hierarchical permission trees — nil means a root-level permission.
type CreatePermissionCommand struct {
	Name        string
	ParentID    *uuid.UUID
	Description *string
}

// CreatePermissionHandler persists new permissions via the repository.
// No domain events are emitted — permissions are structural metadata, not runtime state changes.
type CreatePermissionHandler struct {
	repo   domain.PermissionRepository
	logger logger.Log
}

// NewCreatePermissionHandler wires dependencies for permission creation.
func NewCreatePermissionHandler(
	repo domain.PermissionRepository,
	logger logger.Log,
) *CreatePermissionHandler {
	return &CreatePermissionHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle creates a permission, optionally sets its description, and persists it.
// Returns nil on success; propagates repository errors (e.g., duplicate name, invalid parent) to the caller.
func (h *CreatePermissionHandler) Handle(ctx context.Context, cmd CreatePermissionCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreatePermissionHandler.Handle")
	defer func() { end(err) }()

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
