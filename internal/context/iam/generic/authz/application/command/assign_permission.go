package command

import (
	"context"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	authzevent "gct/internal/context/iam/generic/authz/domain/event"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// AssignPermissionCommand represents an intent to grant a permission to a role in the RBAC hierarchy.
// This creates a role-permission binding; all users holding this role gain the permission immediately.
type AssignPermissionCommand struct {
	RoleID       authzentity.RoleID
	PermissionID authzentity.PermissionID
}

// AssignPermissionHandler binds a permission to a role and emits a PermissionGranted event.
// The event enables downstream consumers (e.g., cache invalidation, real-time authorization updates) to react.
// Event publish failures are logged but do not roll back the assignment.
type AssignPermissionHandler struct {
	rolePermRepo authzrepo.RolePermissionRepository
	eventBus     application.EventBus
	logger       logger.Log
}

// NewAssignPermissionHandler wires dependencies for permission assignment.
func NewAssignPermissionHandler(
	rolePermRepo authzrepo.RolePermissionRepository,
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
func (h *AssignPermissionHandler) Handle(ctx context.Context, cmd AssignPermissionCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "AssignPermissionHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "AssignPermission", "role")()

	if err := h.rolePermRepo.Assign(ctx, cmd.RoleID, cmd.PermissionID); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "AssignPermission", Entity: "role", EntityID: cmd.RoleID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	event := authzevent.NewPermissionGranted(cmd.RoleID.UUID(), cmd.PermissionID.UUID())
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "AssignPermission", Entity: "role", EntityID: cmd.RoleID, Err: err}.KV()...)
	}

	return nil
}
