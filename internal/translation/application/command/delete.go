package command

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

// DeleteTranslationCommand represents an intent to permanently remove a translation entry.
// Once deleted, any UI referencing this key+language will fall back to the default locale or show a missing-key placeholder.
type DeleteTranslationCommand struct {
	ID uuid.UUID
}

// DeleteTranslationHandler performs hard-delete of translations through the repository.
// No domain events are emitted — callers needing cache invalidation should handle that separately.
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

// Handle deletes the translation identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteTranslationHandler) Handle(ctx context.Context, cmd DeleteTranslationCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete translation: %v", err)
		return err
	}
	return nil
}
