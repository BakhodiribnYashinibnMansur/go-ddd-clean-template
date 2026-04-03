package command

import (
	"context"

	"gct/internal/session/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// RevokeSessionCommand holds the input for revoking a single session.
type RevokeSessionCommand struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// RevokeSessionHandler publishes a SessionRevokeRequested event.
// The actual revocation is performed by the User BC, which subscribes to this event.
type RevokeSessionHandler struct {
	eventBus application.EventBus
	logger   logger.Log
}

// NewRevokeSessionHandler creates a new RevokeSessionHandler.
func NewRevokeSessionHandler(eventBus application.EventBus, l logger.Log) *RevokeSessionHandler {
	return &RevokeSessionHandler{eventBus: eventBus, logger: l}
}

// Handle publishes a revoke-requested event for a single session.
func (h *RevokeSessionHandler) Handle(ctx context.Context, cmd RevokeSessionCommand) error {
	event := domain.NewSessionRevokeRequested(cmd.UserID, cmd.SessionID)

	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Errorc(ctx, "failed to publish session.revoke_requested",
			logger.F{Op: "RevokeSession", Entity: "session", EntityID: cmd.SessionID, Err: err}.KV()...)
		return err
	}

	return nil
}
