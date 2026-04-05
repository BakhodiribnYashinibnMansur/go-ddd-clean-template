package command

import (
	"context"

	"gct/internal/context/iam/generic/authz/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreatePermissionCommand represents an intent to register a new permission in the authorization system.
// ParentID enables hierarchical permission trees — nil means a root-level permission.
type CreatePermissionCommand struct {
	Name        string
	ParentID    *domain.PermissionID
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
	defer logger.SlowOp(h.logger, ctx, "CreatePermission", "permission")()

	var parentUUID *uuid.UUID
	if cmd.ParentID != nil {
		u := cmd.ParentID.UUID()
		parentUUID = &u
	}
	perm := domain.NewPermission(cmd.Name, parentUUID)
	if cmd.Description != nil {
		perm.SetDescription(cmd.Description)
	}

	if err := h.repo.Save(ctx, perm); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreatePermission", Entity: "permission", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
