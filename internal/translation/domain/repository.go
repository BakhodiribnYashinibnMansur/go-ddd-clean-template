package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TranslationFilter carries filtering parameters for listing translations.
type TranslationFilter struct {
	Key      *string
	Language *string
	Group    *string
	Limit    int64
	Offset   int64
}

// TranslationRepository is the write-side repository for the Translation aggregate.
type TranslationRepository interface {
	Save(ctx context.Context, entity *Translation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Translation, error)
	Update(ctx context.Context, entity *Translation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter TranslationFilter) ([]*Translation, int64, error)
}

// TranslationView is a read-model DTO for translations.
type TranslationView struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`
	Language  string    `json:"language"`
	Value     string    `json:"value"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TranslationReadRepository is the read-side repository returning projected views.
type TranslationReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*TranslationView, error)
	List(ctx context.Context, filter TranslationFilter) ([]*TranslationView, int64, error)
}
