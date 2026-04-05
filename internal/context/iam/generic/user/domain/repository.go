package domain

import (
	"context"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// UserView is a read-model DTO for API responses. Intentionally omits sensitive fields
// (password hash, sessions) that should never leave the domain layer.
type UserView struct {
	ID         uuid.UUID      `json:"id"`
	Phone      string         `json:"phone"`
	Email      *string        `json:"email,omitempty"`
	Username   *string        `json:"username,omitempty"`
	RoleID     *uuid.UUID     `json:"role_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Active     bool           `json:"active"`
	IsApproved bool           `json:"is_approved"`
}

// UsersFilter carries optional filtering and pagination parameters for listing users.
// Nil pointer fields are ignored by the repository (no filtering on that dimension).
type UsersFilter struct {
	Phone      *string
	Email      *string
	Active     *bool
	IsApproved *bool
	Pagination *shared.Pagination
}

// UserRepository is the write-side persistence contract for the User aggregate.
// It extends the generic Repository with phone/email lookup methods needed for sign-in and uniqueness checks.
// FindByPhone/FindByEmail must return ErrUserNotFound when no match exists.
type UserRepository interface {
	shared.Repository[User]
	FindByPhone(ctx context.Context, phone Phone) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindDefaultRoleID(ctx context.Context) (uuid.UUID, error)
}

// UserReadRepository provides read-only access returning lightweight UserView projections.
// It should never be used for write operations or aggregate reconstruction.
type UserReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*UserView, error)
	List(ctx context.Context, filter UsersFilter) ([]*UserView, int64, error)
	FindSessionByID(ctx context.Context, id uuid.UUID) (*shared.AuthSession, error)
	FindUserForAuth(ctx context.Context, id uuid.UUID) (*shared.AuthUser, error)
}
