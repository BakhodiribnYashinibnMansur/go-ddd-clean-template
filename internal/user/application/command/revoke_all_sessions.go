package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// RevokeAllSessionsCommand holds the input for revoking all user sessions.
type RevokeAllSessionsCommand struct {
	UserID uuid.UUID
}

// RevokeAllSessionsHandler handles the RevokeAllSessionsCommand.
type RevokeAllSessionsHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewRevokeAllSessionsHandler creates a new RevokeAllSessionsHandler.
func NewRevokeAllSessionsHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *RevokeAllSessionsHandler {
	return &RevokeAllSessionsHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the RevokeAllSessionsCommand.
func (h *RevokeAllSessionsHandler) Handle(ctx context.Context, cmd RevokeAllSessionsCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "RevokeAllSessionsHandler.Handle")
	defer func() { end(err) }()

	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	user.RevokeAllSessions()

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to save user after revoke-all: %v", err)
		return err
	}

	return nil
}
