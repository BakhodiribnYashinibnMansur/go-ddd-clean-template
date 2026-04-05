package command

import (
	"context"

	"gct/internal/context/iam/authz/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeletePolicyCommand represents an intent to permanently remove an authorization policy.
// Once deleted, any access previously governed by this policy falls through to the next matching policy or default deny.
type DeletePolicyCommand struct {
	ID domain.PolicyID
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
func (h *DeletePolicyHandler) Handle(ctx context.Context, cmd DeletePolicyCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeletePolicyHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeletePolicy", "policy")()

	if err := h.repo.Delete(ctx, cmd.ID.UUID()); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeletePolicy", Entity: "policy", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
