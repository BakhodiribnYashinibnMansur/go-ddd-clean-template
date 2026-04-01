package command

import (
	"context"

	"gct/internal/errorcode/domain"
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
// No domain events are emitted — deletion is a silent removal.
type DeleteErrorCodeHandler struct {
	repo   domain.ErrorCodeRepository
	logger logger.Log
}

// NewDeleteErrorCodeHandler wires dependencies for error code deletion.
func NewDeleteErrorCodeHandler(
	repo domain.ErrorCodeRepository,
	logger logger.Log,
) *DeleteErrorCodeHandler {
	return &DeleteErrorCodeHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the error code identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeleteErrorCodeHandler) Handle(ctx context.Context, cmd DeleteErrorCodeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteErrorCodeHandler.Handle")
	defer func() { end(err) }()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete error code: %v", err)
		return err
	}
	return nil
}
