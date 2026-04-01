package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// UpdateUserCommand represents a partial update to a user's profile fields.
// Pointer fields use nil-means-unchanged semantics. Phone, password, and role are excluded —
// use dedicated commands (ChangeRole, etc.) for those privileged mutations.
type UpdateUserCommand struct {
	ID         uuid.UUID
	Email      *string
	Username   *string
	Attributes map[string]string
}

// UpdateUserHandler applies partial profile updates via a load-reconstruct-save cycle.
// Because the User aggregate uses unexported fields, the handler reconstructs the entity with merged values.
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

// Handle loads the user, merges changed fields with existing data, reconstructs the aggregate, and persists it.
// Calls Touch() to update the modification timestamp. Returns domain or repository errors to the caller.
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
