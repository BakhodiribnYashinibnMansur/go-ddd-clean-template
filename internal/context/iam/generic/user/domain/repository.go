package domain

import (
	"context"

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

// UserRepository is the write-side persistence contract for the User aggregate.
// It extends the generic Repository with phone/email lookup methods needed for sign-in and uniqueness checks.
// FindByPhone/FindByEmail must return ErrUserNotFound when no match exists.
type UserRepository interface {
	shared.Repository[User, UserID]
	FindByPhone(ctx context.Context, phone Phone) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindDefaultRoleID(ctx context.Context) (uuid.UUID, error)

	// ActiveSessionCount returns the number of non-revoked, non-expired
	// sessions for the user at the moment of the call. Used by sign-in to
	// enforce the per-user concurrent session cap.
	ActiveSessionCount(ctx context.Context, userID UserID) (int, error)

	// RevokeOldestActiveSession revokes the user's oldest active session
	// (ordered by last_activity ASC NULLS FIRST, created_at ASC) and returns
	// its ID. Returns NilSessionID when the user has no active sessions to
	// revoke. Idempotent.
	RevokeOldestActiveSession(ctx context.Context, userID UserID) (SessionID, error)
}

// UserReadRepository provides read-only access returning lightweight UserView projections.
// It should never be used for write operations or aggregate reconstruction.
type UserReadRepository interface {
	FindByID(ctx context.Context, id UserID) (*UserView, error)
	List(ctx context.Context, filter UsersFilter) ([]*UserView, int64, error)
	FindSessionByID(ctx context.Context, id SessionID) (*shared.AuthSession, error)
	FindUserForAuth(ctx context.Context, id UserID) (*shared.AuthUser, error)
}
