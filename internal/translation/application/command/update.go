package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

// UpdateTranslationCommand represents a partial update to an existing translation record.
// Pointer fields use nil-means-unchanged semantics, so callers only set the fields they want to modify.
type UpdateTranslationCommand struct {
	ID       uuid.UUID
	Key      *string
	Language *string
	Value    *string
	Group    *string
}

// UpdateTranslationHandler applies partial updates to translations via a load-modify-save cycle.
// Event bus failures are logged but do not cause the handler to return an error.
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

// Handle loads the translation by ID, applies the partial update, and persists the result.
// Returns not-found or repository errors to the caller; authorization is the caller's responsibility.
func (h *UpdateTranslationHandler) Handle(ctx context.Context, cmd UpdateTranslationCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateTranslationHandler.Handle")
	defer func() { end(err) }()

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
