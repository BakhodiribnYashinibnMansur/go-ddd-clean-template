package command

import (
	"context"

	"gct/internal/context/admin/integration/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteCommand represents an intent to permanently remove an integration by its unique identifier.
// Once deleted, any webhooks or API keys associated with this integration become inoperative.
type DeleteCommand struct {
	ID domain.IntegrationID
}

// DeleteHandler orchestrates integration deletion through the repository layer.
// It enforces a hard-delete strategy — no soft-delete or event emission is performed.
// Callers are responsible for authorization checks before invoking this handler.
type DeleteHandler struct {
	repo     domain.IntegrationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler wires up the handler with its required dependencies.
func NewDeleteHandler(
	repo domain.IntegrationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteHandler {
	return &DeleteHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle performs the deletion of the integration identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteIntegration", "integration")()

	if err := h.repo.Delete(ctx, cmd.ID.UUID()); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteIntegration", Entity: "integration", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
