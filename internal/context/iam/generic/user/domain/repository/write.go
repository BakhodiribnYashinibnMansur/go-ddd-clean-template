package repository

import (
	"context"

	"gct/internal/context/iam/generic/user/domain/entity"
	shared "gct/internal/kernel/domain"
)

// UserRepository is the write-side persistence contract for the User aggregate.
// It extends the generic Repository with phone/email lookup methods needed for sign-in and uniqueness checks.
// FindByPhone/FindByEmail must return ErrUserNotFound when no match exists.
type UserRepository interface {
	shared.Repository[entity.User, entity.UserID]
	FindByPhone(ctx context.Context, phone entity.Phone) (*entity.User, error)
	FindByEmail(ctx context.Context, email entity.Email) (*entity.User, error)

	// ActiveSessionCount returns the number of non-revoked, non-expired
	// sessions for the user at the moment of the call. Used by sign-in to
	// enforce the per-user concurrent session cap.
	ActiveSessionCount(ctx context.Context, userID entity.UserID) (int, error)

	// RevokeOldestActiveSession revokes the user's oldest active session
	// (ordered by last_activity ASC NULLS FIRST, created_at ASC) and returns
	// its ID. Returns NilSessionID when the user has no active sessions to
	// revoke. Idempotent.
	RevokeOldestActiveSession(ctx context.Context, userID entity.UserID) (entity.SessionID, error)

	// RevokeSessionsByIntegration revokes all active sessions for a user
	// within a specific integration. Returns the count of revoked sessions.
	RevokeSessionsByIntegration(ctx context.Context, userID entity.UserID, integrationName string) (int, error)
}
