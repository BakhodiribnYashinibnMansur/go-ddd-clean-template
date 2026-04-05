package command

import (
	"context"
	"fmt"

	"gct/internal/context/content/notification/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreateCommand holds the input for creating a new notification.
type CreateCommand struct {
	UserID  uuid.UUID
	Title   string
	Message string
	Type    string
}

// CreateHandler handles the CreateCommand.
type CreateHandler struct {
	repo     domain.NotificationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler creates a new CreateHandler.
func NewCreateHandler(
	repo domain.NotificationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateCommand.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateNotification", "notification")()

	n, err := domain.NewNotification(cmd.UserID, cmd.Title, cmd.Message, cmd.Type)
	if err != nil {
		return fmt.Errorf("create_notification: %w", err)
	}

	if err := h.repo.Save(ctx, n); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateNotification", Entity: "notification", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, n.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateNotification", Entity: "notification", Err: err}.KV()...)
	}

	return nil
}
