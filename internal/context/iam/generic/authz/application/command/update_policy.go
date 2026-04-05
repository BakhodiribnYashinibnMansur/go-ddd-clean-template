package command

import (
	"context"

	"gct/internal/context/iam/generic/authz/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UpdatePolicyCommand represents a partial update to an existing authorization policy.
// Nil pointer fields are left unchanged. Conditions use a full-replace strategy — pass nil to keep existing, pass a map to overwrite.
type UpdatePolicyCommand struct {
	ID         domain.PolicyID
	Effect     *domain.PolicyEffect
	Priority   *int
	Conditions map[string]any
}

// UpdatePolicyHandler applies partial modifications to an existing policy using a fetch-mutate-persist pattern.
// Callers should be aware that policy changes take effect immediately on the next authorization evaluation.
type UpdatePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewUpdatePolicyHandler wires dependencies for policy updates.
func NewUpdatePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *UpdatePolicyHandler {
	return &UpdatePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle fetches the policy by ID, applies non-nil field updates, and persists the changes.
// Returns a repository error if the policy is not found or the update fails.
func (h *UpdatePolicyHandler) Handle(ctx context.Context, cmd UpdatePolicyCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdatePolicyHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdatePolicy", "policy")()

	policy, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
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
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdatePolicy", Entity: "policy", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
