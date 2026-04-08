package file

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
	"gct/internal/kernel/domain"
	"gct/internal/kernel/outbox"
)

type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ ...domain.DomainEvent) error { return nil }
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error     { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                                  {}
func (m *mockLogger) Debugf(_ string, _ ...any)                       {}
func (m *mockLogger) Debugw(_ string, _ ...any)                       {}
func (m *mockLogger) Info(_ ...any)                                   {}
func (m *mockLogger) Infof(_ string, _ ...any)                        {}
func (m *mockLogger) Infow(_ string, _ ...any)                        {}
func (m *mockLogger) Warn(_ ...any)                                   {}
func (m *mockLogger) Warnf(_ string, _ ...any)                        {}
func (m *mockLogger) Warnw(_ string, _ ...any)                        {}
func (m *mockLogger) Error(_ ...any)                                  {}
func (m *mockLogger) Errorf(_ string, _ ...any)                       {}
func (m *mockLogger) Errorw(_ string, _ ...any)                       {}
func (m *mockLogger) Fatal(_ ...any)                                  {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                       {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                       {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)    {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)     {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)     {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)    {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)    {}

func TestNewBoundedContext(t *testing.T) {
	bc := NewBoundedContext(nil, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})
	if bc == nil {
		t.Fatal("expected non-nil BoundedContext")
	}
	if bc.CreateFile == nil {
		t.Error("CreateFile handler not wired")
	}
	if bc.GetFile == nil {
		t.Error("GetFile handler not wired")
	}
	if bc.ListFiles == nil {
		t.Error("ListFiles handler not wired")
	}
}
