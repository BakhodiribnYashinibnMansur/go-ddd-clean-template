package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/authz/domain"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock RoleRepository ---

type mockRoleRepository struct {
	savedRole   *domain.Role
	updatedRole *domain.Role
	findByIDFn  func(ctx context.Context, id uuid.UUID) (*domain.Role, error)
}

func (m *mockRoleRepository) Save(ctx context.Context, role *domain.Role) error {
	m.savedRole = role
	return nil
}

func (m *mockRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	m.updatedRole = role
	return nil
}

func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockRoleRepository) List(ctx context.Context, pagination shared.Pagination) ([]*domain.Role, int64, error) {
	return nil, 0, nil
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

func (m *mockLogger) Debug(args ...any)                                            {}
func (m *mockLogger) Debugf(template string, args ...any)                          {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Info(args ...any)                                             {}
func (m *mockLogger) Infof(template string, args ...any)                           {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                       {}
func (m *mockLogger) Warn(args ...any)                                             {}
func (m *mockLogger) Warnf(template string, args ...any)                           {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                       {}
func (m *mockLogger) Error(args ...any)                                            {}
func (m *mockLogger) Errorf(template string, args ...any)                          {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Fatal(args ...any)                                            {}
func (m *mockLogger) Fatalf(template string, args ...any)                          {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Debugc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Infoc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Warnc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Errorc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Fatalc(ctx context.Context, msg string, keysAndValues ...any) {}

// --- Tests ---

func TestCreateRoleHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockRoleRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateRoleHandler(repo, eventBus, log)

	desc := "Administrator role"
	cmd := CreateRoleCommand{
		Name:        "admin",
		Description: &desc,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedRole == nil {
		t.Fatal("expected role to be saved, but it was nil")
	}

	if repo.savedRole.Name() != "admin" {
		t.Errorf("expected name 'admin', got '%s'", repo.savedRole.Name())
	}

	if repo.savedRole.Description() == nil || *repo.savedRole.Description() != "Administrator role" {
		t.Error("expected description to be 'Administrator role'")
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected at least one event to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "authz.role_created" {
		t.Errorf("expected event authz.role_created, got %s", eventBus.publishedEvents[0].EventName())
	}
}

func TestCreateRoleHandler_NoDescription(t *testing.T) {
	t.Parallel()

	repo := &mockRoleRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateRoleHandler(repo, eventBus, log)

	cmd := CreateRoleCommand{
		Name: "viewer",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedRole == nil {
		t.Fatal("expected role to be saved")
	}

	if repo.savedRole.Name() != "viewer" {
		t.Errorf("expected name 'viewer', got '%s'", repo.savedRole.Name())
	}

	if repo.savedRole.Description() != nil {
		t.Error("expected nil description")
	}
}
