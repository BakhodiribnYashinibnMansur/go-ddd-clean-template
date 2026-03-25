package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/translation/domain"
)

// CreateTranslationCommand holds the input for creating a new translation.
type CreateTranslationCommand struct {
	Key      string
	Language string
	Value    string
	Group    string
}

// CreateTranslationHandler handles the CreateTranslationCommand.
type CreateTranslationHandler struct {
	repo     domain.TranslationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateTranslationHandler creates a new CreateTranslationHandler.
func NewCreateTranslationHandler(
	repo domain.TranslationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateTranslationHandler {
	return &CreateTranslationHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateTranslationCommand.
func (h *CreateTranslationHandler) Handle(ctx context.Context, cmd CreateTranslationCommand) error {
	t := domain.NewTranslation(cmd.Key, cmd.Language, cmd.Value, cmd.Group)

	if err := h.repo.Save(ctx, t); err != nil {
		h.logger.Errorf("failed to save translation: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, t.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
