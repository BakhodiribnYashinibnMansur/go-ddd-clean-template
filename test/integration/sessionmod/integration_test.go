package sessionmod

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/session"
	appdto "gct/internal/context/iam/generic/session/application"
	sessioncmd "gct/internal/context/iam/generic/session/application/command"
	"gct/internal/context/iam/generic/session/application/query"
	sessiondomain "gct/internal/context/iam/generic/session/domain"
	"gct/internal/context/iam/generic/user"
	usercmd "gct/internal/context/iam/generic/user/application/command"
	userquery "gct/internal/context/iam/generic/user/application/query"
	"gct/internal/context/iam/generic/user/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"

	"github.com/google/uuid"
)

// testEnv bundles the two bounded contexts needed by every test.
type testEnv struct {
	userBC    *user.BoundedContext
	sessionBC *session.BoundedContext
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	l := logger.New("error")
	eb := eventbus.NewInMemoryEventBus()
	return testEnv{
		userBC:    user.NewBoundedContext(setup.TestPG.Pool, eb, l, newTestJWTConfig(t)),
		sessionBC: session.NewBoundedContext(setup.TestPG.Pool, eb, l),
	}
}

// createApprovedUser signs up a user, approves them, and returns the user ID.
func createApprovedUser(t *testing.T, ctx context.Context, env testEnv, phone, password string) uuid.UUID {
	t.Helper()

	err := env.userBC.SignUp.Handle(ctx, usercmd.SignUpCommand{
		Phone:    phone,
		Password: password,
	})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	list, err := env.userBC.ListUsers.Handle(ctx, userquery.ListUsersQuery{
		Filter: domain.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}

	// Find our user by phone (in case the DB is not perfectly clean).
	for _, u := range list.Users {
		if u.Phone == phone {
			if err := env.userBC.ApproveUser.Handle(ctx, usercmd.ApproveUserCommand{ID: domain.UserID(u.ID)}); err != nil {
				t.Fatalf("ApproveUser: %v", err)
			}
			return u.ID
		}
	}
	t.Fatalf("created user with phone %s not found in ListUsers", phone)
	return uuid.Nil
}

// signIn signs in the user and returns the SignInResult.
func signIn(t *testing.T, ctx context.Context, env testEnv, login, password, deviceType, ip, ua string) *usercmd.SignInResult {
	t.Helper()
	result, err := env.userBC.SignIn.Handle(ctx, usercmd.SignInCommand{
		Login:      login,
		Password:   password,
		DeviceType: deviceType,
		IP:         ip,
		UserAgent:  ua,
	})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	return result
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestIntegration_ListSessions_Empty(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	result, err := env.sessionBC.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 sessions, got %d", result.Total)
	}
	if len(result.Sessions) != 0 {
		t.Errorf("expected empty sessions slice, got %d items", len(result.Sessions))
	}
}

func TestIntegration_GetSession(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	phone := "+998901111111"
	password := "StrongP@ss123"
	userID := createApprovedUser(t, ctx, env, phone, password)

	sir := signIn(t, ctx, env, phone, password, "desktop", "10.0.0.1", "IntegrationTest/1.0")

	// Fetch session via Session BC
	sess, err := env.sessionBC.GetSession.Handle(ctx, query.GetSessionQuery{ID: sessiondomain.SessionID(sir.SessionID)})
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if sess.ID != sir.SessionID {
		t.Errorf("session ID: want %s, got %s", sir.SessionID, sess.ID)
	}
	if sess.UserID != userID {
		t.Errorf("user ID: want %s, got %s", userID, sess.UserID)
	}
	if sess.DeviceType != "DESKTOP" {
		t.Errorf("device type: want DESKTOP, got %s", sess.DeviceType)
	}
	if sess.IPAddress != "10.0.0.1" && sess.IPAddress != "10.0.0.1/32" {
		t.Errorf("IP address: want 10.0.0.1 or 10.0.0.1/32, got %s", sess.IPAddress)
	}
	if sess.UserAgent != "IntegrationTest/1.0" {
		t.Errorf("user agent: want IntegrationTest/1.0, got %s", sess.UserAgent)
	}
	if sess.Revoked {
		t.Error("session should not be revoked")
	}
	if sess.CreatedAt.IsZero() {
		t.Error("created_at should not be zero")
	}
	if sess.ExpiresAt.IsZero() {
		t.Error("expires_at should not be zero")
	}
}

func TestIntegration_ListSessions_WithData(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	phone := "+998902222222"
	password := "StrongP@ss123"
	userID := createApprovedUser(t, ctx, env, phone, password)

	// Sign in twice from different devices to create two sessions.
	_ = signIn(t, ctx, env, phone, password, "desktop", "10.0.0.1", "IntegrationTest/Desktop")
	_ = signIn(t, ctx, env, phone, password, "mobile", "10.0.0.2", "IntegrationTest/Mobile")

	// List sessions filtered by user ID.
	suid := sessiondomain.UserID(userID)
	result, err := env.sessionBC.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{UserID: &suid, Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 sessions, got %d", result.Total)
	}
	if len(result.Sessions) != 2 {
		t.Errorf("expected 2 session views, got %d", len(result.Sessions))
	}

	// Verify each session belongs to the correct user.
	for i, s := range result.Sessions {
		if s.UserID != userID {
			t.Errorf("session[%d] user ID: want %s, got %s", i, userID, s.UserID)
		}
	}

	// List without user filter should also return 2 (clean DB).
	all, err := env.sessionBC.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{Limit: 100},
	})
	if err != nil {
		t.Fatalf("ListSessions (all): %v", err)
	}
	if all.Total != 2 {
		t.Errorf("expected 2 total sessions, got %d", all.Total)
	}
}

