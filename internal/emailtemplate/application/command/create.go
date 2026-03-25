package command

import (
	"context"

	"gct/internal/emailtemplate/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateCommand holds the input for creating a new email template.
type CreateCommand struct {
	Name      string
	Subject   string
	HTMLBody  string
	TextBody  string
	Variables []string
}

// CreateHandler handles the CreateCommand.
type CreateHandler struct {
	repo     domain.EmailTemplateRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler creates a new CreateHandler.
func NewCreateHandler(
	repo domain.EmailTemplateRepository,
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
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) error {
	et := domain.NewEmailTemplate(cmd.Name, cmd.Subject, cmd.HTMLBody, cmd.TextBody, cmd.Variables)

	if err := h.repo.Save(ctx, et); err != nil {
		h.logger.Errorf("failed to save email template: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, et.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
