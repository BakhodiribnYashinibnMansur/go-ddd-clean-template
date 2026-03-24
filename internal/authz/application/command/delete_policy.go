package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeletePolicyCommand holds the input for deleting a policy.
type DeletePolicyCommand struct {
	ID uuid.UUID
}

// DeletePolicyHandler handles the DeletePolicyCommand.
type DeletePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewDeletePolicyHandler creates a new DeletePolicyHandler.
func NewDeletePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *DeletePolicyHandler {
	return &DeletePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeletePolicyCommand.
func (h *DeletePolicyHandler) Handle(ctx context.Context, cmd DeletePolicyCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete policy: %v", err)
		return err
	}

	return nil
}
