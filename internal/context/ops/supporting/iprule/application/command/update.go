package command

import (
	"context"
	"time"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// UpdateIPRuleCommand represents a partial update to an existing IP rule identified by ID.
// Pointer fields implement patch semantics — nil means "leave unchanged," non-nil means "overwrite."
// Changing Action or IPAddress takes effect immediately for subsequent traffic evaluations.
type UpdateIPRuleCommand struct {
	ID        ipruleentity.IPRuleID
	IPAddress *string
	Action    *string
	Reason    *string
	ExpiresAt *time.Time
}

// UpdateIPRuleHandler applies partial modifications to an existing IP rule via fetch-then-update.
// Domain events are emitted so downstream caches or firewalls can refresh their rule sets.
type UpdateIPRuleHandler struct {
	repo      iprulerepo.IPRuleRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewUpdateIPRuleHandler wires up the handler with its required dependencies.
func NewUpdateIPRuleHandler(
	repo iprulerepo.IPRuleRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *UpdateIPRuleHandler {
	return &UpdateIPRuleHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle fetches the IP rule by ID, applies the patch via domain logic, and persists the result.
// Returns a repository error if the rule is not found. Event publish failures are logged but non-fatal.
func (h *UpdateIPRuleHandler) Handle(ctx context.Context, cmd UpdateIPRuleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateIPRuleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateIPRule", "ip_rule")()

	r, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	r.Update(cmd.IPAddress, cmd.Action, cmd.Reason, cmd.ExpiresAt)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Update(ctx, q, r); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateIPRule", Entity: "ip_rule", EntityID: cmd.ID.String(), Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, r.Events)
}
