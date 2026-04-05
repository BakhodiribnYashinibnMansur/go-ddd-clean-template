package command

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"
	"gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock Repository ---

type mockUserRepository struct {
	savedUser   *domain.User
	updatedUser *domain.User
	findByIDFn  func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *mockUserRepository) Save(ctx context.Context, entity *domain.User) error {
	m.savedUser = entity
	return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Update(ctx context.Context, entity *domain.User) error {
	m.updatedUser = entity
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, filter shared.Pagination) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) FindByPhone(ctx context.Context, phone domain.Phone) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

// --- Mock EventBus ---

type mockEventBus struct {
	publishedEvents []shared.DomainEvent
}

func (m *mockEventBus) Publish(ctx context.Context, events ...shared.DomainEvent) error {
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *mockEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	return nil
}

// --- Mock Logger ---

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                          {}
func (m *mockLogger) Debugf(template string, args ...any)                        {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Info(args ...any)                                           {}
func (m *mockLogger) Infof(template string, args ...any)                         {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Warn(args ...any)                                           {}
func (m *mockLogger) Warnf(template string, args ...any)                         {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Error(args ...any)                                          {}
func (m *mockLogger) Errorf(template string, args ...any)                        {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Fatal(args ...any)                                          {}
func (m *mockLogger) Fatalf(template string, args ...any)                        {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Debugc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Infoc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Warnc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Errorc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Fatalc(ctx context.Context, msg string, keysAndValues ...any) {}

// --- Tests ---

func TestCreateUserHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateUserHandler(repo, eventBus, log)

	email := "test@example.com"
	username := "testuser"

	cmd := CreateUserCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
		Email:    &email,
		Username: &username,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedUser == nil {
		t.Fatal("expected user to be saved, but it was nil")
	}

	if repo.savedUser.Phone().Value() != "+998901234567" {
		t.Errorf("expected phone +998901234567, got %s", repo.savedUser.Phone().Value())
	}

	if repo.savedUser.Email() == nil || repo.savedUser.Email().Value() != "test@example.com" {
		t.Error("expected email to be set to test@example.com")
	}

	if repo.savedUser.Username() == nil || *repo.savedUser.Username() != "testuser" {
		t.Error("expected username to be set to testuser")
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected at least one event to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "user.created" {
		t.Errorf("expected event user.created, got %s", eventBus.publishedEvents[0].EventName())
	}
}

func TestCreateUserHandler_InvalidPhone(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateUserHandler(repo, eventBus, log)

	cmd := CreateUserCommand{
		Phone:    "invalid",
		Password: "StrongP@ss123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid phone, got nil")
	}

	if repo.savedUser != nil {
		t.Error("expected no user to be saved for invalid phone")
	}
}

func TestCreateUserHandler_WeakPassword(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateUserHandler(repo, eventBus, log)

	cmd := CreateUserCommand{
		Phone:    "+998901234567",
		Password: "short",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for weak password, got nil")
	}
}
