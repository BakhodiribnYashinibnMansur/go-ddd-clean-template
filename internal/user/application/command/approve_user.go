package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// ApproveUserCommand holds the input for approving a user.
type ApproveUserCommand struct {
	ID uuid.UUID
}

// ApproveUserHandler handles the ApproveUserCommand.
type ApproveUserHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewApproveUserHandler creates a new ApproveUserHandler.
func NewApproveUserHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *ApproveUserHandler {
	return &ApproveUserHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the ApproveUserCommand.
func (h *ApproveUserHandler) Handle(ctx context.Context, cmd ApproveUserCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ApproveUserHandler.Handle")
	defer func() { end(err) }()

	user, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	user.Approve()

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to save approved user: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Errorf("failed to publish approve events: %v", err)
	}

	return nil
}
