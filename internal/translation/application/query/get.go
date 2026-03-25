package query

import (
	"context"

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
}

// NewGetTranslationHandler creates a new GetTranslationHandler.
func NewGetTranslationHandler(readRepo domain.TranslationReadRepository) *GetTranslationHandler {
	return &GetTranslationHandler{readRepo: readRepo}
}

// Handle executes the GetTranslationQuery and returns a TranslationView.
func (h *GetTranslationHandler) Handle(ctx context.Context, q GetTranslationQuery) (*appdto.TranslationView, error) {
	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
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
