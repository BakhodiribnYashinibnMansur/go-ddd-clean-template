package domain

import (
	"context"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// UserView is a read-model DTO used by the read side. It will be fully defined
// in the application layer (application/dto.go); declared here as a placeholder
// so that the read-repository interface compiles.
type UserView struct {
	ID         uuid.UUID      `json:"id"`
	Phone      string         `json:"phone"`
	Email      *string        `json:"email,omitempty"`
	Username   *string        `json:"username,omitempty"`
	RoleID     *uuid.UUID     `json:"role_id,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Active     bool           `json:"active"`
	IsApproved bool           `json:"is_approved"`
}

// UsersFilter carries filtering and pagination parameters for listing users.
type UsersFilter struct {
	Phone      *string
	Email      *string
	Active     *bool
	IsApproved *bool
	Pagination *shared.Pagination
}

// UserRepository is the write-side repository for the User aggregate.
type UserRepository interface {
	shared.Repository[User]
	FindByPhone(ctx context.Context, phone Phone) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
}

// UserReadRepository is the read-side repository returning projected views.
type UserReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*UserView, error)
	List(ctx context.Context, filter UsersFilter) ([]*UserView, int64, error)
}
