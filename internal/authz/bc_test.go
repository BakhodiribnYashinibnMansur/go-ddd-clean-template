package authz

import (
	"context"
	"testing"

	"gct/internal/shared/application"
	"gct/internal/shared/domain"
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
	bc := NewBoundedContext(nil, &mockEventBus{}, &mockLogger{})
	if bc == nil {
		t.Fatal("expected non-nil BoundedContext")
	}
	// Commands - Roles
	if bc.CreateRole == nil {
		t.Error("CreateRole handler not wired")
	}
	if bc.UpdateRole == nil {
		t.Error("UpdateRole handler not wired")
	}
	if bc.DeleteRole == nil {
		t.Error("DeleteRole handler not wired")
	}
	// Commands - Permissions
	if bc.CreatePermission == nil {
		t.Error("CreatePermission handler not wired")
	}
	if bc.DeletePermission == nil {
		t.Error("DeletePermission handler not wired")
	}
	// Commands - Policies
	if bc.CreatePolicy == nil {
		t.Error("CreatePolicy handler not wired")
	}
	if bc.UpdatePolicy == nil {
		t.Error("UpdatePolicy handler not wired")
	}
	if bc.DeletePolicy == nil {
		t.Error("DeletePolicy handler not wired")
	}
	if bc.TogglePolicy == nil {
		t.Error("TogglePolicy handler not wired")
	}
	// Commands - Scopes
	if bc.CreateScope == nil {
		t.Error("CreateScope handler not wired")
	}
	if bc.DeleteScope == nil {
		t.Error("DeleteScope handler not wired")
	}
	// Commands - Assignments
	if bc.AssignPermission == nil {
		t.Error("AssignPermission handler not wired")
	}
	if bc.AssignScope == nil {
		t.Error("AssignScope handler not wired")
	}
	// Queries
	if bc.GetRole == nil {
		t.Error("GetRole handler not wired")
	}
	if bc.ListRoles == nil {
		t.Error("ListRoles handler not wired")
	}
	if bc.ListPermissions == nil {
		t.Error("ListPermissions handler not wired")
	}
	if bc.ListPolicies == nil {
		t.Error("ListPolicies handler not wired")
	}
	if bc.ListScopes == nil {
		t.Error("ListScopes handler not wired")
	}
	if bc.CheckAccess == nil {
		t.Error("CheckAccess handler not wired")
	}
}
