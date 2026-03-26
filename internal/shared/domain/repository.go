package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the generic write-side persistence contract for aggregate roots.
// Implementations must map domain aggregates to and from their storage representation.
// FindByID should return a fully reconstituted aggregate (including child entities).
// Delete performs a hard delete — use BaseEntity.SoftDelete for logical deletion before calling Update.
type Repository[T any] interface {
	Save(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter Pagination) ([]*T, int64, error)
}
