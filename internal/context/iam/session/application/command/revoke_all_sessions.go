package command

import (
	"context"

	"gct/internal/context/iam/session/domain"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/google/uuid"
)

// RevokeAllSessionsCommand holds the input for revoking all sessions of a user.
type RevokeAllSessionsCommand struct {
	UserID uuid.UUID
}

// RevokeAllSessionsHandler publishes a SessionRevokeAllRequested event.
// The actual revocation is performed by the User BC, which subscribes to this event.
type RevokeAllSessionsHandler struct {
	eventBus application.EventBus
	logger   logger.Log
}

// NewRevokeAllSessionsHandler creates a new RevokeAllSessionsHandler.
func NewRevokeAllSessionsHandler(eventBus application.EventBus, l logger.Log) *RevokeAllSessionsHandler {
	return &RevokeAllSessionsHandler{eventBus: eventBus, logger: l}
}

// Handle publishes a revoke-all-requested event for all user sessions.
func (h *RevokeAllSessionsHandler) Handle(ctx context.Context, cmd RevokeAllSessionsCommand) error {
	event := domain.NewSessionRevokeAllRequested(cmd.UserID)

	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Errorc(ctx, "failed to publish session.revoke_all_requested",
			logger.F{Op: "RevokeAllSessions", Entity: "session", EntityID: cmd.UserID, Err: err}.KV()...)
		return err
	}

	return nil
}
