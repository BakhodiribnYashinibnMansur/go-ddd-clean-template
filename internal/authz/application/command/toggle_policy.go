package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// TogglePolicyCommand holds the input for toggling a policy's active state.
type TogglePolicyCommand struct {
	ID uuid.UUID
}

// TogglePolicyHandler handles the TogglePolicyCommand.
type TogglePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewTogglePolicyHandler creates a new TogglePolicyHandler.
func NewTogglePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *TogglePolicyHandler {
	return &TogglePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the TogglePolicyCommand.
func (h *TogglePolicyHandler) Handle(ctx context.Context, cmd TogglePolicyCommand) error {
	policy, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	policy.Toggle()

	if err := h.repo.Update(ctx, policy); err != nil {
		h.logger.Errorf("failed to toggle policy: %v", err)
		return err
	}

	return nil
}
