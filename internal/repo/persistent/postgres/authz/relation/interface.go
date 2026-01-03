package relation

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

type RepoI interface {
	Create(ctx context.Context, r *domain.Relation) error
	Get(ctx context.Context, filter *domain.RelationFilter) (*domain.Relation, error)
	Gets(ctx context.Context, filter *domain.RelationsFilter) ([]*domain.Relation, int, error)
	Update(ctx context.Context, r *domain.Relation) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddUser(ctx context.Context, relationID, userID uuid.UUID) error
	RemoveUser(ctx context.Context, relationID, userID uuid.UUID) error
}
