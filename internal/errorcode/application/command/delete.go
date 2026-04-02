package command

import (
	"context"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// DeleteErrorCodeCommand represents an intent to permanently remove a standardized error code.
// Callers should ensure no API handlers still reference this code before deletion to avoid runtime lookup failures.
type DeleteErrorCodeCommand struct {
	ID uuid.UUID
}

// DeleteErrorCodeHandler performs hard deletion of an error code via the repository.
// Publishes an ErrorCodeDeleted event so subscribers can evict the code from in-memory caches.
type DeleteErrorCodeHandler struct {
	repo     domain.ErrorCodeRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteErrorCodeHandler wires dependencies for error code deletion.
func NewDeleteErrorCodeHandler(
	repo domain.ErrorCodeRepository,
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

	ec, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if err = h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete error code: %v", err)
		return err
	}

	event := domain.NewErrorCodeDeleted(cmd.ID, ec.Code())
	if pubErr := h.eventBus.Publish(ctx, event); pubErr != nil {
		h.logger.Errorf("failed to publish error code deleted event: %v", pubErr)
	}

	return nil
}
