package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"
)

// DeleteScopeCommand holds the input for deleting a scope.
type DeleteScopeCommand struct {
	Path   string
	Method string
}

// DeleteScopeHandler handles the DeleteScopeCommand.
type DeleteScopeHandler struct {
	repo   domain.ScopeRepository
	logger logger.Log
}

// NewDeleteScopeHandler creates a new DeleteScopeHandler.
func NewDeleteScopeHandler(
	repo domain.ScopeRepository,
	logger logger.Log,
) *DeleteScopeHandler {
	return &DeleteScopeHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteScopeCommand.
func (h *DeleteScopeHandler) Handle(ctx context.Context, cmd DeleteScopeCommand) error {
	if err := h.repo.Delete(ctx, cmd.Path, cmd.Method); err != nil {
		h.logger.Errorf("failed to delete scope: %v", err)
		return err
	}

	return nil
}
