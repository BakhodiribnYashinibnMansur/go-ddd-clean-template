package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

// UpdateTranslationCommand holds the input for updating a translation.
type UpdateTranslationCommand struct {
	ID       uuid.UUID
	Key      *string
	Language *string
	Value    *string
	Group    *string
}

// UpdateTranslationHandler handles the UpdateTranslationCommand.
type UpdateTranslationHandler struct {
	repo     domain.TranslationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateTranslationHandler creates a new UpdateTranslationHandler.
func NewUpdateTranslationHandler(
	repo domain.TranslationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateTranslationHandler {
	return &UpdateTranslationHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateTranslationCommand.
func (h *UpdateTranslationHandler) Handle(ctx context.Context, cmd UpdateTranslationCommand) error {
	t, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	t.Update(cmd.Key, cmd.Language, cmd.Value, cmd.Group)

	if err := h.repo.Update(ctx, t); err != nil {
		h.logger.Errorf("failed to update translation: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, t.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
