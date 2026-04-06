package entity

import (
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// UserView is a read-model DTO for API responses. Intentionally omits sensitive fields
// (password hash, sessions) that should never leave the domain layer.
type UserView struct {
	ID         UserID            `json:"id"`
	Phone      string            `json:"phone"`
	Email      *string           `json:"email,omitempty"`
	Username   *string           `json:"username,omitempty"`
	RoleID     *uuid.UUID        `json:"role_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Active     bool              `json:"active"`
	IsApproved bool              `json:"is_approved"`
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
