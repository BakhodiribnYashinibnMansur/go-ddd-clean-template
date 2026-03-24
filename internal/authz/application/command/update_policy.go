package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdatePolicyCommand holds the input for updating an existing policy.
type UpdatePolicyCommand struct {
	ID         uuid.UUID
	Effect     *domain.PolicyEffect
	Priority   *int
	Conditions map[string]any
}

// UpdatePolicyHandler handles the UpdatePolicyCommand.
type UpdatePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewUpdatePolicyHandler creates a new UpdatePolicyHandler.
func NewUpdatePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *UpdatePolicyHandler {
	return &UpdatePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the UpdatePolicyCommand.
func (h *UpdatePolicyHandler) Handle(ctx context.Context, cmd UpdatePolicyCommand) error {
	policy, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if cmd.Effect != nil {
		policy.SetEffect(*cmd.Effect)
	}
	if cmd.Priority != nil {
		policy.SetPriority(*cmd.Priority)
	}
	if cmd.Conditions != nil {
		policy.SetConditions(cmd.Conditions)
	}

	if err := h.repo.Update(ctx, policy); err != nil {
		h.logger.Errorf("failed to update policy: %v", err)
		return err
	}

	return nil
}
