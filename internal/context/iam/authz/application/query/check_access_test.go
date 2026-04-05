package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"errors"
	"testing"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock logger (implements logger.Log)
// ---------------------------------------------------------------------------

type mockLogger struct{}

func (ml *mockLogger) Debug(_ ...any)                                       {}
func (ml *mockLogger) Debugf(_ string, _ ...any)                            {}
func (ml *mockLogger) Debugw(_ string, _ ...any)                            {}
func (ml *mockLogger) Info(_ ...any)                                        {}
func (ml *mockLogger) Infof(_ string, _ ...any)                             {}
func (ml *mockLogger) Infow(_ string, _ ...any)                             {}
func (ml *mockLogger) Warn(_ ...any)                                        {}
func (ml *mockLogger) Warnf(_ string, _ ...any)                             {}
func (ml *mockLogger) Warnw(_ string, _ ...any)                             {}
func (ml *mockLogger) Error(_ ...any)                                       {}
func (ml *mockLogger) Errorf(_ string, _ ...any)                            {}
func (ml *mockLogger) Errorw(_ string, _ ...any)                            {}
func (ml *mockLogger) Fatal(_ ...any)                                       {}
func (ml *mockLogger) Fatalf(_ string, _ ...any)                            {}
func (ml *mockLogger) Fatalw(_ string, _ ...any)                            {}
func (ml *mockLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (ml *mockLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (ml *mockLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (ml *mockLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (ml *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

// ---------------------------------------------------------------------------
// Tests: CheckAccessHandler
// ---------------------------------------------------------------------------

func TestCheckAccessHandler_Allowed(t *testing.T) {
	roleID := uuid.New()
	repo := &mockAuthzReadRepository{
		checkAccessFn: func(_ context.Context, rid uuid.UUID, path, method string, _ domain.EvaluationContext) (bool, error) {
			if rid == roleID && path == "/api/v1/users" && method == "GET" {
				return true, nil
			}
			return false, nil
		},
	}

	handler := NewCheckAccessHandler(repo, logger.Noop())

	allowed, err := handler.Handle(context.Background(), CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !allowed {
		t.Error("expected access to be allowed")
	}
}

func TestCheckAccessHandler_Denied(t *testing.T) {
	roleID := uuid.New()
	repo := &mockAuthzReadRepository{
		checkAccessFn: func(_ context.Context, _ uuid.UUID, _, _ string, _ domain.EvaluationContext) (bool, error) {
			return false, nil
		},
	}

	handler := NewCheckAccessHandler(repo, logger.Noop())

	allowed, err := handler.Handle(context.Background(), CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/admin",
		Method:  "DELETE",
		EvalCtx: domain.EvaluationContext{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if allowed {
		t.Error("expected access to be denied")
	}
}

func TestCheckAccessHandler_RepoError(t *testing.T) {
	repoErr := errors.New("database connection failed")
	repo := &mockAuthzReadRepository{
		checkAccessFn: func(_ context.Context, _ uuid.UUID, _, _ string, _ domain.EvaluationContext) (bool, error) {
			return false, repoErr
		},
	}

	handler := NewCheckAccessHandler(repo, logger.Noop())

	allowed, err := handler.Handle(context.Background(), CheckAccessQuery{
		RoleID:  uuid.New(),
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{},
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
