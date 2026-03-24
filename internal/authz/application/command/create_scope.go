package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"
)

// CreateScopeCommand holds the input for creating a new scope.
type CreateScopeCommand struct {
	Path   string
	Method string
}

// CreateScopeHandler handles the CreateScopeCommand.
type CreateScopeHandler struct {
	repo   domain.ScopeRepository
	logger logger.Log
}

// NewCreateScopeHandler creates a new CreateScopeHandler.
func NewCreateScopeHandler(
	repo domain.ScopeRepository,
	logger logger.Log,
) *CreateScopeHandler {
	return &CreateScopeHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the CreateScopeCommand.
func (h *CreateScopeHandler) Handle(ctx context.Context, cmd CreateScopeCommand) error {
	scope := domain.Scope{
		Path:   cmd.Path,
		Method: cmd.Method,
	}

	if err := h.repo.Save(ctx, scope); err != nil {
		h.logger.Errorf("failed to save scope: %v", err)
		return err
	}

	return nil
}
