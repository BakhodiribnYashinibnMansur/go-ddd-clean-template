package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

// UserRepoI defines the interface for user repository operations.
type UserRepoI interface {
	// Create creates a new user in the database.
	Create(ctx context.Context, u domain.User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id int64) (domain.User, error)

	// GetByPhone retrieves a user by their phone number.
	GetByPhone(ctx context.Context, phone string) (domain.User, error)

	// Update updates an existing user's information.
	Update(ctx context.Context, u domain.User) error

	// Delete soft deletes a user by their ID.
	Delete(ctx context.Context, id int64) error
}
