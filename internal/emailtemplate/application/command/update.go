package command

import (
	"context"

	"gct/internal/emailtemplate/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateCommand holds the input for updating an email template.
type UpdateCommand struct {
	ID        uuid.UUID
	Name      *string
	Subject   *string
	HTMLBody  *string
	TextBody  *string
	Variables []string
}

// UpdateHandler handles the UpdateCommand.
type UpdateHandler struct {
	repo     domain.EmailTemplateRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler creates a new UpdateHandler.
func NewUpdateHandler(
	repo domain.EmailTemplateRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateHandler {
	return &UpdateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateCommand.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) error {
	et, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	et.UpdateDetails(cmd.Name, cmd.Subject, cmd.HTMLBody, cmd.TextBody, cmd.Variables)

	if err := h.repo.Update(ctx, et); err != nil {
		h.logger.Errorf("failed to update email template: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, et.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
