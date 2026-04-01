package errorx

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNewServiceErrorLogger(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "user-service")

	if logger == nil {
		t.Fatal("expected non-nil ServiceErrorLogger")
	}
	if logger.serviceName != "user-service" {
		t.Errorf("expected serviceName 'user-service', got %q", logger.serviceName)
	}
	if logger.errorLogger == nil {
		t.Error("expected non-nil inner errorLogger")
	}
}

func TestServiceErrorLogger_LogError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "auth-service")

	err := logger.LogError(context.Background(), "AUTH_FAILED", "authentication failed", nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.createCallCount != 1 {
		t.Errorf("expected 1 repo.Create call, got %d", repo.createCallCount)
	}
}

func TestServiceErrorLogger_LogError_WithContextValues(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "test-service")

	reqID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", reqID)
	ctx = context.WithValue(ctx, "user_id", userID)

	err := logger.LogError(ctx, "TEST_CODE", "test message", nil, map[string]any{"key": "val"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.createCallCount != 1 {
		t.Errorf("expected 1 repo.Create call, got %d", repo.createCallCount)
	}
	if repo.lastInput.RequestID == nil {
		t.Error("expected RequestID to be extracted from context")
	}
	if repo.lastInput.UserID == nil {
		t.Error("expected UserID to be extracted from context")
	}
}

func TestServiceErrorLogger_LogError_WithoutContextValues(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "test-service")

	err := logger.LogError(context.Background(), "TEST_CODE", "test message", nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.lastInput.RequestID != nil {
		t.Error("expected nil RequestID when not in context")
	}
	if repo.lastInput.UserID != nil {
		t.Error("expected nil UserID when not in context")
	}
}

func TestServiceErrorLogger_LogDatabaseError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "user-service")

	err := logger.LogDatabaseError(context.Background(), nil, "INSERT", "users")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.createCallCount != 1 {
		t.Errorf("expected 1 repo.Create call, got %d", repo.createCallCount)
	}
}

func TestServiceErrorLogger_LogBusinessError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "order-service")

	details := map[string]any{
		"order_id": "ord-123",
		"reason":   "insufficient balance",
	}

	err := logger.LogBusinessError(context.Background(), "ORDER_FAILED", "order processing failed", nil, details)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestServiceErrorLogger_LogError_RepoFailure(t *testing.T) {
	repoErr := ErrTestRepoFailure
	repo := &mockRepository{err: repoErr}
	log := &mockLogger{}
	logger := NewServiceErrorLogger(repo, log, "test-service")

	err := logger.LogError(context.Background(), "TEST", "test", nil, nil)
	if err == nil {
		t.Fatal("expected error when repo fails")
	}
	if err != repoErr {
		t.Errorf("expected repo error, got %v", err)
	}
}
