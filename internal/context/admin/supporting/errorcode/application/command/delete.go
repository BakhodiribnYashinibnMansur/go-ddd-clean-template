package command

import (
	"context"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcodeevent "gct/internal/context/admin/supporting/errorcode/domain/event"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteErrorCodeCommand represents an intent to permanently remove a standardized error code.
// Callers should ensure no API handlers still reference this code before deletion to avoid runtime lookup failures.
type DeleteErrorCodeCommand struct {
	ID errcodeentity.ErrorCodeID
}

// DeleteErrorCodeHandler performs hard deletion of an error code via the repository.
// Publishes an ErrorCodeDeleted event so subscribers can evict the code from in-memory caches.
type DeleteErrorCodeHandler struct {
	repo     errcoderepo.ErrorCodeRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteErrorCodeHandler wires dependencies for error code deletion.
func NewDeleteErrorCodeHandler(
	repo errcoderepo.ErrorCodeRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteErrorCodeHandler {
	return &DeleteErrorCodeHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the error code to capture its code string, deletes it, and publishes a deleted event.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeleteErrorCodeHandler) Handle(ctx context.Context, cmd DeleteErrorCodeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteErrorCodeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteErrorCode", "error_code")()

	ec, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err = h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteErrorCode", Entity: "error_code", EntityID: cmd.ID.String(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	event := errcodeevent.NewErrorCodeDeleted(cmd.ID.UUID(), ec.Code())
	if pubErr := h.eventBus.Publish(ctx, event); pubErr != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "DeleteErrorCode", Entity: "error_code", EntityID: cmd.ID.String(), Err: pubErr}.KV()...)
	}

	return nil
}
