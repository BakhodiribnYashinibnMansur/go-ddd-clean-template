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

	// BulkDeactivate sets active=false for all users with the given IDs.
	BulkDeactivate(ctx context.Context, ids []string) error

	// BulkDelete soft-deletes all users with the given IDs.
	BulkDelete(ctx context.Context, ids []string) error

	// Approve sets is_approved=true for the user with the given ID.
	Approve(ctx context.Context, id string) error

	// ChangeRole updates the role_id for the user with the given ID by role name.
	ChangeRole(ctx context.Context, id, role string) error
}
