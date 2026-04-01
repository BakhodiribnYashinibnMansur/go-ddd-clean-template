package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// DeleteUserCommand represents an intent to soft-delete a user by their unique identifier.
// The user is deactivated and marked as deleted but not physically removed from the database.
type DeleteUserCommand struct {
	ID uuid.UUID
}

// DeleteUserHandler performs a two-step soft-delete: deactivation followed by a soft-delete timestamp.
// The user record is preserved for audit/recovery; domain events are emitted for downstream cleanup.
type DeleteUserHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteUserHandler creates a new DeleteUserHandler.
func NewDeleteUserHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
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

	user, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	user.Deactivate()
	user.SoftDelete()

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to delete user: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
