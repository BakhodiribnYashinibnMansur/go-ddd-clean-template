package command

import (
	"context"

	"gct/internal/context/iam/authz/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// TogglePolicyCommand represents an intent to flip a policy between active and inactive states.
// This is an idempotent toggle — calling it twice restores the original state.
// Disabled policies are skipped during authorization evaluation without being deleted.
type TogglePolicyCommand struct {
	ID domain.PolicyID
}

// TogglePolicyHandler orchestrates the enable/disable lifecycle of an authorization policy.
// Changes take effect immediately on the next authorization evaluation — there is no propagation delay.
type TogglePolicyHandler struct {
	repo   domain.PolicyRepository
	logger logger.Log
}

// NewTogglePolicyHandler wires dependencies for policy toggling.
func NewTogglePolicyHandler(
	repo domain.PolicyRepository,
	logger logger.Log,
) *TogglePolicyHandler {
	return &TogglePolicyHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle fetches the policy, inverts its active state, and persists the change.
// Returns a repository error if the policy is not found or the update fails.
func (h *TogglePolicyHandler) Handle(ctx context.Context, cmd TogglePolicyCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "TogglePolicyHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "TogglePolicy", "policy")()

	policy, err := h.repo.FindByID(ctx, cmd.ID.UUID())
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	policy.Toggle()

	if err := h.repo.Update(ctx, policy); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "TogglePolicy", Entity: "policy", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
