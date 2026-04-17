package command

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	shared "gct/internal/kernel/domain"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

// fakeResolver implements IntegrationResolver by returning a fixed
// JWTResolved constructed from a freshly-generated RSA key.
type fakeResolver struct {
	resolved *JWTResolved
	err      error
}

func (f *fakeResolver) Resolve(_ context.Context, _ string) (*JWTResolved, error) {
	return f.resolved, f.err
}

func testJWTConfig(t *testing.T) JWTConfig {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	hasher, err := jwtpkg.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	return JWTConfig{
		Issuer:        "test",
		RefreshHasher: hasher,
		Resolver: &fakeResolver{resolved: &JWTResolved{
			Name:       "gct-client",
			PrivateKey: key,
			KeyID:      "test-kid",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		}},
	}
}

func TestSignInHandler_Handle(t *testing.T) {
	t.Parallel()

	// Create a user with known credentials.
	phone, err := userentity.NewPhone("+998901234567")
	require.NoError(t, err)
	password, err := userentity.NewPasswordFromRaw("StrongP@ss123")
	require.NoError(t, err)

	user, _ := userentity.NewUser(phone, password)
	user.Approve()
	// Clear events from construction/approval so we only check sign-in events.
	user.ClearEvents()

	repo := &mockUserRepository{
		findByIDFn: nil,
	}
	// Override FindByPhone to return our test user.
	repo.findByIDFn = nil

	// We need a custom mock that returns the user by phone.
	phoneRepo := &signInMockRepo{user: user}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignInHandler(phoneRepo, outbox.NewEventCommitter(nil, nil, eventBus, log), log, testJWTConfig(t))

	cmd := SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected sign-in result, got nil")
	}

	if result.UserID != user.ID() {
		t.Errorf("expected user ID %s, got %s", user.ID(), result.UserID)
	}

	if result.SessionID == [16]byte{} {
		t.Error("expected a non-zero session ID")
	}

	// Verify the user was saved (updated) with the new session.
	if phoneRepo.updatedUser == nil {
		t.Fatal("expected user to be updated after sign-in")
	}

	if len(phoneRepo.updatedUser.Sessions()) == 0 {
		t.Error("expected at least one session after sign-in")
	}

	// Verify events were published.
	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected at least one event to be published")
	}
}

func TestSignInHandler_WrongPassword(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	password, _ := userentity.NewPasswordFromRaw("StrongP@ss123")

	user, _ := userentity.NewUser(phone, password)
	user.Approve()

	phoneRepo := &signInMockRepo{user: user}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignInHandler(phoneRepo, outbox.NewEventCommitter(nil, nil, eventBus, log), log, testJWTConfig(t))

	cmd := SignInCommand{
		Login:      "+998901234567",
		Password:   "WrongPassword",
		DeviceType: "desktop",
		IP:         "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestSignInHandler_InactiveUser(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	password, _ := userentity.NewPasswordFromRaw("StrongP@ss123")

	user, _ := userentity.NewUser(phone, password)
	user.Approve()
	user.Deactivate()

	phoneRepo := &signInMockRepo{user: user}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignInHandler(phoneRepo, outbox.NewEventCommitter(nil, nil, eventBus, log), log, testJWTConfig(t))

	cmd := SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for inactive user, got nil")
	}
}

// signInMockRepo is a specialized mock that returns a user by phone.
type signInMockRepo struct {
	mockUserRepository
	user *userentity.User
}

func (m *signInMockRepo) FindByPhone(_ context.Context, phone userentity.Phone) (*userentity.User, error) {
	if m.user != nil && m.user.Phone().Value() == phone.Value() {
		return m.user, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *signInMockRepo) FindByEmail(_ context.Context, email userentity.Email) (*userentity.User, error) {
	if m.user != nil && m.user.Email() != nil && m.user.Email().Value() == email.Value() {
		return m.user, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *signInMockRepo) Update(ctx context.Context, q shared.Querier, entity *userentity.User) error {
	m.updatedUser = entity
	return nil
}

// ---------------------------------------------------------------------------
// Max-concurrent-sessions cap tests
// ---------------------------------------------------------------------------

func TestSignInHandler_EvictsOldestWhenAtCap(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	password, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, password)
	user.Approve()
	user.ClearEvents()

	repo := &signInMockRepo{user: user}
	repo.activeCount = 3 // already at the cap
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	maxFn := func(_ context.Context) int { return 3 }
	handler := NewSignInHandler(repo, outbox.NewEventCommitter(nil, nil, eventBus, log), log, testJWTConfig(t), maxFn)

	cmd := SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.NotNil(t, result)

	if repo.revokedOldest != 1 {
		t.Errorf("expected exactly 1 eviction, got %d", repo.revokedOldest)
	}
}

func TestSignInHandler_NoEvictionBelowCap(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	password, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, password)
	user.Approve()
	user.ClearEvents()

	repo := &signInMockRepo{user: user}
	repo.activeCount = 2 // under the cap — no eviction needed
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	maxFn := func(_ context.Context) int { return 3 }
	handler := NewSignInHandler(repo, outbox.NewEventCommitter(nil, nil, eventBus, log), log, testJWTConfig(t), maxFn)

	cmd := SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.revokedOldest != 0 {
		t.Errorf("expected no eviction, got %d", repo.revokedOldest)
	}
}
