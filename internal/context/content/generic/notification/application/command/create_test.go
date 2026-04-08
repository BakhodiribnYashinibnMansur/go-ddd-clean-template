package command_test

import (
	"context"
	"testing"

	"gct/internal/context/content/generic/notification/application/command"
	notifentity "gct/internal/context/content/generic/notification/domain/entity"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockNotificationRepo struct {
	saved   *notifentity.Notification
	deleted notifentity.NotificationID
}

func (m *mockNotificationRepo) Save(_ context.Context, n *notifentity.Notification) error {
	m.saved = n
	return nil
}

func (m *mockNotificationRepo) FindByID(_ context.Context, _ notifentity.NotificationID) (*notifentity.Notification, error) {
	return nil, nil
}

func (m *mockNotificationRepo) Update(_ context.Context, _ *notifentity.Notification) error {
	return nil
}

func (m *mockNotificationRepo) Delete(_ context.Context, id notifentity.NotificationID) error {
	m.deleted = id
	return nil
}

type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error { return nil }
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error     { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                                {}
func (m *mockLogger) Debugf(_ string, _ ...any)                     {}
func (m *mockLogger) Debugw(_ string, _ ...any)                     {}
func (m *mockLogger) Info(_ ...any)                                 {}
func (m *mockLogger) Infof(_ string, _ ...any)                      {}
func (m *mockLogger) Infow(_ string, _ ...any)                      {}
func (m *mockLogger) Warn(_ ...any)                                 {}
func (m *mockLogger) Warnf(_ string, _ ...any)                      {}
func (m *mockLogger) Warnw(_ string, _ ...any)                      {}
func (m *mockLogger) Error(_ ...any)                                {}
func (m *mockLogger) Errorf(_ string, _ ...any)                     {}
func (m *mockLogger) Errorw(_ string, _ ...any)                     {}
func (m *mockLogger) Fatal(_ ...any)                                {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                     {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                     {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)  {}

// --- Tests ---

func TestCreateHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockNotificationRepo{}
	handler := command.NewCreateHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	userID := uuid.New()
	cmd := command.CreateCommand{
		UserID:  userID,
		Title:   "New Login",
		Message: "You logged in from a new device",
		Type:    "INFO",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected notification to be saved")
	}
	if repo.saved.UserID() != userID {
		t.Fatalf("expected userID %s, got %s", userID, repo.saved.UserID())
	}
	if repo.saved.Title() != "New Login" {
		t.Fatalf("expected title 'New Login', got %s", repo.saved.Title())
	}
	if repo.saved.Message() != "You logged in from a new device" {
		t.Fatalf("expected correct message, got %s", repo.saved.Message())
	}
	if repo.saved.Type() != "INFO" {
		t.Fatalf("expected type INFO, got %s", repo.saved.Type())
	}
}
