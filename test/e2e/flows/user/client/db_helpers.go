package client

import (
	"testing"
	"time"

	"gct/test/e2e/common/setup"
	"github.com/google/uuid"
)

// testUser is a lightweight struct for test assertions (replaces legacy domain.User).
type testUser struct {
	ID         uuid.UUID
	Phone      *string
	Email      *string
	Username   *string
	Active     bool
	IsApproved bool
}

// testSession is a lightweight struct for test assertions (replaces legacy domain.Session).
type testSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Revoked   bool
	ExpiresAt time.Time
}

// getUserFromDB retrieves a user from the database using direct SQL.
func getUserFromDB(t *testing.T, userID string) *testUser {
	t.Helper()

	uid, err := uuid.Parse(userID)
	if err != nil {
		t.Logf("invalid user ID: %v", err)
		return nil
	}

	ctx := t.Context()
	var u testUser
	err = setup.TestPG.Pool.QueryRow(ctx,
		`SELECT id, phone, email, username, active, is_approved FROM users WHERE id = $1 AND deleted_at = 0`, uid,
	).Scan(&u.ID, &u.Phone, &u.Email, &u.Username, &u.Active, &u.IsApproved)
	if err != nil {
		t.Logf("getUserFromDB error: %v", err)
		return nil
	}

	return &u
}

// getSessionFromDB retrieves a session from the database using direct SQL.
func getSessionFromDB(t *testing.T, sessionID string) *testSession {
	t.Helper()

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		t.Logf("invalid session ID: %v", err)
		return nil
	}

	ctx := t.Context()
	var s testSession
	err = setup.TestPG.Pool.QueryRow(ctx,
		`SELECT id, user_id, revoked, expires_at FROM session WHERE id = $1`, sid,
	).Scan(&s.ID, &s.UserID, &s.Revoked, &s.ExpiresAt)
	if err != nil {
		t.Logf("getSessionFromDB error: %v", err)
		return nil
	}

	return &s
}
