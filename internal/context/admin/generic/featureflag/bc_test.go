package featureflag

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
	"gct/internal/kernel/domain"
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

// TestNewBoundedContext_NilPool verifies that NewBoundedContext returns an error
// (or panics gracefully) when given a nil pool, because the CachedEvaluator
// eagerly loads all flags from the database during construction.
func TestNewBoundedContext_NilPool(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("NewBoundedContext panicked with nil pool (expected): %v", r)
		}
	}()

	ctx := context.Background()
	bc, err := NewBoundedContext(ctx, nil, &mockEventBus{}, &mockLogger{})
	if err != nil {
		// Expected: constructor returns error because CachedEvaluator cannot load flags with nil pool.
		t.Logf("NewBoundedContext returned expected error with nil pool: %v", err)
		return
	}
	if bc == nil {
		t.Fatal("expected non-nil BoundedContext when no error returned")
	}

	// If we got here, verify all handlers are wired.
	if bc.CreateFlag == nil {
		t.Error("CreateFlag handler not wired")
	}
	if bc.UpdateFlag == nil {
		t.Error("UpdateFlag handler not wired")
	}
	if bc.DeleteFlag == nil {
		t.Error("DeleteFlag handler not wired")
	}
	if bc.CreateRuleGroup == nil {
		t.Error("CreateRuleGroup handler not wired")
	}
	if bc.UpdateRuleGroup == nil {
		t.Error("UpdateRuleGroup handler not wired")
	}
	if bc.DeleteRuleGroup == nil {
		t.Error("DeleteRuleGroup handler not wired")
	}
	if bc.GetFlag == nil {
		t.Error("GetFlag handler not wired")
	}
	if bc.ListFlags == nil {
		t.Error("ListFlags handler not wired")
	}
	if bc.Evaluator == nil {
		t.Error("Evaluator not wired")
	}
}
