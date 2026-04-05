package command

import (
	"context"

	"gct/internal/context/content/notification/domain"
	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// DeleteCommand holds the input for deleting a notification.
type DeleteCommand struct {
	ID uuid.UUID
}

// DeleteHandler handles the DeleteCommand.
type DeleteHandler struct {
	repo     domain.NotificationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler creates a new DeleteHandler.
func NewDeleteHandler(
	repo domain.NotificationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteHandler {
	return &DeleteHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the DeleteCommand.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteNotification", "notification")()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteNotification", Entity: "notification", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
