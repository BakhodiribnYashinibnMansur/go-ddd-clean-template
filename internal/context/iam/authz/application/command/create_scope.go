package command

import (
	"context"

	"gct/internal/context/iam/authz/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// CreateScopeCommand represents an intent to register an API endpoint (path + HTTP method) as a protected scope.
// Scopes are the finest-grained authorization targets — permissions are mapped to scopes to control endpoint access.
type CreateScopeCommand struct {
	Path   string
	Method string
}

// CreateScopeHandler persists new API scopes via the repository.
// No domain events are emitted — scopes are structural metadata consumed during authorization evaluation.
type CreateScopeHandler struct {
	repo   domain.ScopeRepository
	logger logger.Log
}

// NewCreateScopeHandler wires dependencies for scope creation.
func NewCreateScopeHandler(
	repo domain.ScopeRepository,
	logger logger.Log,
) *CreateScopeHandler {
	return &CreateScopeHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle persists the scope defined by Path and Method.
// Returns nil on success; propagates repository errors (e.g., duplicate path+method pair) to the caller.
func (h *CreateScopeHandler) Handle(ctx context.Context, cmd CreateScopeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateScopeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateScope", "scope")()

	scope := domain.Scope{
		Path:   cmd.Path,
		Method: cmd.Method,
	}

	if err := h.repo.Save(ctx, scope); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateScope", Entity: "scope", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
