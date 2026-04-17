package notification

import (
	"context"
	"testing"

	"gct/internal/kernel/outbox"
)

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
	l := &mockLogger{}
	bc := NewBoundedContext(nil, outbox.NewEventCommitter(nil, nil, nil, l), l)
	if bc == nil {
		t.Fatal("expected non-nil BoundedContext")
	}
	if bc.CreateNotification == nil {
		t.Error("CreateNotification handler not wired")
	}
	if bc.DeleteNotification == nil {
		t.Error("DeleteNotification handler not wired")
	}
	if bc.GetNotification == nil {
		t.Error("GetNotification handler not wired")
	}
	if bc.ListNotifications == nil {
		t.Error("ListNotifications handler not wired")
	}
}
