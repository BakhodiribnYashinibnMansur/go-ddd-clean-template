package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	"gct/internal/shared/infrastructure/pgxutil"
	appdto "gct/internal/translation/application"
	"gct/internal/translation/domain"
)

// ListTranslationsQuery holds the input for listing translations.
type ListTranslationsQuery struct {
	Filter domain.TranslationFilter
}

// ListTranslationsResult holds the output of the list translations query.
type ListTranslationsResult struct {
	Translations []*appdto.TranslationView
	Total        int64
}

// ListTranslationsHandler handles the ListTranslationsQuery.
type ListTranslationsHandler struct {
	readRepo domain.TranslationReadRepository
	logger   logger.Log
}

// NewListTranslationsHandler creates a new ListTranslationsHandler.
func NewListTranslationsHandler(readRepo domain.TranslationReadRepository, l logger.Log) *ListTranslationsHandler {
	return &ListTranslationsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListTranslationsQuery and returns a list of TranslationView with total count.
func (h *ListTranslationsHandler) Handle(ctx context.Context, q ListTranslationsQuery) (result *ListTranslationsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListTranslationsHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListTranslations", Entity: "translation", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*appdto.TranslationView, len(views))
	for i, v := range views {
		items[i] = &appdto.TranslationView{
			ID:        v.ID,
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
