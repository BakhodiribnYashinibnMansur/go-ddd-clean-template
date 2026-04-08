package command

import (
	"context"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcodeevent "gct/internal/context/admin/supporting/errorcode/domain/event"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// DeleteErrorCodeCommand represents an intent to permanently remove a standardized error code.
// Callers should ensure no API handlers still reference this code before deletion to avoid runtime lookup failures.
type DeleteErrorCodeCommand struct {
	ID errcodeentity.ErrorCodeID
}

// DeleteErrorCodeHandler performs hard deletion of an error code via the repository.
// Publishes an ErrorCodeDeleted event so subscribers can evict the code from in-memory caches.
type DeleteErrorCodeHandler struct {
	repo      errcoderepo.ErrorCodeRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewDeleteErrorCodeHandler wires dependencies for error code deletion.
func NewDeleteErrorCodeHandler(
	repo errcoderepo.ErrorCodeRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *DeleteErrorCodeHandler {
	return &DeleteErrorCodeHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle fetches the error code to capture its code string, deletes it, and publishes a deleted event.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeleteErrorCodeHandler) Handle(ctx context.Context, cmd DeleteErrorCodeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteErrorCodeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteErrorCode", "error_code")()

	ec, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	event := errcodeevent.NewErrorCodeDeleted(cmd.ID.UUID(), ec.Code())

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Delete(ctx, cmd.ID); err != nil {
			h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteErrorCode", Entity: "error_code", EntityID: cmd.ID.String(), Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, func() []shareddomain.DomainEvent { return []shareddomain.DomainEvent{event} })
}
