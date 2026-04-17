package command

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userevent "gct/internal/context/iam/generic/user/domain/event"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
)

// DeleteUserCommand represents an intent to soft-delete a user by their unique identifier.
// The user is deactivated and marked as deleted but not physically removed from the database.
type DeleteUserCommand struct {
	ID      userentity.UserID
	ActorID uuid.UUID
}

// DeleteUserHandler performs a two-step soft-delete: deactivation followed by a soft-delete timestamp.
// The user record is preserved for audit/recovery; domain events are emitted for downstream cleanup.
type DeleteUserHandler struct {
	repo      userrepo.UserRepository
	committer *outbox.EventCommitter
	logger    commandLogger
}

// NewDeleteUserHandler creates a new DeleteUserHandler.
func NewDeleteUserHandler(
	repo userrepo.UserRepository,
	committer *outbox.EventCommitter,
	logger commandLogger,
) *DeleteUserHandler {
	return &DeleteUserHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle loads the user, deactivates them, sets the soft-delete timestamp, and persists the update.
// Active sessions are not explicitly revoked here — downstream event handlers should invalidate tokens.
func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUserCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteUserHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteUser", "user")()

	user, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	user.Deactivate()
	user.SoftDelete()
	user.AddEvent(userevent.NewUserDeletedWithChanges(user.ID(), cmd.ActorID))

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Update(ctx, q, user); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "DeleteUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, user.Events)
}
