package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the generic persistence interface for domain entities.
type Repository[T any] interface {
	Save(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter Pagination) ([]*T, int64, error)
}
