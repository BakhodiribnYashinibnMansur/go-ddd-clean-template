package command

import (
	"context"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteErrorCodeCommand holds the input for deleting an error code.
type DeleteErrorCodeCommand struct {
	ID uuid.UUID
}

// DeleteErrorCodeHandler handles the DeleteErrorCodeCommand.
type DeleteErrorCodeHandler struct {
	repo   domain.ErrorCodeRepository
	logger logger.Log
}

// NewDeleteErrorCodeHandler creates a new DeleteErrorCodeHandler.
func NewDeleteErrorCodeHandler(
	repo domain.ErrorCodeRepository,
	logger logger.Log,
) *DeleteErrorCodeHandler {
	return &DeleteErrorCodeHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteErrorCodeCommand.
func (h *DeleteErrorCodeHandler) Handle(ctx context.Context, cmd DeleteErrorCodeCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete error code: %v", err)
		return err
	}
	return nil
}
