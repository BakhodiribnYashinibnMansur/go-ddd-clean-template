package domain

import (
	"context"
	"time"
)

// TranslationFilter carries optional filtering parameters. Nil fields are ignored by the repository.
type TranslationFilter struct {
	Key      *string
	Language *string
	Group    *string
	Limit    int64
	Offset   int64
}

// TranslationRepository is the write-side persistence contract for the Translation aggregate.
// Implementations must return ErrTranslationNotFound from FindByID when no row matches.
type TranslationRepository interface {
	Save(ctx context.Context, entity *Translation) error
	FindByID(ctx context.Context, id TranslationID) (*Translation, error)
	Update(ctx context.Context, entity *Translation) error
	Delete(ctx context.Context, id TranslationID) error
	List(ctx context.Context, filter TranslationFilter) ([]*Translation, int64, error)
}

// TranslationView is a flat read-model projection for API responses, bypassing aggregate reconstruction.
type TranslationView struct {
	ID        TranslationID `json:"id"`
	Key       string        `json:"key"`
	Language  string        `json:"language"`
	Value     string        `json:"value"`
	Group     string        `json:"group"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// TranslationReadRepository provides read-only access optimized for listing and detail views.
type TranslationReadRepository interface {
	FindByID(ctx context.Context, id TranslationID) (*TranslationView, error)
	List(ctx context.Context, filter TranslationFilter) ([]*TranslationView, int64, error)
}
