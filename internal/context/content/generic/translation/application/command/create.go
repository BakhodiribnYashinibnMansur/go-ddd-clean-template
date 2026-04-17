package command

import (
	"context"

	translationentity "gct/internal/context/content/generic/translation/domain/entity"
	translationrepo "gct/internal/context/content/generic/translation/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
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
	repo      translationrepo.TranslationRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateTranslationHandler creates a new CreateTranslationHandler.
func NewCreateTranslationHandler(
	repo translationrepo.TranslationRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateTranslationHandler {
	return &CreateTranslationHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists the new translation and publishes resulting domain events.
// Returns repository errors (e.g., duplicate key+language, connection failure) directly to the caller.
func (h *CreateTranslationHandler) Handle(ctx context.Context, cmd CreateTranslationCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateTranslationHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateTranslation", "translation")()

	t := translationentity.NewTranslation(cmd.Key, cmd.Language, cmd.Value, cmd.Group)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, t); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateTranslation", Entity: "translation", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, t.Events)
}
