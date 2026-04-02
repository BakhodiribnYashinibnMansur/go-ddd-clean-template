package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	"gct/internal/shared/infrastructure/pgxutil"
	appdto "gct/internal/translation/application"
	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

// GetTranslationQuery holds the input for getting a single translation.
type GetTranslationQuery struct {
	ID uuid.UUID
}

// GetTranslationHandler handles the GetTranslationQuery.
type GetTranslationHandler struct {
	readRepo domain.TranslationReadRepository
	logger   logger.Log
}

// NewGetTranslationHandler creates a new GetTranslationHandler.
func NewGetTranslationHandler(readRepo domain.TranslationReadRepository, l logger.Log) *GetTranslationHandler {
	return &GetTranslationHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetTranslationQuery and returns a TranslationView.
func (h *GetTranslationHandler) Handle(ctx context.Context, q GetTranslationQuery) (result *appdto.TranslationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetTranslationHandler.Handle")
	defer func() { end(err) }()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetTranslation", Entity: "translation", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.TranslationView{
		ID:        v.ID,
		Key:       v.Key,
		Language:  v.Language,
		Value:     v.Value,
		Group:     v.Group,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}, nil
}
