package command_test

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"
	"gct/internal/context/ops/generic/systemerror/application/command"
	"gct/internal/context/ops/generic/systemerror/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockSystemErrorRepo struct {
	saved *domain.SystemError
}

func (m *mockSystemErrorRepo) Save(_ context.Context, se *domain.SystemError) error {
	m.saved = se
	return nil
}

func (m *mockSystemErrorRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.SystemError, error) {
	return nil, nil
}

func (m *mockSystemErrorRepo) Update(_ context.Context, _ *domain.SystemError) error {
	return nil
}

func (m *mockSystemErrorRepo) List(_ context.Context, _ domain.SystemErrorFilter) ([]*domain.SystemError, int64, error) {
	return nil, 0, nil
}

type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error { return nil }
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error {
	return nil
}

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

func TestCreateSystemErrorHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockSystemErrorRepo{}
	handler := command.NewCreateSystemErrorHandler(repo, &mockEventBus{}, &mockLogger{})

	cmd := command.CreateSystemErrorCommand{
		Code:     "DB_CONNECTION_FAILED",
		Message:  "could not connect to database",
		Severity: "FATAL",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected system error to be saved")
	}
	if repo.saved.Code() != "DB_CONNECTION_FAILED" {
		t.Fatalf("expected code DB_CONNECTION_FAILED, got %s", repo.saved.Code())
	}
	if repo.saved.Message() != "could not connect to database" {
		t.Fatalf("expected correct message, got %s", repo.saved.Message())
	}
	if repo.saved.Severity() != "FATAL" {
		t.Fatalf("expected severity FATAL, got %s", repo.saved.Severity())
	}
}
