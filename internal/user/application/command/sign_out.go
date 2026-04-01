package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// SignOutCommand holds the input for user sign-out.
type SignOutCommand struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// SignOutHandler handles the SignOutCommand.
type SignOutHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewSignOutHandler creates a new SignOutHandler.
func NewSignOutHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *SignOutHandler {
	return &SignOutHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the SignOutCommand.
func (h *SignOutHandler) Handle(ctx context.Context, cmd SignOutCommand) error {
	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.RevokeSession(cmd.SessionID); err != nil {
		return err
	}

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to save user after sign-out: %v", err)
		return err
	}

	return nil
}
