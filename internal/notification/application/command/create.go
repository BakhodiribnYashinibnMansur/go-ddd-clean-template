package command

import (
	"context"

	"gct/internal/notification/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

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

	n := domain.NewNotification(cmd.UserID, cmd.Title, cmd.Message, cmd.Type)

	if err := h.repo.Save(ctx, n); err != nil {
		h.logger.Errorf("failed to save notification: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, n.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
