package command

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

// DeleteTranslationCommand holds the input for deleting a translation.
type DeleteTranslationCommand struct {
	ID uuid.UUID
}

// DeleteTranslationHandler handles the DeleteTranslationCommand.
type DeleteTranslationHandler struct {
	repo   domain.TranslationRepository
	logger logger.Log
}

// NewDeleteTranslationHandler creates a new DeleteTranslationHandler.
func NewDeleteTranslationHandler(
	repo domain.TranslationRepository,
	logger logger.Log,
) *DeleteTranslationHandler {
	return &DeleteTranslationHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteTranslationCommand.
func (h *DeleteTranslationHandler) Handle(ctx context.Context, cmd DeleteTranslationCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete translation: %v", err)
		return err
	}
	return nil
}
