package session

import (
	"testing"

	"gct/internal/domain"
	"gct/test/e2e/common/setup"
	"github.com/google/uuid"
)

// getSessionFromDB retrieves a session from the database using repository layer
func getSessionFromDB(t *testing.T, sessionID string) *domain.Session {
	t.Helper()

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		t.Logf("invalid session ID: %v", err)
		return nil
	}

	ctx := t.Context()
	filter := &domain.SessionFilter{
		ID: &sid,
	}

	session, err := setup.TestRepo.Persistent.Postgres.User.SessionRepo.Get(ctx, filter)
	if err != nil {
		t.Logf("getSessionFromDB error: %v", err)
		return nil
	}

	return session
}
