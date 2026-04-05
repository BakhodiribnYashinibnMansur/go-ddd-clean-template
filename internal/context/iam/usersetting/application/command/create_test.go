package command_test

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"
	"gct/internal/context/iam/usersetting/application/command"
	"gct/internal/context/iam/usersetting/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockUserSettingRepo struct {
	upserted *domain.UserSetting
	deleted  uuid.UUID
	findFn   func(ctx context.Context, userID uuid.UUID, key string) (*domain.UserSetting, error)
}

func (m *mockUserSettingRepo) Upsert(_ context.Context, us *domain.UserSetting) error {
	m.upserted = us
	return nil
}

func (m *mockUserSettingRepo) FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*domain.UserSetting, error) {
	if m.findFn != nil {
		return m.findFn(ctx, userID, key)
	}
	return nil, domain.ErrUserSettingNotFound
}

func (m *mockUserSettingRepo) Delete(_ context.Context, id uuid.UUID) error {
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

func TestUpsertUserSettingHandler_Create(t *testing.T) {
	t.Parallel()

	repo := &mockUserSettingRepo{}
	handler := command.NewUpsertUserSettingHandler(repo, &mockEventBus{}, &mockLogger{})

	userID := uuid.New()
	cmd := command.UpsertUserSettingCommand{
		UserID: userID,
		Key:    "theme",
		Value:  "dark",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.upserted == nil {
		t.Fatal("expected user setting to be upserted")
	}
	if repo.upserted.UserID() != userID {
		t.Fatalf("expected userID %s, got %s", userID, repo.upserted.UserID())
	}
	if repo.upserted.Key() != "theme" {
		t.Fatalf("expected key 'theme', got %s", repo.upserted.Key())
	}
	if repo.upserted.Value() != "dark" {
		t.Fatalf("expected value 'dark', got %s", repo.upserted.Value())
	}
}

func TestUpsertUserSettingHandler_Update(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	existing := domain.NewUserSetting(userID, "theme", "light")

	repo := &mockUserSettingRepo{
		findFn: func(_ context.Context, uid uuid.UUID, key string) (*domain.UserSetting, error) {
			if uid == userID && key == "theme" {
				return existing, nil
			}
			return nil, domain.ErrUserSettingNotFound
		},
	}
	handler := command.NewUpsertUserSettingHandler(repo, &mockEventBus{}, &mockLogger{})

	cmd := command.UpsertUserSettingCommand{
		UserID: userID,
		Key:    "theme",
		Value:  "dark",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.upserted == nil {
		t.Fatal("expected user setting to be upserted")
	}
	if repo.upserted.Value() != "dark" {
		t.Fatalf("expected updated value 'dark', got %s", repo.upserted.Value())
	}
}
