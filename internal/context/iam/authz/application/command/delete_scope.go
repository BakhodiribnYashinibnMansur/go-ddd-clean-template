package command

import (
	"context"

	"gct/internal/context/iam/authz/domain"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
)

// DeleteScopeCommand represents an intent to remove a protected API scope identified by its path and HTTP method.
// Deleting a scope may cascade to permission-scope assignments depending on FK constraints.
type DeleteScopeCommand struct {
	Path   string
	Method string
}

// DeleteScopeHandler performs hard deletion of an API scope via the repository.
// No domain events are emitted; callers needing downstream notifications should handle that at a higher layer.
type DeleteScopeHandler struct {
	repo   domain.ScopeRepository
	logger logger.Log
}

// NewDeleteScopeHandler wires dependencies for scope deletion.
func NewDeleteScopeHandler(
	repo domain.ScopeRepository,
	logger logger.Log,
) *DeleteScopeHandler {
	return &DeleteScopeHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the scope identified by the Path+Method composite key.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeleteScopeHandler) Handle(ctx context.Context, cmd DeleteScopeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteScopeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteScope", "scope")()

	if err := h.repo.Delete(ctx, cmd.Path, cmd.Method); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteScope", Entity: "scope", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
