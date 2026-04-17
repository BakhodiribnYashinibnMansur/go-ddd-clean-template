package command

import (
	"context"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// DeleteIPRuleCommand represents an intent to permanently remove an IP rule by its unique identifier.
// Once deleted, any traffic previously matched by this rule will fall through to the default policy.
type DeleteIPRuleCommand struct {
	ID ipruleentity.IPRuleID
}

// DeleteIPRuleHandler orchestrates IP rule deletion through the repository layer.
// It enforces a hard-delete strategy — no soft-delete or audit trail is maintained at this level.
// Callers are responsible for authorization checks before invoking this handler.
type DeleteIPRuleHandler struct {
	repo      iprulerepo.IPRuleRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewDeleteIPRuleHandler wires up the handler with its required dependencies.
func NewDeleteIPRuleHandler(
	repo iprulerepo.IPRuleRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *DeleteIPRuleHandler {
	return &DeleteIPRuleHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle performs the deletion of the IP rule identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteIPRuleHandler) Handle(ctx context.Context, cmd DeleteIPRuleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteIPRuleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteIPRule", "ip_rule")()

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Delete(ctx, q, cmd.ID); err != nil {
			h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteIPRule", Entity: "ip_rule", EntityID: cmd.ID.String(), Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, func() []shareddomain.DomainEvent { return nil })
}
