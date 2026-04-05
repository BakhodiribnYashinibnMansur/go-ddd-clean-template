package command

import (
	"context"
	"testing"
	"time"

	"gct/internal/context/ops/iprule/domain"
	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockIPRuleRepo struct {
	saved   *domain.IPRule
	updated *domain.IPRule
	deleted uuid.UUID
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.IPRule, error)
}

func (m *mockIPRuleRepo) Save(_ context.Context, e *domain.IPRule) error {
	m.saved = e
	return nil
}

func (m *mockIPRuleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.IPRule, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrIPRuleNotFound
}

func (m *mockIPRuleRepo) Update(_ context.Context, e *domain.IPRule) error {
	m.updated = e
	return nil
}

func (m *mockIPRuleRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = id
	return nil
}

func (m *mockIPRuleRepo) List(_ context.Context, _ domain.IPRuleFilter) ([]*domain.IPRule, int64, error) {
	return nil, 0, nil
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

func TestCreateIPRuleHandler_Handle(t *testing.T) {
	repo := &mockIPRuleRepo{}
	eb := &mockEventBus{}
	handler := NewCreateIPRuleHandler(repo, eb, &mockLogger{})

	expires := time.Now().Add(24 * time.Hour)
	cmd := CreateIPRuleCommand{
		IPAddress: "192.168.1.100",
		Action:    "DENY",
		Reason:    "suspicious activity",
		ExpiresAt: &expires,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected ip rule to be saved")
	}
	if repo.saved.IPAddress() != "192.168.1.100" {
		t.Errorf("expected ip 192.168.1.100, got %s", repo.saved.IPAddress())
	}
	if repo.saved.Action() != "DENY" {
		t.Errorf("expected action DENY, got %s", repo.saved.Action())
	}
	if repo.saved.Reason() != "suspicious activity" {
		t.Errorf("expected reason 'suspicious activity', got %s", repo.saved.Reason())
	}
	if repo.saved.ExpiresAt() == nil {
		t.Error("expected expiresAt to be set")
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "iprule.created" {
		t.Errorf("expected iprule.created, got %s", eb.published[0].EventName())
	}
}

func TestCreateIPRuleHandler_PermanentRule(t *testing.T) {
	repo := &mockIPRuleRepo{}
	handler := NewCreateIPRuleHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateIPRuleCommand{
		IPAddress: "10.0.0.1",
		Action:    "ALLOW",
		Reason:    "trusted",
		ExpiresAt: nil,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected ip rule to be saved")
	}
	if repo.saved.ExpiresAt() != nil {
		t.Error("expected expiresAt to be nil for permanent rule")
	}
}
