package command

import (
	"context"

	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/systemerror/domain"

	"github.com/google/uuid"
)

// ResolveErrorCommand represents an intent to mark a system error as resolved by a specific user.
// This is an irreversible status transition — once resolved, the error cannot be re-opened.
type ResolveErrorCommand struct {
	ID         uuid.UUID
	ResolvedBy uuid.UUID
}

// ResolveErrorHandler transitions a system error to the resolved state via a load-modify-save cycle.
// Callers are responsible for verifying that ResolvedBy refers to a user with sufficient privileges.
type ResolveErrorHandler struct {
	repo     domain.SystemErrorRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewResolveErrorHandler creates a new ResolveErrorHandler.
func NewResolveErrorHandler(
	repo domain.SystemErrorRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *ResolveErrorHandler {
	return &ResolveErrorHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle loads the system error, marks it resolved, persists the update, and publishes domain events.
// Returns not-found or repository errors; event bus failures are logged but do not fail the operation.
func (h *ResolveErrorHandler) Handle(ctx context.Context, cmd ResolveErrorCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ResolveErrorHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ResolveError", "system_error")()

	se, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	se.Resolve(cmd.ResolvedBy)

	if err := h.repo.Update(ctx, se); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "ResolveError", Entity: "system_error", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, se.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "ResolveError", Entity: "system_error", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
