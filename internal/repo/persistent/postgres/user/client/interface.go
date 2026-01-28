package client

import (
	"context"

	"gct/internal/domain"
	"github.com/google/uuid"
)

// UserRepoI defines the interface for user repository operations.
type RepoI interface {
	// Create creates a new user in the database.
	Create(ctx context.Context, u *domain.User) error

	// GetByPhone retrieves a user by their phone number.
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)

	// Update updates an existing user's information.
	Update(ctx context.Context, u *domain.User) error

	// Delete soft deletes a user by their ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// Get retrieves a single user based on the provided filter.
	Get(ctx context.Context, filter *domain.UserFilter) (*domain.User, error)

	// Users retrieves users based on the provided filter and pagination.
	Gets(ctx context.Context, filter *domain.UsersFilter) ([]*domain.User, int, error)
}
