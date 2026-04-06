package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock logger (implements logger.Log)
// ---------------------------------------------------------------------------

type mockLogger struct{}

func (ml *mockLogger) Debug(_ ...any)                               {}
func (ml *mockLogger) Debugf(_ string, _ ...any)                    {}
func (ml *mockLogger) Debugw(_ string, _ ...any)                    {}
func (ml *mockLogger) Info(_ ...any)                                {}
func (ml *mockLogger) Infof(_ string, _ ...any)                     {}
func (ml *mockLogger) Infow(_ string, _ ...any)                     {}
func (ml *mockLogger) Warn(_ ...any)                                {}
func (ml *mockLogger) Warnf(_ string, _ ...any)                     {}
func (ml *mockLogger) Warnw(_ string, _ ...any)                     {}
func (ml *mockLogger) Error(_ ...any)                               {}
func (ml *mockLogger) Errorf(_ string, _ ...any)                    {}
func (ml *mockLogger) Errorw(_ string, _ ...any)                    {}
func (ml *mockLogger) Fatal(_ ...any)                               {}
func (ml *mockLogger) Fatalf(_ string, _ ...any)                    {}
func (ml *mockLogger) Fatalw(_ string, _ ...any)                    {}
func (ml *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (ml *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (ml *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (ml *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (ml *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

// ---------------------------------------------------------------------------
// Tests: CheckAccessHandler
// ---------------------------------------------------------------------------

func TestCheckAccessHandler_Allowed(t *testing.T) {
	t.Parallel()

	roleID := authzentity.NewRoleID()
	repo := &mockAuthzReadRepository{
		checkAccessFn: func(_ context.Context, rid authzentity.RoleID, path, method string, _ authzentity.EvaluationContext) (bool, error) {
			if rid == roleID && path == "/api/v1/users" && method == "GET" {
				return true, nil
			}
			return false, nil
		},
	}

	handler := NewCheckAccessHandler(repo, logger.Noop())

	allowed, err := handler.Handle(context.Background(), CheckAccessQuery{
		RoleID:  authzentity.RoleID(roleID),
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: authzentity.EvaluationContext{},
	})
	require.NoError(t, err)

	if !allowed {
		t.Error("expected access to be allowed")
	}
}

func TestCheckAccessHandler_Denied(t *testing.T) {
	t.Parallel()

	roleID := authzentity.NewRoleID()
	repo := &mockAuthzReadRepository{
		checkAccessFn: func(_ context.Context, _ authzentity.RoleID, _, _ string, _ authzentity.EvaluationContext) (bool, error) {
			return false, nil
		},
	}

	handler := NewCheckAccessHandler(repo, logger.Noop())

	allowed, err := handler.Handle(context.Background(), CheckAccessQuery{
		RoleID:  authzentity.RoleID(roleID),
		Path:    "/api/v1/admin",
		Method:  "DELETE",
		EvalCtx: authzentity.EvaluationContext{},
	})
	require.NoError(t, err)

	if allowed {
		t.Error("expected access to be denied")
	}
}

func TestCheckAccessHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("database connection failed")
	repo := &mockAuthzReadRepository{
		checkAccessFn: func(_ context.Context, _ authzentity.RoleID, _, _ string, _ authzentity.EvaluationContext) (bool, error) {
			return false, repoErr
		},
	}

	handler := NewCheckAccessHandler(repo, logger.Noop())

	allowed, err := handler.Handle(context.Background(), CheckAccessQuery{
		RoleID:  authzentity.RoleID(uuid.New()),
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: authzentity.EvaluationContext{},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if allowed {
		t.Error("expected allowed to be false on error")
	}

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
