package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// UpdateUserCommand holds the input for updating an existing user.
type UpdateUserCommand struct {
	ID         uuid.UUID
	Email      *string
	Username   *string
	Attributes map[string]any
}

// UpdateUserHandler handles the UpdateUserCommand.
type UpdateUserHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateUserHandler creates a new UpdateUserHandler.
func NewUpdateUserHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateUserHandler {
	return &UpdateUserHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateUserCommand.
func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand) error {
	user, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// Build options for the updated fields and reconstruct.
	// Since the domain entity uses unexported fields, we reconstruct with
	// the updated values while preserving existing data.
	email := user.Email()
	if cmd.Email != nil {
		e, err := domain.NewEmail(*cmd.Email)
		if err != nil {
			return err
		}
		email = &e
	}

	username := user.Username()
	if cmd.Username != nil {
		username = cmd.Username
	}

	attributes := user.Attributes()
	if cmd.Attributes != nil {
		attributes = cmd.Attributes
	}

	updated := domain.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		email,
		username,
		user.Password(),
		user.RoleID(),
		attributes,
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		user.Sessions(),
	)
	updated.Touch()

	if err := h.repo.Update(ctx, updated); err != nil {
		h.logger.Errorf("failed to update user: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, updated.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
