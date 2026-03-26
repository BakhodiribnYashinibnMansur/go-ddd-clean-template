package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeletePolicyCommand represents an intent to permanently remove an authorization policy.
// Once deleted, any access previously governed by this policy falls through to the next matching policy or default deny.
type DeletePolicyCommand struct {
	ID uuid.UUID
}

// DeletePolicyHandler performs hard deletion of an authorization policy via the repository.
// Callers are responsible for verifying that removing this policy does not inadvertently grant or deny critical access.
type DeletePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewDeletePolicyHandler wires dependencies for policy deletion.
func NewDeletePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *DeletePolicyHandler {
	return &DeletePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle deletes the policy identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeletePolicyHandler) Handle(ctx context.Context, cmd DeletePolicyCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete policy: %v", err)
		return err
	}

	return nil
}
