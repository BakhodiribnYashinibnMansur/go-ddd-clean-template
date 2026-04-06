package repository

import (
	"context"
	"time"

	"gct/internal/context/content/generic/translation/domain/entity"
)

// TranslationView is a flat read-model projection for API responses, bypassing aggregate reconstruction.
type TranslationView struct {
	ID        entity.TranslationID `json:"id"`
	Key       string               `json:"key"`
	Language  string               `json:"language"`
	Value     string               `json:"value"`
	Group     string               `json:"group"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// TranslationReadRepository provides read-only access optimized for listing and detail views.
type TranslationReadRepository interface {
	FindByID(ctx context.Context, id entity.TranslationID) (*TranslationView, error)
	List(ctx context.Context, filter TranslationFilter) ([]*TranslationView, int64, error)
}
