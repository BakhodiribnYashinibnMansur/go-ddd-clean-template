package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// DeleteUserCommand holds the input for soft-deleting a user.
type DeleteUserCommand struct {
	ID uuid.UUID
}

// DeleteUserHandler handles the DeleteUserCommand.
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

// Handle executes the DeleteUserCommand.
func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUserCommand) error {
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
