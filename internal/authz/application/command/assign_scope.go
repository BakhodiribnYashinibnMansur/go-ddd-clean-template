package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// AssignScopeCommand represents an intent to bind an API scope (path + method) to a permission.
// This mapping determines which API endpoints a permission protects during authorization evaluation.
type AssignScopeCommand struct {
	PermissionID uuid.UUID
	Path         string
	Method       string
}

// AssignScopeHandler binds a scope to a permission via the permission-scope repository.
// No domain events are emitted — scope assignments are structural and take effect on the next authorization check.
type AssignScopeHandler struct {
	permScopeRepo domain.PermissionScopeRepository
	logger        logger.Log
}

// NewAssignScopeHandler wires dependencies for scope assignment.
func NewAssignScopeHandler(
	permScopeRepo domain.PermissionScopeRepository,
	logger logger.Log,
) *AssignScopeHandler {
	return &AssignScopeHandler{
		permScopeRepo: permScopeRepo,
		logger:        logger,
	}
}

// Handle creates the permission-scope binding for the given permission ID, path, and method.
// Returns nil on success; propagates repository errors (e.g., duplicate binding, invalid FK) to the caller.
func (h *AssignScopeHandler) Handle(ctx context.Context, cmd AssignScopeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "AssignScopeHandler.Handle")
	defer func() { end(err) }()

	if err := h.permScopeRepo.Assign(ctx, cmd.PermissionID, cmd.Path, cmd.Method); err != nil {
		h.logger.Errorf("failed to assign scope to permission: %v", err)
		return err
	}

	return nil
}
