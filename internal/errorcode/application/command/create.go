package command

import (
	"context"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateErrorCodeCommand holds the input for creating a new error code.
type CreateErrorCodeCommand struct {
	Code       string
	Message    string
	HTTPStatus int
	Category   string
	Severity   string
	Retryable  bool
	RetryAfter int
	Suggestion string
}

// CreateErrorCodeHandler handles the CreateErrorCodeCommand.
type CreateErrorCodeHandler struct {
	repo     domain.ErrorCodeRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateErrorCodeHandler creates a new CreateErrorCodeHandler.
func NewCreateErrorCodeHandler(
	repo domain.ErrorCodeRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateErrorCodeHandler {
	return &CreateErrorCodeHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateErrorCodeCommand.
func (h *CreateErrorCodeHandler) Handle(ctx context.Context, cmd CreateErrorCodeCommand) error {
	ec := domain.NewErrorCode(
		cmd.Code, cmd.Message, cmd.HTTPStatus,
		cmd.Category, cmd.Severity,
		cmd.Retryable, cmd.RetryAfter, cmd.Suggestion,
	)

	if err := h.repo.Save(ctx, ec); err != nil {
		h.logger.Errorf("failed to save error code: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, ec.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
