package command

import (
	"context"

	"gct/internal/context/iam/generic/authz/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// AssignScopeCommand represents an intent to bind an API scope (path + method) to a permission.
// This mapping determines which API endpoints a permission protects during authorization evaluation.
type AssignScopeCommand struct {
	PermissionID domain.PermissionID
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
	defer logger.SlowOp(h.logger, ctx, "AssignScope", "role")()

	if err := h.permScopeRepo.Assign(ctx, cmd.PermissionID, cmd.Path, cmd.Method); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "AssignScope", Entity: "role", EntityID: cmd.PermissionID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
