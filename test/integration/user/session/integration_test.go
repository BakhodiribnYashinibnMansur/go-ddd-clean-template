package session

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"gct/internal/context/iam/generic/session"
	sessionapp "gct/internal/context/iam/generic/session/application"
	sessionquery "gct/internal/context/iam/generic/session/application/query"
	sessiondomain "gct/internal/context/iam/generic/session/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	userquery "gct/internal/context/iam/generic/user/application/query"
	"gct/internal/context/iam/generic/user/domain"
	"gct/test/integration/common/setup"
)

func newTestJWTConfig(t *testing.T) command.JWTConfig {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	return command.JWTConfig{
		PrivateKey: key,
		Issuer:     "gct-test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}
}

func TestIntegration_ListAndGetSessions(t *testing.T) {
	cleanDB(t)
	l := logger.New("error")
	eb := eventbus.NewInMemoryEventBus()

	userBC := user.NewBoundedContext(setup.TestPG.Pool, eb, l, newTestJWTConfig(t))
	sessionBC := session.NewBoundedContext(setup.TestPG.Pool, eb, l)
	ctx := context.Background()

	// Create and approve user
	err := userBC.SignUp.Handle(ctx, command.SignUpCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	list, _ := userBC.ListUsers.Handle(ctx, userquery.ListUsersQuery{
		Filter: domain.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID
	_ = userBC.ApproveUser.Handle(ctx, command.ApproveUserCommand{ID: domain.UserID(userID)})

	// Sign in to create a session
	signInResult, err := userBC.SignIn.Handle(ctx, command.SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "10.0.0.1",
		UserAgent:  "IntegrationTest/1.0",
	})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	// List sessions
	filterUserID := sessiondomain.UserID(userID)
	sessions, err := sessionBC.ListSessions.Handle(ctx, sessionquery.ListSessionsQuery{
		Filter: sessionapp.SessionsFilter{UserID: &filterUserID, Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if sessions.Total != 1 {
		t.Fatalf("expected 1 session, got %d", sessions.Total)
	}

	// Get session by ID
	sess, err := sessionBC.GetSession.Handle(ctx, sessionquery.GetSessionQuery{ID: sessiondomain.SessionID(signInResult.SessionID)})
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess.UserID != userID.UUID() {
		t.Errorf("user ID mismatch: %s vs %s", sess.UserID, userID)
	}
}
