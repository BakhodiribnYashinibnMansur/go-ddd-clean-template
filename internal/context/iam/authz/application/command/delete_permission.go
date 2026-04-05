package command

import (
	"context"

	"gct/internal/context/iam/authz/domain"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// DeletePermissionCommand represents an intent to permanently remove a permission.
// Deleting a permission may cascade to role-permission assignments depending on the repository's FK constraints.
type DeletePermissionCommand struct {
	ID uuid.UUID
}

// DeletePermissionHandler performs hard deletion of a permission via the repository.
// No domain events are emitted; callers needing cascade notifications should handle that at a higher layer.
type DeletePermissionHandler struct {
	repo   domain.PermissionRepository
	logger logger.Log
}

// NewDeletePermissionHandler wires dependencies for permission deletion.
func NewDeletePermissionHandler(
	repo domain.PermissionRepository,
	logger logger.Log,
) *DeletePermissionHandler {
	return &DeletePermissionHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the permission identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, FK violation) to the caller.
func (h *DeletePermissionHandler) Handle(ctx context.Context, cmd DeletePermissionCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeletePermissionHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeletePermission", "permission")()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeletePermission", Entity: "permission", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
