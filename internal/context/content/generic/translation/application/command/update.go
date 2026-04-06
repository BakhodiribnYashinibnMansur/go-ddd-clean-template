package command

import (
	"context"

	translationentity "gct/internal/context/content/generic/translation/domain/entity"
	translationrepo "gct/internal/context/content/generic/translation/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UpdateTranslationCommand represents a partial update to an existing translation record.
// Pointer fields use nil-means-unchanged semantics, so callers only set the fields they want to modify.
type UpdateTranslationCommand struct {
	ID       translationentity.TranslationID
	Key      *string
	Language *string
	Value    *string
	Group    *string
}

// UpdateTranslationHandler applies partial updates to translations via a load-modify-save cycle.
// Event bus failures are logged but do not cause the handler to return an error.
type UpdateTranslationHandler struct {
	repo     translationrepo.TranslationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateTranslationHandler creates a new UpdateTranslationHandler.
func NewUpdateTranslationHandler(
	repo translationrepo.TranslationRepository,
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
	defer logger.SlowOp(h.logger, ctx, "UpdateTranslation", "translation")()

	t, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	t.Update(cmd.Key, cmd.Language, cmd.Value, cmd.Group)

	if err := h.repo.Update(ctx, t); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateTranslation", Entity: "translation", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, t.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateTranslation", Entity: "translation", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
