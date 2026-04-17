package repository

import (
	"context"

	"gct/internal/context/content/generic/translation/domain/entity"
	shareddomain "gct/internal/kernel/domain"
)

// TranslationRepository is the write-side persistence contract for the Translation aggregate.
// Implementations must return ErrTranslationNotFound from FindByID when no row matches.
type TranslationRepository interface {
	Save(ctx context.Context, q shareddomain.Querier, e *entity.Translation) error
	FindByID(ctx context.Context, id entity.TranslationID) (*entity.Translation, error)
	Update(ctx context.Context, q shareddomain.Querier, e *entity.Translation) error
	Delete(ctx context.Context, q shareddomain.Querier, id entity.TranslationID) error
	List(ctx context.Context, filter TranslationFilter) ([]*entity.Translation, int64, error)
}

// TranslationFilter carries optional filtering parameters. Nil fields are ignored by the repository.
type TranslationFilter struct {
	Key      *string
	Language *string
	Group    *string
	Limit    int64
	Offset   int64
}
