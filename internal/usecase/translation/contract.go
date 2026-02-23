package translation

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// Repository defines translation database operations.
type Repository interface {
	Upsert(ctx context.Context, entityType string, entityID uuid.UUID, langCode string, data map[string]string) error
	Gets(ctx context.Context, filter domain.TranslationFilter) ([]*domain.Translation, error)
	Delete(ctx context.Context, filter domain.TranslationFilter) error
}

// UseCaseI defines the translation business logic interface.
type UseCaseI interface {
	Upsert(ctx context.Context, entityType string, entityID uuid.UUID, req domain.UpsertTranslationsRequest) error
	Gets(ctx context.Context, filter domain.TranslationFilter) (domain.EntityTranslations, error)
	Delete(ctx context.Context, filter domain.TranslationFilter) error
}
