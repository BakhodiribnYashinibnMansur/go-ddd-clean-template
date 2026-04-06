package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/content/generic/translation/application/dto"
	translationrepo "gct/internal/context/content/generic/translation/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListTranslationsQuery holds the input for listing translations.
type ListTranslationsQuery struct {
	Filter translationrepo.TranslationFilter
}

// ListTranslationsResult holds the output of the list translations query.
type ListTranslationsResult struct {
	Translations []*dto.TranslationView
	Total        int64
}

// ListTranslationsHandler handles the ListTranslationsQuery.
type ListTranslationsHandler struct {
	readRepo translationrepo.TranslationReadRepository
	logger   logger.Log
}

// NewListTranslationsHandler creates a new ListTranslationsHandler.
func NewListTranslationsHandler(readRepo translationrepo.TranslationReadRepository, l logger.Log) *ListTranslationsHandler {
	return &ListTranslationsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListTranslationsQuery and returns a list of TranslationView with total count.
func (h *ListTranslationsHandler) Handle(ctx context.Context, q ListTranslationsQuery) (result *ListTranslationsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListTranslationsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListTranslations", "translation")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListTranslations", Entity: "translation", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*dto.TranslationView, len(views))
	for i, v := range views {
		items[i] = &dto.TranslationView{
			ID:        uuid.UUID(v.ID),
			Key:       v.Key,
			Language:  v.Language,
			Value:     v.Value,
			Group:     v.Group,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListTranslationsResult{
		Translations: items,
		Total:        total,
	}, nil
}
