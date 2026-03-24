package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreatePolicyCommand holds the input for creating a new policy.
type CreatePolicyCommand struct {
	PermissionID uuid.UUID
	Effect       domain.PolicyEffect
	Priority     int
	Conditions   map[string]any
}

// CreatePolicyHandler handles the CreatePolicyCommand.
type CreatePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewCreatePolicyHandler creates a new CreatePolicyHandler.
func NewCreatePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *CreatePolicyHandler {
	return &CreatePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the CreatePolicyCommand.
func (h *CreatePolicyHandler) Handle(ctx context.Context, cmd CreatePolicyCommand) error {
	policy := domain.NewPolicy(cmd.PermissionID, cmd.Effect)
	policy.SetPriority(cmd.Priority)
	if cmd.Conditions != nil {
		policy.SetConditions(cmd.Conditions)
	}

	if err := h.repo.Save(ctx, policy); err != nil {
		h.logger.Errorf("failed to save policy: %v", err)
		return err
	}

	return nil
}
