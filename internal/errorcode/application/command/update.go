package command

import (
	"context"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateErrorCodeCommand holds the input for updating an existing error code.
type UpdateErrorCodeCommand struct {
	ID         uuid.UUID
	Message    string
	HTTPStatus int
	Category   string
	Severity   string
	Retryable  bool
	RetryAfter int
	Suggestion string
}

// UpdateErrorCodeHandler handles the UpdateErrorCodeCommand.
type UpdateErrorCodeHandler struct {
	repo     domain.ErrorCodeRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateErrorCodeHandler creates a new UpdateErrorCodeHandler.
func NewUpdateErrorCodeHandler(
	repo domain.ErrorCodeRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateErrorCodeHandler {
	return &UpdateErrorCodeHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateErrorCodeCommand.
func (h *UpdateErrorCodeHandler) Handle(ctx context.Context, cmd UpdateErrorCodeCommand) error {
	ec, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	ec.Update(
		cmd.Message, cmd.HTTPStatus,
		cmd.Category, cmd.Severity,
		cmd.Retryable, cmd.RetryAfter, cmd.Suggestion,
	)

	if err := h.repo.Update(ctx, ec); err != nil {
		h.logger.Errorf("failed to update error code: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, ec.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
