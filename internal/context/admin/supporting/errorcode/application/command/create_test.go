package command

import (
	"context"
	"testing"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockErrorCodeRepo struct {
	saved   *errcodeentity.ErrorCode
	updated *errcodeentity.ErrorCode
	deleted errcodeentity.ErrorCodeID
	findFn  func(ctx context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error)
}

func (m *mockErrorCodeRepo) Save(_ context.Context, e *errcodeentity.ErrorCode) error {
	m.saved = e
	return nil
}

func (m *mockErrorCodeRepo) FindByID(ctx context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, errcodeentity.ErrErrorCodeNotFound
}

func (m *mockErrorCodeRepo) Update(_ context.Context, e *errcodeentity.ErrorCode) error {
	m.updated = e
	return nil
}

func (m *mockErrorCodeRepo) Delete(_ context.Context, id errcodeentity.ErrorCodeID) error {
	m.deleted = id
	return nil
}

type mockEventBus struct {
	published []shared.DomainEvent
}

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                               {}
func (m *mockLogger) Debugf(_ string, _ ...any)                    {}
func (m *mockLogger) Debugw(_ string, _ ...any)                    {}
func (m *mockLogger) Info(_ ...any)                                {}
func (m *mockLogger) Infof(_ string, _ ...any)                     {}
func (m *mockLogger) Infow(_ string, _ ...any)                     {}
func (m *mockLogger) Warn(_ ...any)                                {}
func (m *mockLogger) Warnf(_ string, _ ...any)                     {}
func (m *mockLogger) Warnw(_ string, _ ...any)                     {}
func (m *mockLogger) Error(_ ...any)                               {}
func (m *mockLogger) Errorf(_ string, _ ...any)                    {}
func (m *mockLogger) Errorw(_ string, _ ...any)                    {}
func (m *mockLogger) Fatal(_ ...any)                               {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                    {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

// --- Tests ---

func TestCreateErrorCodeHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockErrorCodeRepo{}
	eb := &mockEventBus{}
	handler := NewCreateErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	cmd := CreateErrorCodeCommand{
		Code:       "AUTH_001",
		Message:    "unauthorized access",
		HTTPStatus: 401,
		Category:   "auth",
		Severity:   "high",
		Retryable:  false,
		RetryAfter: 0,
		Suggestion: "check your token",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected error code to be saved")
	}
	if repo.saved.Code() != "AUTH_001" {
		t.Errorf("expected code AUTH_001, got %s", repo.saved.Code())
	}
	if repo.saved.Message() != "unauthorized access" {
		t.Errorf("expected message 'unauthorized access', got %s", repo.saved.Message())
	}
	if repo.saved.HTTPStatus() != 401 {
		t.Errorf("expected httpStatus 401, got %d", repo.saved.HTTPStatus())
	}
	if repo.saved.Category() != "auth" {
		t.Errorf("expected category auth, got %s", repo.saved.Category())
	}
	if repo.saved.Severity() != "high" {
		t.Errorf("expected severity high, got %s", repo.saved.Severity())
	}
	if repo.saved.Retryable() {
		t.Error("expected retryable false")
	}
	if repo.saved.Suggestion() != "check your token" {
		t.Errorf("expected suggestion 'check your token', got %s", repo.saved.Suggestion())
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "errorcode.created" {
		t.Errorf("expected errorcode.created, got %s", eb.published[0].EventName())
	}
}

func TestCreateErrorCodeHandler_MinimalFields(t *testing.T) {
	t.Parallel()

	repo := &mockErrorCodeRepo{}
	handler := NewCreateErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), CreateErrorCodeCommand{
		Code:       "ERR_BASIC",
		Message:    "basic",
		HTTPStatus: 500,
		Category:   "general",
		Severity:   "low",
	})
	require.NoError(t, err)
	if repo.saved == nil {
		t.Fatal("expected error code to be saved")
	}
	if repo.saved.RetryAfter() != 0 {
		t.Errorf("expected retryAfter 0, got %d", repo.saved.RetryAfter())
	}
}
