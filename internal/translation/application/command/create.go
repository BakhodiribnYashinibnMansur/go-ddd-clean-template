package command

import (
	"context"

	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/translation/domain"
)

// CreateTranslationCommand represents an intent to add a new localized string entry.
// The Key+Language pair must be unique; the repository will reject duplicates.
type CreateTranslationCommand struct {
	Key      string
	Language string
	Value    string
	Group    string
}

// CreateTranslationHandler orchestrates translation creation and domain event publication.
// Event bus failures are logged but do not roll back the persisted translation.
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

// Handle persists the new translation and publishes resulting domain events.
// Returns repository errors (e.g., duplicate key+language, connection failure) directly to the caller.
func (h *CreateTranslationHandler) Handle(ctx context.Context, cmd CreateTranslationCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateTranslationHandler.Handle")
	defer func() { end(err) }()

	t := domain.NewTranslation(cmd.Key, cmd.Language, cmd.Value, cmd.Group)

	if err := h.repo.Save(ctx, t); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateTranslation", Entity: "translation", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, t.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateTranslation", Entity: "translation", Err: err}.KV()...)
	}

	return nil
}
