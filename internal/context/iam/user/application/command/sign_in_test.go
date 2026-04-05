package command

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
)

func testJWTConfig(t *testing.T) JWTConfig {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return JWTConfig{
		PrivateKey: key,
		Issuer:     "test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}
}

func TestSignInHandler_Handle(t *testing.T) {
	// Create a user with known credentials.
	phone, err := domain.NewPhone("+998901234567")
	if err != nil {
		t.Fatalf("failed to create phone: %v", err)
	}
	password, err := domain.NewPasswordFromRaw("StrongP@ss123")
	if err != nil {
		t.Fatalf("failed to create password: %v", err)
	}

	user := domain.NewUser(phone, password)
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

	handler := NewSignInHandler(phoneRepo, eventBus, log, testJWTConfig(t))

	cmd := SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
	phone, _ := domain.NewPhone("+998901234567")
	password, _ := domain.NewPasswordFromRaw("StrongP@ss123")

	user := domain.NewUser(phone, password)
	user.Approve()

	phoneRepo := &signInMockRepo{user: user}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignInHandler(phoneRepo, eventBus, log, testJWTConfig(t))

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
	phone, _ := domain.NewPhone("+998901234567")
	password, _ := domain.NewPasswordFromRaw("StrongP@ss123")

	user := domain.NewUser(phone, password)
	user.Approve()
	user.Deactivate()

	phoneRepo := &signInMockRepo{user: user}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignInHandler(phoneRepo, eventBus, log, testJWTConfig(t))

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
	user *domain.User
}

func (m *signInMockRepo) FindByPhone(_ context.Context, phone domain.Phone) (*domain.User, error) {
	if m.user != nil && m.user.Phone().Value() == phone.Value() {
		return m.user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *signInMockRepo) FindByEmail(_ context.Context, email domain.Email) (*domain.User, error) {
	if m.user != nil && m.user.Email() != nil && m.user.Email().Value() == email.Value() {
		return m.user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *signInMockRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *signInMockRepo) Update(ctx context.Context, entity *domain.User) error {
	m.updatedUser = entity
	return nil
}
