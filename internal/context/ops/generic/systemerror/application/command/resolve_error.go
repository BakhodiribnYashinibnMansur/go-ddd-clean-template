package command

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
	syserrentity "gct/internal/context/ops/generic/systemerror/domain/entity"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"

	"github.com/google/uuid"
)

// ResolveErrorCommand represents an intent to mark a system error as resolved by a specific user.
// This is an irreversible status transition — once resolved, the error cannot be re-opened.
type ResolveErrorCommand struct {
	ID         syserrentity.SystemErrorID
	ResolvedBy uuid.UUID
}

// ResolveErrorHandler transitions a system error to the resolved state via a load-modify-save cycle.
// Callers are responsible for verifying that ResolvedBy refers to a user with sufficient privileges.
type ResolveErrorHandler struct {
	repo      syserrrepo.SystemErrorRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewResolveErrorHandler creates a new ResolveErrorHandler.
func NewResolveErrorHandler(
	repo syserrrepo.SystemErrorRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *ResolveErrorHandler {
	return &ResolveErrorHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle loads the system error, marks it resolved, persists the update, and publishes domain events.
// Returns not-found or repository errors; event bus failures are logged but do not fail the operation.
func (h *ResolveErrorHandler) Handle(ctx context.Context, cmd ResolveErrorCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ResolveErrorHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ResolveError", "system_error")()

	se, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	se.Resolve(cmd.ResolvedBy)

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, se); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "ResolveError", Entity: "system_error", EntityID: cmd.ID.String(), Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, se.Events)
}
