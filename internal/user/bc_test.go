package user

import (
	"context"
	"testing"

	"gct/internal/shared/application"
	"gct/internal/shared/domain"
	"gct/internal/user/application/command"
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
	jwtCfg := command.JWTConfig{} // zero-value config is fine for wiring test
	bc := NewBoundedContext(nil, &mockEventBus{}, &mockLogger{}, jwtCfg)
	if bc == nil {
		t.Fatal("expected non-nil BoundedContext")
	}
	// Commands
	if bc.CreateUser == nil {
		t.Error("CreateUser handler not wired")
	}
	if bc.UpdateUser == nil {
		t.Error("UpdateUser handler not wired")
	}
	if bc.DeleteUser == nil {
		t.Error("DeleteUser handler not wired")
	}
	if bc.SignIn == nil {
		t.Error("SignIn handler not wired")
	}
	if bc.SignUp == nil {
		t.Error("SignUp handler not wired")
	}
	if bc.SignOut == nil {
		t.Error("SignOut handler not wired")
	}
	if bc.ApproveUser == nil {
		t.Error("ApproveUser handler not wired")
	}
	if bc.ChangeRole == nil {
		t.Error("ChangeRole handler not wired")
	}
	if bc.BulkAction == nil {
		t.Error("BulkAction handler not wired")
	}
	if bc.RevokeAll == nil {
		t.Error("RevokeAll handler not wired")
	}
	// Queries
	if bc.GetUser == nil {
		t.Error("GetUser handler not wired")
	}
	if bc.ListUsers == nil {
		t.Error("ListUsers handler not wired")
	}
	if bc.FindSession == nil {
		t.Error("FindSession handler not wired")
	}
	if bc.FindUserForAuth == nil {
		t.Error("FindUserForAuth handler not wired")
	}
}
