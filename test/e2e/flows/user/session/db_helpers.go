package session

import (
	"testing"
	"time"

	"gct/test/e2e/common/setup"
	"github.com/google/uuid"
)

// testSession is a lightweight struct for test assertions (replaces legacy domain.Session).
type testSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Revoked   bool
	ExpiresAt time.Time
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
