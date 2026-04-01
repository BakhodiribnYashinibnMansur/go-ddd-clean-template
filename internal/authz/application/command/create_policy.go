package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreatePolicyCommand represents an intent to create an authorization policy binding a permission to an effect.
// Priority determines evaluation order when multiple policies match; Conditions enable attribute-based access control (ABAC).
type CreatePolicyCommand struct {
	PermissionID uuid.UUID
	Effect       domain.PolicyEffect
	Priority     int
	Conditions   map[string]any
}

// CreatePolicyHandler persists new authorization policies via the repository.
// No domain events are emitted — policy evaluation relies on direct repository reads.
type CreatePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewCreatePolicyHandler wires dependencies for policy creation.
func NewCreatePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *CreatePolicyHandler {
	return &CreatePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle creates a policy with the specified effect and priority, optionally attaches conditions, and persists it.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreatePolicyHandler) Handle(ctx context.Context, cmd CreatePolicyCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreatePolicyHandler.Handle")
	defer func() { end(err) }()

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