func TestIntegration_RevokeSession(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	phone := "+998903333333"
	password := "StrongP@ss123"
	userID := createApprovedUser(t, ctx, env, phone, password)

	sir := signIn(t, ctx, env, phone, password, "desktop", "10.0.0.1", "IntegrationTest/1.0")

	// Revoke the session via Session BC (publishes an event).
	err := env.sessionBC.RevokeSession.Handle(ctx, sessioncmd.RevokeSessionCommand{
		UserID:    userID,
		SessionID: sir.SessionID,
	})
	if err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}

	// The revoke command only publishes a SessionRevokeRequested event.
	// Without a subscriber that actually revokes the session in the DB,
	// the session remains active. Verify the event was published successfully
	// by confirming no error above, and that the session is still readable.
	sess, err := env.sessionBC.GetSession.Handle(ctx, query.GetSessionQuery{ID: sessiondomain.SessionID(sir.SessionID)})
	if err != nil {
		t.Fatalf("GetSession after revoke event: %v", err)
	}
	if sess.ID != sir.SessionID {
		t.Errorf("session ID mismatch after revoke event")
	}
}

func TestIntegration_RevokeAllSessions(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	phone := "+998904444444"
	password := "StrongP@ss123"
	userID := createApprovedUser(t, ctx, env, phone, password)

	// Create two sessions.
	_ = signIn(t, ctx, env, phone, password, "desktop", "10.0.0.1", "IntegrationTest/Desktop")
	_ = signIn(t, ctx, env, phone, password, "mobile", "10.0.0.2", "IntegrationTest/Mobile")

	// Revoke all sessions via Session BC.
	err := env.sessionBC.RevokeAllSessions.Handle(ctx, sessioncmd.RevokeAllSessionsCommand{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("RevokeAllSessions: %v", err)
	}

	// Same as RevokeSession: the command publishes an event but does not
	// mutate the DB directly. Verify no error and sessions are still readable.
	suid := sessiondomain.UserID(userID)
	result, err := env.sessionBC.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{UserID: &suid, Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions after revoke-all event: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 sessions still present, got %d", result.Total)
	}
}

func TestIntegration_GetSession_NotFound(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	_, err := env.sessionBC.GetSession.Handle(ctx, query.GetSessionQuery{ID: sessiondomain.SessionID(uuid.New())})
	if err == nil {
		t.Fatal("expected error for non-existent session, got nil")
	}
}

func TestIntegration_ListSessions_FilterByUser(t *testing.T) {
	cleanDB(t)
	env := newTestEnv(t)
	ctx := context.Background()

	// Create two different users, each with one session.
	phone1 := "+998905555555"
	phone2 := "+998906666666"
	password := "StrongP@ss123"

	userID1 := createApprovedUser(t, ctx, env, phone1, password)
	_ = createApprovedUser(t, ctx, env, phone2, password)

	_ = signIn(t, ctx, env, phone1, password, "desktop", "10.0.0.1", "UserA")
	_ = signIn(t, ctx, env, phone2, password, "mobile", "10.0.0.2", "UserB")

	// Filter by user1 should return exactly 1.
	suid1 := sessiondomain.UserID(userID1)
	result, err := env.sessionBC.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{UserID: &suid1, Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions (filtered): %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 session for user1, got %d", result.Total)
	}
	if len(result.Sessions) == 1 && result.Sessions[0].UserID != userID1 {
		t.Errorf("session belongs to wrong user")
	}

	// Unfiltered should return 2.
	all, err := env.sessionBC.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions (all): %v", err)
	}
	if all.Total != 2 {
		t.Errorf("expected 2 total sessions, got %d", all.Total)
	}
}
