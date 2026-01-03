package session

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"gct/internal/domain"
	"gct/test/e2e/common/setup"
)

// getSessionFromDB retrieves a session from the database using repository layer
func getSessionFromDB(t *testing.T, sessionID string) *domain.Session {
	t.Helper()

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		t.Logf("invalid session ID: %v", err)
		return nil
	}

	ctx := context.Background()
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

// getUserFromDB retrieves a user from the database using repository layer
func getUserFromDB(t *testing.T, userID string) *domain.User {
	t.Helper()

	uid, err := uuid.Parse(userID)
	if err != nil {
		t.Logf("invalid user ID: %v", err)
		return nil
	}

	ctx := context.Background()
	filter := &domain.UserFilter{
		ID: &uid,
	}

	user, err := setup.TestRepo.Persistent.Postgres.User.Client.Get(ctx, filter)
	if err != nil {
		t.Logf("getUserFromDB error: %v", err)
		return nil
	}

	return user
}
