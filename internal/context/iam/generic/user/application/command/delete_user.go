package command

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteUserCommand represents an intent to soft-delete a user by their unique identifier.
// The user is deactivated and marked as deleted but not physically removed from the database.
type DeleteUserCommand struct {
	ID userentity.UserID
}

// DeleteUserHandler performs a two-step soft-delete: deactivation followed by a soft-delete timestamp.
// The user record is preserved for audit/recovery; domain events are emitted for downstream cleanup.
type DeleteUserHandler struct {
	repo     userrepo.UserRepository
	eventBus application.EventBus
	logger   commandLogger
}

// NewDeleteUserHandler creates a new DeleteUserHandler.
func NewDeleteUserHandler(
	repo userrepo.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
) *DeleteUserHandler {
	return &DeleteUserHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
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

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "DeleteUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "DeleteUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
